// Form submission handler. Ported from the former Go backend
// (backend/handler.go): same order of checks, status codes and messages.

import type { Attachment, FormKind, Submission } from './email';

// Gmail rejects messages over 25 MB and base64 adds a third, so anything
// larger could never have been delivered; it also keeps us inside the
// Worker's 128 MB memory limit.
export const MAX_BODY_BYTES = 18 * 1024 * 1024;

export interface Deps {
	allow(ip: string): Promise<boolean>;
	verifyTurnstile(token: string, ip: string): Promise<boolean>;
	send(kind: FormKind, sub: Submission): Promise<void>;
	log(msg: string): void;
}

const json = (status: number, body: unknown) =>
	new Response(JSON.stringify(body) + '\n', {
		status,
		headers: { 'Content-Type': 'application/json' }
	});
const error = (status: number, msg: string) => json(status, { error: msg });

/**
 * seccostudio.com/* and http://www.seccostudio.com/* -> https://www.seccostudio.com/*
 * (both were Traefik middlewares on the old cluster).
 */
export function canonicalRedirect(reqUrl: string, canonicalHost: string | undefined): string | null {
	if (!canonicalHost) return null;
	const url = new URL(reqUrl);
	const apex = canonicalHost.replace(/^www\./, '');
	const isApex = url.hostname === apex;
	const isPlainHttp = url.hostname === canonicalHost && url.protocol === 'http:';
	if (!isApex && !isPlainHttp) return null;
	url.hostname = canonicalHost;
	url.protocol = 'https:';
	url.port = '';
	return url.toString();
}

export function clientIP(req: Request): string {
	return (
		req.headers.get('CF-Connecting-IP') ??
		req.headers.get('X-Forwarded-For')?.split(',')[0].trim() ??
		'unknown'
	);
}

export async function handleSubmit(req: Request, kind: FormKind, deps: Deps): Promise<Response> {
	const ip = clientIP(req);

	if (!(await deps.allow(ip))) {
		deps.log(`rate limited: ${ip}`);
		return error(429, 'Zbyt wiele prób. Spróbuj ponownie za godzinę.');
	}

	let form: FormData;
	try {
		const type = req.headers.get('Content-Type') ?? '';
		if (!type.toLowerCase().startsWith('multipart/form-data')) throw new Error('not multipart');
		if (Number(req.headers.get('Content-Length') ?? 0) > MAX_BODY_BYTES) throw new Error('too large');
		const body = await req.arrayBuffer();
		if (body.byteLength > MAX_BODY_BYTES) throw new Error('too large');
		form = await new Response(body, {
			headers: { 'Content-Type': type }
		}).formData();
	} catch {
		return error(400, 'Nieprawidłowe dane formularza.');
	}

	const answers: Record<string, string> = {};
	const attachments: Attachment[] = [];
	const multi = new Map<string, string[]>();
	for (const [key, value] of form.entries()) {
		if (typeof value === 'string') {
			multi.set(key, [...(multi.get(key) ?? []), value]);
		} else {
			attachments.push({
				name: value.name,
				contentType: value.type,
				data: new Uint8Array(await value.arrayBuffer())
			});
		}
	}
	const first = (key: string) => multi.get(key)?.[0] ?? '';

	let verified = false;
	try {
		verified = await deps.verifyTurnstile(first('cf-turnstile-response'), ip);
	} catch (err) {
		deps.log(`turnstile error: ${err}`);
	}
	if (!verified) {
		deps.log(`turnstile verification failed: ip=${ip}`);
		return error(403, 'Weryfikacja nie powiodła się. Odśwież stronę i spróbuj ponownie.');
	}

	const firstName = first('first_name').trim();
	const email = first('email').trim();
	if (!firstName || !email) return error(400, 'Imię i email są wymagane.');
	if (!email.includes('@') || !email.includes('.') || /[\r\n]/.test(email))
		return error(400, 'Nieprawidłowy adres email.');

	for (const [key, values] of multi) {
		if (key !== 'cf-turnstile-response') answers[key] = values.join(', ');
	}

	const sub: Submission = {
		firstName,
		lastName: first('last_name').trim(),
		email,
		answers,
		attachments
	};

	try {
		await deps.send(kind, sub);
	} catch (err) {
		deps.log(`email send failed: ${err}`);
		return error(500, 'Nie udało się wysłać wiadomości. Spróbuj ponownie później.');
	}

	deps.log(`submission received: kind=${kind} attachments=${attachments.length}`);
	return json(200, { success: true });
}

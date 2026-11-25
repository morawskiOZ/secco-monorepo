// Cloudflare Worker for www.seccostudio.com.
// Static files are served by Workers Static Assets; this script only handles
// /api/*, the apex -> www redirect and the SPA fallback.

import { connect } from 'cloudflare:sockets';
import { buildMessage, type FormKind } from './email';
import { canonicalRedirect, handleSubmit, type Deps } from './handler';
import { RateLimiter } from './ratelimit';
import { sendMail, type Conn } from './smtp';

export { RateLimiter };

interface Env {
	ASSETS: Fetcher;
	RATE_LIMITER: DurableObjectNamespace<RateLimiter>;
	CANONICAL_HOST?: string;
	SMTP_HOST: string;
	SMTP_PORT: string;
	SMTP_USER: string;
	SMTP_PASS: string;
	RECIPIENT_EMAIL: string;
	TURNSTILE_SECRET_KEY?: string;
}

const TURNSTILE_VERIFY_URL = 'https://challenges.cloudflare.com/turnstile/v0/siteverify';

function deps(env: Env): Deps {
	return {
		async allow(ip) {
			try {
				return await env.RATE_LIMITER.get(env.RATE_LIMITER.idFromName(ip)).allow();
			} catch (err) {
				console.error(`rate limiter unavailable: ${err}`);
				return true; // Turnstile still protects the form.
			}
		},
		async verifyTurnstile(token, ip) {
			if (!env.TURNSTILE_SECRET_KEY) return true;
			const res = await fetch(TURNSTILE_VERIFY_URL, {
				method: 'POST',
				body: new URLSearchParams({
					secret: env.TURNSTILE_SECRET_KEY,
					response: token,
					remoteip: ip
				}),
				signal: AbortSignal.timeout(10_000)
			});
			return ((await res.json()) as { success?: boolean }).success === true;
		},
		async send(kind: FormKind, sub) {
			const wrap = (s: Socket): Conn => ({
				readable: s.readable,
				writable: s.writable,
				startTls: () => wrap(s.startTls()),
				close: () => s.close()
			});
			await sendMail({
				connect: () =>
					wrap(
						connect(
							{ hostname: env.SMTP_HOST, port: Number(env.SMTP_PORT) },
							{ secureTransport: 'starttls', allowHalfOpen: false }
						)
					),
				clientName: 'seccostudio.com',
				user: env.SMTP_USER,
				pass: env.SMTP_PASS,
				from: env.SMTP_USER,
				to: env.RECIPIENT_EMAIL,
				message: buildMessage(
					{ smtpUser: env.SMTP_USER, recipientEmail: env.RECIPIENT_EMAIL },
					kind,
					sub
				)
			});
		},
		log: (msg) => console.log(msg)
	};
}

export default {
	async fetch(req: Request, env: Env): Promise<Response> {
		const url = new URL(req.url);

		const redirect = canonicalRedirect(req.url, env.CANONICAL_HOST);
		if (redirect) return Response.redirect(redirect, 308);

		if (url.pathname === '/api/health' && req.method === 'GET') {
			return new Response('{"status":"ok"}', { headers: { 'Content-Type': 'application/json' } });
		}
		if (url.pathname === '/api/submit' && req.method === 'POST') {
			return handleSubmit(req, 'valuation', deps(env));
		}
		if (url.pathname === '/api/submit-preferences' && req.method === 'POST') {
			return handleSubmit(req, 'preferences', deps(env));
		}
		if (url.pathname.startsWith('/api/')) {
			return new Response('Method Not Allowed\n', { status: 405 });
		}

		// Static file, or the SvelteKit SPA shell for client-side routes
		// (the Go server answered every unknown path with 200.html).
		const res = await env.ASSETS.fetch(req);
		if (res.status !== 404) return res;
		const shell = await env.ASSETS.fetch(new Request(new URL('/200', url), { headers: req.headers }));
		return new Response(shell.body, { status: 200, headers: shell.headers });
	}
} satisfies ExportedHandler<Env>;

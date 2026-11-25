import { describe, expect, it, vi } from 'vitest';
import { canonicalRedirect, handleSubmit, MAX_BODY_BYTES, type Deps } from '../src/handler';
import { slidingWindowAllow } from '../src/window';

function deps(over: Partial<Deps> = {}) {
	return {
		allow: vi.fn(async () => true),
		verifyTurnstile: vi.fn(async () => true),
		send: vi.fn(async () => {}),
		log: vi.fn(),
		...over
	} satisfies Deps;
}

function post(fields: Record<string, string | string[] | File>, headers: Record<string, string> = {}) {
	const fd = new FormData();
	for (const [k, v] of Object.entries(fields)) {
		if (Array.isArray(v)) v.forEach((x) => fd.append(k, x));
		else fd.append(k, v);
	}
	return new Request('https://www.example.com/api/submit', {
		method: 'POST',
		body: fd,
		headers: { 'CF-Connecting-IP': '203.0.113.7', ...headers }
	});
}

const valid = { first_name: ' Jan ', email: 'jan@example.com', 'cf-turnstile-response': 'tok' };

describe('handleSubmit', () => {
	it('sends the email and returns success', async () => {
		const d = deps();
		const res = await handleSubmit(
			post({ ...valid, last_name: 'K', q11: ['a', 'b'], q7: new File([new Uint8Array([1, 2, 3])], 'rzut.pdf', { type: 'application/pdf' }) }),
			'valuation',
			d
		);
		expect(res.status).toBe(200);
		expect(res.headers.get('Content-Type')).toBe('application/json');
		expect(await res.json()).toEqual({ success: true });
		expect(d.allow).toHaveBeenCalledWith('203.0.113.7');
		expect(d.verifyTurnstile).toHaveBeenCalledWith('tok', '203.0.113.7');
		const [kind, sub] = (d.send as any).mock.calls[0] as [string, any];
		expect(kind).toBe('valuation');
		expect(sub.firstName).toBe('Jan');
		expect(sub.lastName).toBe('K');
		expect(sub.answers).toEqual({ first_name: ' Jan ', email: 'jan@example.com', last_name: 'K', q11: 'a, b' });
		expect(sub.attachments).toHaveLength(1);
		expect(sub.attachments[0].name).toBe('rzut.pdf');
		expect(sub.attachments[0].contentType).toBe('application/pdf');
		expect([...sub.attachments[0].data]).toEqual([1, 2, 3]);
	});

	it('returns 429 before anything else when rate limited', async () => {
		const d = deps({ allow: vi.fn(async () => false) });
		const res = await handleSubmit(post(valid), 'valuation', d);
		expect(res.status).toBe(429);
		expect(await res.json()).toEqual({ error: 'Zbyt wiele prób. Spróbuj ponownie za godzinę.' });
		expect(d.verifyTurnstile).not.toHaveBeenCalled();
	});

	it('returns 400 for a non-multipart body', async () => {
		for (const body of ['nope', new URLSearchParams(valid)]) {
			const req = new Request('https://x/api/submit', { method: 'POST', body });
			const res = await handleSubmit(req, 'valuation', deps());
			expect(res.status).toBe(400);
			expect(await res.json()).toEqual({ error: 'Nieprawidłowe dane formularza.' });
		}
	});

	it('returns 400 for an oversized body without reading it', async () => {
		const res = await handleSubmit(
			post(valid, { 'Content-Length': String(MAX_BODY_BYTES + 1) }),
			'valuation',
			deps()
		);
		expect(res.status).toBe(400);
	});

	it('returns 403 when Turnstile fails or errors', async () => {
		for (const verifyTurnstile of [vi.fn(async () => false), vi.fn(async () => Promise.reject(new Error('net')))]) {
			const d = deps({ verifyTurnstile });
			const res = await handleSubmit(post(valid), 'valuation', d);
			expect(res.status).toBe(403);
			expect(await res.json()).toEqual({
				error: 'Weryfikacja nie powiodła się. Odśwież stronę i spróbuj ponownie.'
			});
			expect(d.send).not.toHaveBeenCalled();
		}
	});

	it('validates required fields and email shape', async () => {
		let res = await handleSubmit(post({ ...valid, first_name: '  ' }), 'valuation', deps());
		expect(res.status).toBe(400);
		expect(await res.json()).toEqual({ error: 'Imię i email są wymagane.' });
		res = await handleSubmit(post({ ...valid, email: 'nope' }), 'valuation', deps());
		expect(res.status).toBe(400);
		expect(await res.json()).toEqual({ error: 'Nieprawidłowy adres email.' });
	});

	it('returns 500 when sending fails', async () => {
		const d = deps({ send: vi.fn(async () => Promise.reject(new Error('smtp down'))) });
		const res = await handleSubmit(post(valid), 'preferences', d);
		expect(res.status).toBe(500);
		expect(await res.json()).toEqual({
			error: 'Nie udało się wysłać wiadomości. Spróbuj ponownie później.'
		});
	});
});

describe('slidingWindowAllow', () => {
	it('allows 5 per hour and recovers as old hits expire', () => {
		const hour = 3_600_000;
		let hits: number[] = [];
		for (let i = 0; i < 5; i++) {
			const r = slidingWindowAllow(hits, 1000 + i);
			expect(r.allowed).toBe(true);
			hits = r.hits;
		}
		const blocked = slidingWindowAllow(hits, 2000);
		expect(blocked.allowed).toBe(false);
		expect(blocked.hits).toHaveLength(5);
		const later = slidingWindowAllow(blocked.hits, 1000 + hour + 1);
		expect(later.allowed).toBe(true);
		expect(later.hits).toHaveLength(4);
	});
});

describe('canonicalRedirect', () => {
	it('sends the apex to www over https, keeping path and query', () => {
		expect(canonicalRedirect('http://seccostudio.com/x?y=1', 'www.seccostudio.com')).toBe(
			'https://www.seccostudio.com/x?y=1'
		);
	});
	it('upgrades plain http on www to https', () => {
		expect(canonicalRedirect('http://www.seccostudio.com/form?a=b', 'www.seccostudio.com')).toBe(
			'https://www.seccostudio.com/form?a=b'
		);
	});
	it('leaves www, workers.dev and unconfigured deployments alone', () => {
		expect(canonicalRedirect('https://www.seccostudio.com/x', 'www.seccostudio.com')).toBeNull();
		expect(canonicalRedirect('https://secco-studio.x.workers.dev/', 'www.seccostudio.com')).toBeNull();
		expect(canonicalRedirect('https://seccostudio.com/', undefined)).toBeNull();
		expect(canonicalRedirect('http://localhost:8787/', 'www.seccostudio.com')).toBeNull();
	});
});

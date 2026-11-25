import { describe, expect, it } from 'vitest';
import { dotStuff, sendMail, SmtpError, type Conn } from '../src/smtp';

const enc = new TextEncoder();
const dec = new TextDecoder();

/** Scripted fake SMTP server: replies once per client line (or once for DATA payload). */
function fakeServer(replies: Record<string, string>) {
	const log: string[] = [];
	let tls = false;
	const make = (): Conn => {
		let push!: (s: string) => void;
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				push = (s) => c.enqueue(enc.encode(s));
				if (!tls) push('220 fake ESMTP\r\n');
			}
		});
		let inData = false;
		let pending = '';
		const writable = new WritableStream<Uint8Array>({
			write(chunk) {
				pending += dec.decode(chunk);
				if (inData) {
					if (!pending.endsWith('\r\n.\r\n')) return;
					log.push('<payload>' + pending.slice(0, -5));
					pending = '';
					inData = false;
					push(replies['.'] ?? '250 queued\r\n');
					return;
				}
				let nl;
				while ((nl = pending.indexOf('\r\n')) >= 0) {
					const line = pending.slice(0, nl);
					pending = pending.slice(nl + 2);
					log.push((tls ? 'tls:' : 'plain:') + line);
					const verb = line.split(/[ :]/)[0];
					if (verb === 'DATA') inData = true;
					push(replies[verb] ?? '250 ok\r\n');
				}
			}
		});
		return {
			readable,
			writable,
			startTls() {
				tls = true;
				return make();
			},
			async close() {}
		};
	};
	return { connect: make, log };
}

const ok = {
	EHLO: '250-fake\r\n250-STARTTLS\r\n250 AUTH PLAIN\r\n',
	STARTTLS: '220 go ahead\r\n',
	AUTH: '235 accepted\r\n',
	DATA: '354 send\r\n',
	QUIT: '221 bye\r\n'
};

const base = {
	clientName: 'test',
	user: 'u@example.com',
	pass: 'secret',
	from: 'u@example.com',
	to: 'inbox@example.com'
};

describe('sendMail', () => {
	it('runs STARTTLS, authenticates only after TLS, and sends dot-stuffed data', async () => {
		const srv = fakeServer(ok);
		await sendMail({
			...base,
			connect: srv.connect,
			message: [enc.encode('Subject: x\r\n\r\n.hidden\r\nline\r\n'), enc.encode('.second\r\n')]
		});
		expect(srv.log).toEqual([
			'plain:EHLO test',
			'plain:STARTTLS',
			'tls:EHLO test',
			'tls:AUTH PLAIN ' + btoa('\0u@example.com\0secret'),
			'tls:MAIL FROM:<u@example.com>',
			'tls:RCPT TO:<inbox@example.com>',
			'tls:DATA',
			'<payload>Subject: x\r\n\r\n..hidden\r\nline\r\n..second',
			'tls:QUIT'
		]);
	});

	it('fails with the server reply when auth is rejected', async () => {
		const srv = fakeServer({ ...ok, AUTH: '535 5.7.8 bad credentials\r\n' });
		await expect(
			sendMail({ ...base, connect: srv.connect, message: [enc.encode('x\r\n')] })
		).rejects.toThrow(/AUTH: 535/);
		expect(srv.log.some((l) => l.includes('MAIL FROM'))).toBe(false);
	});

	it('fails when the message is rejected after DATA', async () => {
		const srv = fakeServer({ ...ok, '.': '552 too big\r\n' });
		await expect(
			sendMail({ ...base, connect: srv.connect, message: [enc.encode('x\r\n')] })
		).rejects.toBeInstanceOf(SmtpError);
	});

	it('rejects addresses that could inject SMTP commands', async () => {
		const srv = fakeServer(ok);
		await expect(
			sendMail({ ...base, to: 'a@b.c>\r\nRCPT TO:<x@y.z', connect: srv.connect, message: [] })
		).rejects.toThrow(/invalid address/);
	});
});

describe('dotStuff', () => {
	it('returns the same chunk when nothing needs escaping', () => {
		const c = enc.encode('QUJD\r\nREVG\r\n');
		expect(dotStuff(c)).toBe(c);
		const d = enc.encode('a.b\r\n');
		expect(dotStuff(d)).toBe(d);
	});
});

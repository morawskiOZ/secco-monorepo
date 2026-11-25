import { describe, expect, it } from 'vitest';
import { base64Lines, buildMessage, encodeSubject } from '../src/email';

const cfg = { smtpUser: 'studio@example.com', recipientEmail: 'inbox@example.com' };
const dec = (chunks: Uint8Array[]) => Buffer.concat(chunks).toString('utf8');

describe('encodeSubject', () => {
	it('leaves plain ASCII untouched', () => {
		expect(encodeSubject('Nowa wycena - Jan Kowalski')).toBe('Nowa wycena - Jan Kowalski');
	});

	it('Q-encodes non-ASCII as RFC 2047 words of at most 75 chars', () => {
		const out = encodeSubject('Nowa wycena — Żaneta Łąćka-Śmigły z bardzo długim nazwiskiem rodowym');
		for (const word of out.split('\r\n ')) {
			expect(word).toMatch(/^=\?utf-8\?q\?[^ ?]*\?=$/);
			expect(word.length).toBeLessThanOrEqual(75);
		}
		const decoded = out
			.split('\r\n ')
			.map((w) => w.slice(10, -2).replace(/_/g, ' '))
			.join('')
			.replace(/((?:=[0-9A-F]{2})+)/g, (m) =>
				new TextDecoder().decode(
					Uint8Array.from(m.split('=').slice(1).map((h) => parseInt(h, 16)))
				)
			);
		expect(decoded).toBe('Nowa wycena — Żaneta Łąćka-Śmigły z bardzo długim nazwiskiem rodowym');
	});
});

describe('base64Lines', () => {
	it('emits CRLF-terminated lines of at most 76 chars that round-trip', () => {
		for (const size of [0, 1, 57, 58, 57 * 4096, 57 * 4096 + 1, 600_000]) {
			const data = new Uint8Array(size).map((_, i) => (i * 31) % 251);
			const text = dec(base64Lines(data));
			if (size === 0) {
				expect(text).toBe('');
				continue;
			}
			expect(text.endsWith('\r\n')).toBe(true);
			const lines = text.slice(0, -2).split('\r\n');
			expect(lines.slice(0, -1).every((l: string) => l.length === 76)).toBe(true);
			expect(lines.at(-1)!.length).toBeLessThanOrEqual(76);
			expect(Buffer.from(lines.join(''), 'base64').equals(Buffer.from(data))).toBe(true);
		}
	});
});

describe('buildMessage', () => {
	it('builds the valuation email like the Go backend did', () => {
		const text = dec(
			buildMessage(cfg, 'valuation', {
				firstName: 'Jan',
				lastName: 'Kowalski',
				email: 'jan@example.com',
				answers: { q2: '55 m2', q1: 'Instagram', first_name: 'Jan', email: 'jan@example.com', junk: 'x', q8: '' },
				attachments: []
			})
		);
		const [head, body] = text.split('\r\n\r\n', 2);
		expect(head.split('\r\n')).toEqual([
			'From: studio@example.com',
			'To: inbox@example.com',
			'Reply-To: jan@example.com',
			'Subject: =?utf-8?q?Nowa_wycena_=E2=80=94_Jan_Kowalski?=',
			'MIME-Version: 1.0',
			'Content-Type: text/plain; charset=utf-8'
		]);
		expect(text.slice(head.length + 4)).toBe(
			[
				'Nowe zgłoszenie z formularza wyceny',
				'=====================================',
				'',
				'Jak trafiłeś na tą ankietę?:',
				'Instagram',
				'',
				'Metraż powierzchni projektowej:',
				'55 m2',
				'',
				'Imię:',
				'Jan',
				'',
				'Email:',
				'jan@example.com',
				'',
				''
			].join('\r\n')
		);
		expect(body).toContain('Nowe zgłoszenie');
	});

	it('uses the preferences template and labels', () => {
		const text = dec(
			buildMessage(cfg, 'preferences', {
				firstName: 'Ala',
				lastName: 'Nowak',
				email: 'ala@example.com',
				answers: { p3: 'tak', first_name: 'Ala' },
				attachments: []
			})
		);
		expect(text).toContain('Subject: =?utf-8?q?Ankieta_preferencji_=E2=80=94_Ala_Nowak?=');
		expect(text).toContain(
			'Nowe zgłoszenie z ankiety preferencji klienta\r\n==============================================\r\n\r\nImię:\r\nAla\r\n\r\nAlergicy:\r\ntak\r\n\r\n'
		);
	});

	it('attaches files as wrapped base64 in multipart/mixed and lists them in the body', () => {
		const data = new Uint8Array(100).map((_, i) => i);
		const text = dec(
			buildMessage(cfg, 'valuation', {
				firstName: 'Jan',
				lastName: '',
				email: 'jan@example.com',
				answers: {},
				attachments: [{ name: 'rzut "A".pdf', contentType: '', data }]
			})
		);
		const boundary = /Content-Type: multipart\/mixed; boundary="?([^"\r\n]+)"?/.exec(text)![1];
		const parts = text.split(`--${boundary}`);
		expect(parts).toHaveLength(4);
		expect(parts[3]).toBe('--\r\n');
		expect(parts[1]).toContain('Przesłane pliki:\r\n  - rzut "A".pdf\r\n');
		expect(parts[2]).toContain('Content-Type: application/octet-stream\r\n');
		expect(parts[2]).toContain('Content-Transfer-Encoding: base64\r\n');
		expect(parts[2]).toContain('Content-Disposition: attachment; filename="rzut \\"A\\".pdf"\r\n');
		const b64 = parts[2].split('\r\n\r\n')[1].trim();
		for (const line of b64.split('\r\n')) expect(line.length).toBeLessThanOrEqual(76);
		expect([...Buffer.from(b64.replace(/\r\n/g, ''), 'base64')]).toEqual([...data]);
	});

	it('strips CR/LF from values that end up in headers', () => {
		const text = dec(
			buildMessage(cfg, 'valuation', {
				firstName: 'Jan\r\nBcc: evil@example.com',
				lastName: '',
				email: 'jan@example.com\r\nBcc: evil@example.com',
				answers: {},
				attachments: []
			})
		);
		const head = text.split('\r\n\r\n')[0];
		expect(head).not.toMatch(/^Bcc:/m);
		expect(head.split('\r\n')).toHaveLength(6);
	});
});

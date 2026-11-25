// Builds the notification email. Ported 1:1 from the former Go backend
// (backend/email.go): same subjects, labels, ordering and plain-text layout.

export interface Attachment {
	name: string;
	contentType: string;
	data: Uint8Array;
}

export interface Submission {
	firstName: string;
	lastName: string;
	email: string;
	answers: Record<string, string>;
	attachments: Attachment[];
}

export interface MailConfig {
	smtpUser: string;
	recipientEmail: string;
	/** Prepended to the subject, e.g. "[TEST] " on staging. */
	subjectPrefix?: string;
}

export type FormKind = 'valuation' | 'preferences';

type Labels = ReadonlyArray<readonly [key: string, label: string]>;

const valuationLabels: Labels = [
	['q1', 'Jak trafiłeś na tą ankietę?'],
	['q1_other', 'Jak trafiłeś — inne'],
	['q2', 'Metraż powierzchni projektowej'],
	['q3', 'Powierzchnia — całość czy poszczególne pomieszczenia'],
	['q4', 'Lokalizacja inwestycji'],
	['q5', 'Stan deweloperski czy rynek wtórny'],
	['q6', 'Planowane oddanie inwestycji'],
	['q7_info', 'Rzut inwestycji'],
	['q8', 'Piętro'],
	['q9', 'Skosy'],
	['q10', 'Styl wnętrzarski'],
	['q10_other', 'Styl — inne'],
	['q11', 'Zakres projektu'],
	['q12', 'Nadzór autorski'],
	['q13', 'Kiedy rozpocząć prace'],
	['first_name', 'Imię'],
	['last_name', 'Nazwisko'],
	['phone', 'Telefon'],
	['email', 'Email']
];

const preferenceLabels: Labels = [
	['first_name', 'Imię'],
	['last_name', 'Nazwisko'],
	['phone', 'Telefon'],
	['email', 'Email'],
	['p2', 'Ile osób użytkuje mieszkanie'],
	['p3', 'Alergicy'],
	['p4', 'Praca w domu / biuro / warsztat'],
	['p5', 'Meble i dekoracje z poprzedniego mieszkania'],
	['p6', 'Zmiana położenia przyłączy hydraulicznych / elektrycznych'],
	['p7', 'Ulubione kolory'],
	['p8', 'Kolory i materiały do unikania'],
	['p9', 'Wykończenie podłogi w pokojach'],
	['p10', 'Szerokość zmywarki'],
	['p11', 'Lodówka w kuchni'],
	['p12', 'Dodatkowy sprzęt w kuchni'],
	['p13', 'Wyspa w kuchni'],
	['p14', 'Kuchenka'],
	['p15', 'Wymiary łóżka w sypialni głównej'],
	['p16', 'Sanitariaty w łazience'],
	['p17', 'Miejsce na ładowanie sprzętów elektrycznych'],
	['p18', 'Rodzaj ogrzewania'],
	['p19', 'Ingerencja w konstrukcję'],
	['p20', 'Inne sprzęty codziennego użytku'],
	['p21', 'Preferowany styl'],
	['p21_other', 'Preferowany styl — inne'],
	['p22', 'Zwierzęta'],
	['p23', 'Sport i sprzęt sportowy'],
	['p24', 'Instrumenty muzyczne'],
	['p25', 'Goście / sofa rozkładana'],
	['p26', 'Odkurzacz centralny'],
	['p27', 'System inteligentnego domu'],
	['p29', 'Sugestie i specjalne potrzeby']
];

const templates = {
	valuation: {
		subject: 'Nowa wycena',
		intro: 'Nowe zgłoszenie z formularza wyceny\n=====================================\n\n',
		labels: valuationLabels
	},
	preferences: {
		subject: 'Ankieta preferencji',
		intro:
			'Nowe zgłoszenie z ankiety preferencji klienta\n==============================================\n\n',
		labels: preferenceLabels
	}
} as const;

const enc = new TextEncoder();
const oneLine = (s: string) => s.replace(/[\r\n]+/g, ' ');

/** RFC 2047 Q-encoding, matching Go's mime.QEncoding.Encode("utf-8", s). */
export function encodeSubject(s: string): string {
	if (/^[\x20-\x7e]*$/.test(s) && !s.includes('=?')) return s;
	const maxLen = 75 - '=?utf-8?q?'.length - '?='.length;
	const words: string[] = [];
	let cur = '';
	for (const ch of s) {
		let piece = '';
		for (const b of enc.encode(ch)) {
			if (b === 0x20) piece += '_';
			else if (b > 0x20 && b < 0x7f && b !== 0x3d && b !== 0x3f && b !== 0x5f)
				piece += String.fromCharCode(b);
			else piece += '=' + b.toString(16).toUpperCase().padStart(2, '0');
		}
		if (cur.length + piece.length > maxLen) {
			words.push(cur);
			cur = '';
		}
		cur += piece;
	}
	words.push(cur);
	return words.map((w) => `=?utf-8?q?${w}?=`).join('\r\n ');
}

function toBase64(data: Uint8Array): string {
	const native = (data as unknown as { toBase64?: () => string }).toBase64;
	if (typeof native === 'function') return native.call(data);
	let bin = '';
	for (let i = 0; i < data.length; i += 0x8000)
		bin += String.fromCharCode(...data.subarray(i, i + 0x8000));
	return btoa(bin);
}

const LINE = 76;
const BYTES_PER_LINE = 57;
const LINES_PER_CHUNK = 4096;

/**
 * Base64 with CRLF after every 76 characters (RFC 2045), produced in ~300 KB
 * chunks. Works on bytes and native calls only: a regex over the whole string
 * costs ~2 ms of CPU per MB, which matters under the Workers CPU limit.
 */
export function base64Lines(data: Uint8Array): Uint8Array[] {
	const out: Uint8Array[] = [];
	const scratch = new Uint8Array(LINE * LINES_PER_CHUNK);
	const step = BYTES_PER_LINE * LINES_PER_CHUNK;
	for (let pos = 0; pos < data.length; pos += step) {
		const { written } = enc.encodeInto(toBase64(data.subarray(pos, pos + step)), scratch);
		const lines = Math.ceil(written / LINE);
		const chunk = new Uint8Array(written + lines * 2);
		let o = 0;
		for (let i = 0; i < written; i += LINE) {
			const row = scratch.subarray(i, Math.min(i + LINE, written));
			chunk.set(row, o);
			o += row.length;
			chunk[o++] = 13;
			chunk[o++] = 10;
		}
		out.push(chunk);
	}
	return out;
}

/** Returns the full RFC 5322 message as CRLF-terminated chunks. */
export function buildMessage(cfg: MailConfig, kind: FormKind, sub: Submission): Uint8Array[] {
	const t = templates[kind];
	const subject = oneLine(`${cfg.subjectPrefix ?? ''}${t.subject} — ${sub.firstName} ${sub.lastName}`);

	let body: string = t.intro;
	for (const [key, label] of t.labels) {
		const v = sub.answers[key];
		if (v) body += `${label}:\n${v}\n\n`;
	}
	if (sub.attachments.length > 0) {
		body += 'Przesłane pliki:\n';
		for (const att of sub.attachments) body += `  - ${oneLine(att.name)}\n`;
		body += '\n';
	}
	body = body.replace(/\r\n|\r|\n/g, '\r\n');

	const headers = [
		`From: ${oneLine(cfg.smtpUser)}`,
		`To: ${oneLine(cfg.recipientEmail)}`,
		`Reply-To: ${oneLine(sub.email)}`,
		`Subject: ${encodeSubject(subject)}`,
		'MIME-Version: 1.0'
	];

	if (sub.attachments.length === 0) {
		headers.push('Content-Type: text/plain; charset=utf-8');
		return [enc.encode(headers.join('\r\n') + '\r\n\r\n' + body)];
	}

	const boundary = '----SeccoFormBoundary' + crypto.randomUUID().replace(/-/g, '');
	headers.push(`Content-Type: multipart/mixed; boundary="${boundary}"`);

	const chunks: Uint8Array[] = [
		enc.encode(
			headers.join('\r\n') +
				`\r\n\r\n--${boundary}\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n` +
				body +
				'\r\n'
		)
	];
	for (const att of sub.attachments) {
		const type = /^[\w.+-]+\/[\w.+-]+$/.test(att.contentType)
			? att.contentType
			: 'application/octet-stream';
		const name = oneLine(att.name).replace(/[\\"]/g, '\\$&');
		chunks.push(
			enc.encode(
				`--${boundary}\r\nContent-Type: ${type}\r\nContent-Transfer-Encoding: base64\r\n` +
					`Content-Disposition: attachment; filename="${name}"\r\n\r\n`
			),
			...base64Lines(att.data)
		);
	}
	chunks.push(enc.encode(`--${boundary}--\r\n`));
	return chunks;
}

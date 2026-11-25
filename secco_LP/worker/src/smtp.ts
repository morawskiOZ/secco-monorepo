// Minimal SMTP submission client (STARTTLS + AUTH PLAIN), enough to replace
// Go's smtp.SendMail. Transport is injected so it can be tested without a
// network; in the Worker it is a `cloudflare:sockets` TCP socket.

export interface Conn {
	readable: ReadableStream<Uint8Array>;
	writable: WritableStream<Uint8Array>;
	startTls(): Conn;
	close(): Promise<void>;
}

export interface SendOptions {
	connect: () => Conn;
	clientName: string;
	user: string;
	pass: string;
	from: string;
	to: string;
	/** Complete message; every chunk must end with CRLF. */
	message: Uint8Array[];
	timeoutMs?: number;
}

export class SmtpError extends Error {}

const enc = new TextEncoder();
const dec = new TextDecoder();

class Session {
	private reader: ReadableStreamDefaultReader<Uint8Array>;
	private writer: WritableStreamDefaultWriter<Uint8Array>;
	private buf = '';

	constructor(private conn: Conn) {
		this.reader = conn.readable.getReader();
		this.writer = conn.writable.getWriter();
	}

	upgrade(): Session {
		this.reader.releaseLock();
		this.writer.releaseLock();
		return new Session(this.conn.startTls());
	}

	write(data: Uint8Array) {
		return this.writer.write(data);
	}

	/** Reads one (possibly multi-line) reply and checks its status code. */
	async expect(code: number, step: string): Promise<string> {
		const lines: string[] = [];
		for (;;) {
			let nl: number;
			while ((nl = this.buf.indexOf('\n')) < 0) {
				const { value, done } = await this.reader.read();
				if (done) throw new SmtpError(`${step}: connection closed`);
				this.buf += dec.decode(value, { stream: true });
			}
			const line = this.buf.slice(0, nl).replace(/\r$/, '');
			this.buf = this.buf.slice(nl + 1);
			lines.push(line);
			if (line[3] !== '-') break;
		}
		const reply = lines.join(' | ');
		if (!reply.startsWith(String(code))) throw new SmtpError(`${step}: ${reply.slice(0, 200)}`);
		return reply;
	}

	async cmd(line: string, code: number, step: string) {
		await this.write(enc.encode(line + '\r\n'));
		return this.expect(code, step);
	}

	async close() {
		try {
			await this.conn.close();
		} catch {
			/* already closed */
		}
	}
}

/** Escapes lines starting with "." (RFC 5321 4.5.2). Chunks start at line starts. */
export function dotStuff(chunk: Uint8Array): Uint8Array {
	if (!chunk.includes(0x2e)) return chunk;
	const text = dec.decode(chunk);
	const stuffed = text.replace(/(^|\r\n)\./g, '$1..');
	return stuffed.length === text.length ? chunk : enc.encode(stuffed);
}

const addr = (a: string) => {
	if (/[\s<>]/.test(a)) throw new SmtpError('invalid address');
	return `<${a}>`;
};

export async function sendMail(o: SendOptions): Promise<void> {
	let s = new Session(o.connect());
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<never>((_, reject) => {
		timer = setTimeout(() => reject(new SmtpError('timeout')), o.timeoutMs ?? 60_000);
	});
	const run = async () => {
		await s.expect(220, 'greeting');
		await s.cmd(`EHLO ${o.clientName}`, 250, 'EHLO');
		await s.cmd('STARTTLS', 220, 'STARTTLS');
		s = s.upgrade();
		await s.cmd(`EHLO ${o.clientName}`, 250, 'EHLO');
		await s.cmd('AUTH PLAIN ' + btoa(`\0${o.user}\0${o.pass}`), 235, 'AUTH');
		await s.cmd(`MAIL FROM:${addr(o.from)}`, 250, 'MAIL FROM');
		await s.cmd(`RCPT TO:${addr(o.to)}`, 250, 'RCPT TO');
		await s.cmd('DATA', 354, 'DATA');
		for (const chunk of o.message) await s.write(dotStuff(chunk));
		await s.cmd('.', 250, 'end of DATA');
		await s.cmd('QUIT', 221, 'QUIT').catch(() => {});
	};
	try {
		await Promise.race([run(), timeout]);
	} finally {
		clearTimeout(timer);
		await s.close();
	}
}

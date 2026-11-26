import { parse } from 'yaml';

export interface Parsed {
	data: Record<string, unknown>;
	body: string;
}

/** Splits a Markdown file into its YAML front matter and body. */
export function parseFrontmatter(raw: string): Parsed {
	const text = raw.replace(/^﻿/, '').replace(/\r\n/g, '\n');
	const m = /^---\n([\s\S]*?)\n---[ \t]*(?:\n|$)/.exec(text);
	if (!m) return { data: {}, body: text.trim() };
	const data = parse(m[1]);
	if (data !== null && (typeof data !== 'object' || Array.isArray(data))) {
		throw new Error('front matter must be a list of "key: value" lines');
	}
	return { data: data ?? {}, body: text.slice(m[0].length).trim() };
}

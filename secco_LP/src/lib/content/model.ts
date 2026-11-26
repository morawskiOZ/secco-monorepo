// Content model: turns Markdown files (YAML front matter + body) into typed
// entries. Pure functions; file access lives in ./index.ts.

import { parseFrontmatter } from './frontmatter';

export interface Image {
	src: string;
	alt: string;
}

interface Entry {
	slug: string;
	title: string;
	summary: string;
	cover: Image | null;
	/** ISO date, YYYY-MM-DD. */
	date: string;
	draft: boolean;
	body: string;
}

export interface Article extends Entry {
	tags: string[];
}

export interface Project extends Entry {
	/** Lower numbers come first in the portfolio. */
	order: number;
	location: string;
	year: string;
	area: string;
	gallery: Image[];
}

export interface Page {
	title: string;
	summary: string;
	draft: boolean;
	body: string;
}

const SLUG = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

const str = (v: unknown): string => (v === undefined || v === null ? '' : String(v).trim());

function image(src: unknown, alt: unknown): Image | null {
	const s = str(src);
	return s ? { src: s, alt: str(alt) } : null;
}

function entry(kind: string, slug: string, raw: string): Entry & { data: Record<string, unknown> } {
	const fail = (msg: string): never => {
		throw new Error(`${kind} "${slug}": ${msg}`);
	};
	if (!SLUG.test(slug)) fail('file name may only contain a-z, 0-9 and single hyphens');
	let parsed;
	try {
		parsed = parseFrontmatter(raw);
	} catch (err) {
		return fail(`invalid front matter (${(err as Error).message})`);
	}
	const { data, body } = parsed;
	const title = str(data.title);
	if (!title) fail('title is required');
	const date = str(data.date).slice(0, 10);
	if (!/^\d{4}-\d{2}-\d{2}$/.test(date) || Number.isNaN(Date.parse(date))) {
		fail('date is required in the form YYYY-MM-DD');
	}
	return {
		slug,
		title,
		summary: str(data.summary),
		cover: image(data.cover, data.cover_alt),
		date,
		draft: data.draft === true,
		body,
		data
	};
}

export function parseArticle(slug: string, raw: string): Article {
	const { data, ...e } = entry('article', slug, raw);
	const tags = Array.isArray(data.tags) ? data.tags.map(str).filter(Boolean) : [];
	return { ...e, tags: [...new Set(tags)] };
}

export function parseProject(slug: string, raw: string): Project {
	const { data, ...e } = entry('project', slug, raw);
	const gallery = Array.isArray(data.gallery)
		? data.gallery
				.map((g) => (g && typeof g === 'object' ? image((g as Image).src ?? (g as { image?: unknown }).image, (g as Image).alt) : image(g, '')))
				.filter((g): g is Image => g !== null)
		: [];
	const order = Number(data.order);
	return {
		...e,
		order: Number.isFinite(order) ? order : 1000,
		location: str(data.location),
		year: str(data.year),
		area: str(data.area),
		gallery
	};
}

export function parsePage(name: string, raw: string): Page {
	const { data, body } = parseFrontmatter(raw);
	const title = str(data.title);
	if (!title) throw new Error(`page "${name}": title is required`);
	return { title, summary: str(data.summary), draft: data.draft === true, body };
}

/** Drafts are only built where drafts are switched on (stage2). */
export function visible<T extends { draft: boolean }>(items: T[], showDrafts: boolean): T[] {
	return showDrafts ? items : items.filter((i) => !i.draft);
}

/** Newest first; ties broken by slug so builds are stable. */
export function sortArticles(items: Article[]): Article[] {
	return [...items].sort((a, b) => b.date.localeCompare(a.date) || a.slug.localeCompare(b.slug));
}

/** By `order`, then newest first. */
export function sortProjects(items: Project[]): Project[] {
	return [...items].sort(
		(a, b) => a.order - b.order || b.date.localeCompare(a.date) || a.slug.localeCompare(b.slug)
	);
}

/** URL-safe form of a tag: "Małe mieszkania" -> "male-mieszkania". */
export function tagSlug(tag: string): string {
	return tag
		.toLowerCase()
		.replace(/ł/g, 'l')
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');
}

export interface Tag {
	slug: string;
	name: string;
	count: number;
}

export function collectTags(articles: Article[]): Tag[] {
	const tags = new Map<string, Tag>();
	for (const a of articles) {
		for (const name of a.tags) {
			const slug = tagSlug(name);
			if (!slug) continue;
			const t = tags.get(slug);
			if (t) t.count++;
			else tags.set(slug, { slug, name, count: 1 });
		}
	}
	return [...tags.values()].sort((a, b) => a.name.localeCompare(b.name, 'pl'));
}

const MONTHS = ['stycznia', 'lutego', 'marca', 'kwietnia', 'maja', 'czerwca', 'lipca', 'sierpnia', 'września', 'października', 'listopada', 'grudnia'];

/** "2026-11-05" -> "5 listopada 2026". */
export function formatDate(iso: string): string {
	const [y, m, d] = iso.split('-').map(Number);
	return `${d} ${MONTHS[m - 1]} ${y}`;
}

import { describe, expect, it } from 'vitest';
import { parseFrontmatter } from '../../src/lib/content/frontmatter';
import { collectTags, formatDate, parseArticle, parsePage, parseProject, sortArticles, sortProjects, tagSlug, visible } from '../../src/lib/content/model';

const md = (front: string, body = 'Treść.') => `---\n${front}\n---\n\n${body}\n`;

describe('parseFrontmatter', () => {
	it('splits YAML and body', () => {
		expect(parseFrontmatter('---\ntitle: A\n---\n\n# H\n\ntext\n')).toEqual({ data: { title: 'A' }, body: '# H\n\ntext' });
	});
	it('handles CRLF, a BOM and an empty body', () => {
		expect(parseFrontmatter('﻿---\r\ntitle: A\r\n---\r\n')).toEqual({ data: { title: 'A' }, body: '' });
	});
	it('treats a file without front matter as body only', () => {
		expect(parseFrontmatter('just text')).toEqual({ data: {}, body: 'just text' });
	});
	it('does not end the front matter at a --- inside the body', () => {
		expect(parseFrontmatter('---\ntitle: A\n---\nx\n\n---\n\ny').body).toBe('x\n\n---\n\ny');
	});
	it('keeps dates as text', () => {
		expect(parseFrontmatter('---\ndate: 2026-11-26\n---\n').data.date).toBe('2026-11-26');
	});
});

describe('parseArticle', () => {
	it('reads all fields', () => {
		const a = parseArticle('jak-wybrac-styl', md('title: Jak wybrać styl?\nsummary: Krótko.\ncover: /media/a.jpg\ncover_alt: Salon\ndate: 2026-11-20\ntags: [Porady, Style, Porady]\ndraft: true'));
		expect(a).toEqual({
			slug: 'jak-wybrac-styl', title: 'Jak wybrać styl?', summary: 'Krótko.',
			cover: { src: '/media/a.jpg', alt: 'Salon' }, date: '2026-11-20', draft: true,
			body: 'Treść.', tags: ['Porady', 'Style']
		});
	});
	it('defaults optional fields', () => {
		const a = parseArticle('a', md('title: T\ndate: 2026-01-02'));
		expect(a).toMatchObject({ summary: '', cover: null, draft: false, tags: [] });
	});
	it('accepts a full timestamp as date', () => {
		expect(parseArticle('a', md('title: T\ndate: 2026-01-02T10:00:00Z')).date).toBe('2026-01-02');
	});
	it.each([
		['title: ""\ndate: 2026-01-02', /article "a": title is required/],
		['title: T', /date is required/],
		['title: T\ndate: 2026-13-45', /date is required/],
		['title: [unclosed', /invalid front matter/]
	])('rejects %j', (front, msg) => {
		expect(() => parseArticle('a', md(front))).toThrow(msg);
	});
	it('rejects file names that would make bad URLs', () => {
		expect(() => parseArticle('Zły Slug', md('title: T\ndate: 2026-01-02'))).toThrow(/file name/);
	});
});

describe('parseProject', () => {
	it('reads the gallery and details', () => {
		const p = parseProject('dom', md('title: Dom\ndate: 2026-02-01\norder: 2\nlocation: Wrocław\nyear: 2025\narea: 120 m²\ngallery:\n  - image: /media/1.jpg\n    alt: Kuchnia\n  - /media/2.jpg\n  - image: ""'));
		expect(p).toMatchObject({
			order: 2, location: 'Wrocław', year: '2025', area: '120 m²',
			gallery: [{ src: '/media/1.jpg', alt: 'Kuchnia' }, { src: '/media/2.jpg', alt: '' }]
		});
	});
	it('puts projects without an order last', () => {
		expect(parseProject('dom', md('title: Dom\ndate: 2026-02-01')).order).toBe(1000);
	});
});

describe('parsePage', () => {
	it('needs a title only', () => {
		expect(parsePage('proces', md('title: Proces projektowy', '## 1'))).toEqual({ title: 'Proces projektowy', summary: '', draft: false, body: '## 1' });
		expect(() => parsePage('proces', 'x')).toThrow(/title is required/);
	});
});

describe('collections', () => {
	const art = (slug: string, date: string, draft = false, tags: string[] = []) =>
		({ ...parseArticle(slug, md(`title: T\ndate: ${date}`)), draft, tags });
	it('hides drafts unless drafts are on', () => {
		const items = [art('a', '2026-01-01'), art('b', '2026-01-02', true)];
		expect(visible(items, false).map((i) => i.slug)).toEqual(['a']);
		expect(visible(items, true)).toHaveLength(2);
	});
	it('sorts articles newest first', () => {
		expect(sortArticles([art('b', '2026-01-01'), art('c', '2026-03-01'), art('a', '2026-01-01')]).map((a) => a.slug)).toEqual(['c', 'a', 'b']);
	});
	it('sorts projects by order, then newest', () => {
		const p = (slug: string, order: string, date: string) => parseProject(slug, md(`title: T\ndate: ${date}${order}`));
		expect(sortProjects([p('x', '', '2026-05-01'), p('y', '\norder: 2', '2026-01-01'), p('z', '\norder: 1', '2025-01-01'), p('w', '', '2026-06-01')]).map((a) => a.slug)).toEqual(['z', 'y', 'w', 'x']);
	});
	it('makes URL-safe tags', () => {
		expect(tagSlug('Małe mieszkania')).toBe('male-mieszkania');
		expect(tagSlug(' Żółć & styl ')).toBe('zolc-styl');
	});
	it('counts tags across articles', () => {
		expect(collectTags([art('a', '2026-01-01', false, ['Porady', 'Kuchnia']), art('b', '2026-01-01', false, ['porady'])])).toEqual([
			{ slug: 'kuchnia', name: 'Kuchnia', count: 1 },
			{ slug: 'porady', name: 'Porady', count: 2 }
		]);
	});
	it('formats dates in Polish', () => {
		expect(formatDate('2026-11-05')).toBe('5 listopada 2026');
	});
});

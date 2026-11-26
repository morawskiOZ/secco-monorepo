import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { BASE_URLS, buildSitemap, contentUrls } from '../../src/lib/content/sitemap';

describe('sitemap', () => {
	it('is byte-identical to the former static file when there is no content', () => {
		const before = readFileSync(new URL('../fixtures/sitemap-landing-only.xml', import.meta.url), 'utf8');
		expect(buildSitemap('https://www.seccostudio.com', BASE_URLS)).toBe(before);
	});
	it('lists published content and leaves out drafts and empty sections', () => {
		const urls = contentUrls(
			[{ slug: 'a', date: '2026-05-01', draft: false }, { slug: 'b', date: '2026-06-01', draft: true }],
			[{ slug: 'p', date: '2026-01-01', draft: true }],
			['porady'],
			true
		);
		expect(urls.map((u) => u.path)).toEqual(['/proces-projektowy', '/tresci', '/tresci/a', '/tresci/tag/porady']);
		expect(urls[1].lastmod).toBe('2026-05-01');
	});
});

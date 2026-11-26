export interface SitemapUrl {
	path: string;
	/** YYYY-MM-DD */
	lastmod: string;
	changefreq: 'weekly' | 'monthly' | 'yearly';
	priority: string;
}

/** The pages that exist with or without the content sections. */
export const BASE_URLS: SitemapUrl[] = [
	{ path: '/', lastmod: '2026-04-15', changefreq: 'monthly', priority: '1.0' },
	{ path: '/polityka-prywatnosci', lastmod: '2026-04-15', changefreq: 'yearly', priority: '0.3' }
];

export function buildSitemap(siteUrl: string, urls: SitemapUrl[]): string {
	const items = urls.map(
		(u) => `  <url>
    <loc>${siteUrl}${u.path}</loc>
    <lastmod>${u.lastmod}</lastmod>
    <changefreq>${u.changefreq}</changefreq>
    <priority>${u.priority}</priority>
  </url>
`
	);
	return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${items.join('')}</urlset>
`;
}

const latest = (dates: string[], fallback: string) => dates.reduce((a, b) => (b > a ? b : a), fallback);

interface Dated {
	slug: string;
	date: string;
	draft: boolean;
}

/** Content URLs; drafts (stage2) are never listed. */
export function contentUrls(articles: Dated[], projects: Dated[], tagSlugs: string[], hasProcess: boolean): SitemapUrl[] {
	const a = articles.filter((x) => !x.draft);
	const p = projects.filter((x) => !x.draft);
	const urls: SitemapUrl[] = [];
	if (p.length > 0) {
		urls.push({ path: '/projekty', lastmod: latest(p.map((x) => x.date), ''), changefreq: 'monthly', priority: '0.9' });
		for (const x of p) urls.push({ path: `/projekty/${x.slug}`, lastmod: x.date, changefreq: 'yearly', priority: '0.8' });
	}
	if (hasProcess) urls.push({ path: '/proces-projektowy', lastmod: BASE_URLS[0].lastmod, changefreq: 'yearly', priority: '0.7' });
	if (a.length > 0) {
		const newest = latest(a.map((x) => x.date), '');
		urls.push({ path: '/tresci', lastmod: newest, changefreq: 'weekly', priority: '0.8' });
		for (const x of a) urls.push({ path: `/tresci/${x.slug}`, lastmod: x.date, changefreq: 'yearly', priority: '0.6' });
		for (const t of tagSlugs) urls.push({ path: `/tresci/tag/${t}`, lastmod: newest, changefreq: 'weekly', priority: '0.3' });
	}
	return urls;
}

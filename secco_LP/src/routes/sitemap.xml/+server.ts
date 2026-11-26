import { articles, page, projects } from '$lib/content';
import { collectTags } from '$lib/content/model';
import { BASE_URLS, buildSitemap, contentUrls } from '$lib/content/sitemap';
import { SITE_URL } from '$lib/site';

export const prerender = true;

export function GET() {
	const published = articles.filter((a) => !a.draft);
	const process = page('proces-projektowy');
	const urls = [
		...BASE_URLS,
		...contentUrls(articles, projects, collectTags(published).map((t) => t.slug), process !== null && !process.draft)
	];
	return new Response(buildSitemap(SITE_URL, urls), { headers: { 'Content-Type': 'application/xml' } });
}

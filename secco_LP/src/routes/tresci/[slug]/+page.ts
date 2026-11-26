import { articles, media } from '$lib/content';
import { found, requireContent } from '$lib/content/guard';
import { renderMarkdown } from '$lib/content/markdown';
import { CONTENT_ENABLED } from '$lib/flags';

// Prerendered, no client JavaScript. With the content sections off the route is
// not built and the URL falls through to the app shell's 404, as before.
export const prerender = CONTENT_ENABLED;
export const csr = false;

export const entries = () => articles.map((a) => ({ slug: a.slug }));

export function load({ params }) {
	requireContent();
	const article = found(articles.find((a) => a.slug === params.slug));
	return { article, html: renderMarkdown(article.body, media) };
}

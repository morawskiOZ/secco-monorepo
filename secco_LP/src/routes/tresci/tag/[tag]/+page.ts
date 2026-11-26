import { articles } from '$lib/content';
import { found, requireContent } from '$lib/content/guard';
import { collectTags, tagSlug } from '$lib/content/model';
import { CONTENT_ENABLED } from '$lib/flags';

// Prerendered, no client JavaScript. With the content sections off the route is
// not built and the URL falls through to the app shell's 404, as before.
export const prerender = CONTENT_ENABLED;
export const csr = false;

export const entries = () => collectTags(articles).map((t) => ({ tag: t.slug }));

export function load({ params }) {
	requireContent();
	const tags = collectTags(articles);
	const tag = found(tags.find((t) => t.slug === params.tag));
	return { tag, tags, articles: articles.filter((a) => a.tags.some((t) => tagSlug(t) === tag.slug)) };
}

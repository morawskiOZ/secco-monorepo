import { media, projects } from '$lib/content';
import { found, requireContent } from '$lib/content/guard';
import { renderMarkdown } from '$lib/content/markdown';
import { CONTENT_ENABLED } from '$lib/flags';

// Prerendered, no client JavaScript. With the content sections off the route is
// not built and the URL falls through to the app shell's 404, as before.
export const prerender = CONTENT_ENABLED;
export const csr = false;

export const entries = () => projects.map((p) => ({ slug: p.slug }));

export function load({ params }) {
	requireContent();
	const project = found(projects.find((p) => p.slug === params.slug));
	return { project, html: renderMarkdown(project.body, media, { zoom: true }) };
}

import { projects } from '$lib/content';
import { requireContent } from '$lib/content/guard';
import { CONTENT_ENABLED } from '$lib/flags';

// Prerendered, no client JavaScript. With the content sections off the route is
// not built and the URL falls through to the app shell's 404, as before.
export const prerender = CONTENT_ENABLED;
export const csr = false;

export function load() {
	requireContent();
	return { projects };
}

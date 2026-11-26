import { error } from '@sveltejs/kit';
import { CONTENT_ENABLED } from '$lib/flags';

/** With the content sections off, their URLs answer like any unknown path. */
export function requireContent(): void {
	if (!CONTENT_ENABLED) error(404, 'Not Found');
}

export function found<T>(value: T | null | undefined): T {
	if (value === null || value === undefined) error(404, 'Not Found');
	return value;
}

export const LIMIT = 5;
export const WINDOW_MS = 60 * 60 * 1000;

/** Sliding window: at most `limit` hits in any WINDOW_MS. Rejected hits are not recorded. */
export function slidingWindowAllow(
	hits: number[],
	now: number,
	limit = LIMIT
): { allowed: boolean; hits: number[] } {
	const valid = hits.filter((t) => t > now - WINDOW_MS);
	if (valid.length >= limit) return { allowed: false, hits: valid };
	return { allowed: true, hits: [...valid, now] };
}

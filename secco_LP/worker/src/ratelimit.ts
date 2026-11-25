// 5 submissions per hour per IP — same policy as the former Go backend,
// which kept the counters in process memory. Here one Durable Object per IP
// holds them, so the limit is exact across all Cloudflare locations.

import { DurableObject } from 'cloudflare:workers';
import { slidingWindowAllow, WINDOW_MS } from './window';

export class RateLimiter extends DurableObject {
	async allow(limit?: number): Promise<boolean> {
		const now = Date.now();
		const prev = (await this.ctx.storage.get<number[]>('hits')) ?? [];
		const { allowed, hits } = slidingWindowAllow(prev, now, limit);
		await this.ctx.storage.put('hits', hits);
		// Clean up after the window so idle IPs leave nothing stored.
		await this.ctx.storage.setAlarm(now + WINDOW_MS + 1000);
		return allowed;
	}

	async alarm(): Promise<void> {
		await this.ctx.storage.deleteAll();
	}
}

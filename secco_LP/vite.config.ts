import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	// See src/lib/flags.ts.
	define: {
		__SECCO_CONTENT__: JSON.stringify(process.env.SECCO_CONTENT === '1'),
		__SECCO_DRAFTS__: JSON.stringify(process.env.SECCO_DRAFTS === '1')
	},
	server: {
		proxy: {
			'/api': process.env.API_URL || 'http://localhost:8080'
		}
	}
});

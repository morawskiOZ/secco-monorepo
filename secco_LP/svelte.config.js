import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		// Without the content sections (SECCO_CONTENT=1) there is no navigation
		// bar: an empty component takes its place, so none of its code or CSS ships.
		alias: {
			$navbar:
				process.env.SECCO_CONTENT === '1'
					? 'src/lib/components/Navbar.svelte'
					: 'src/lib/components/NoNavbar.svelte'
		},
		// Stylesheets are a few kB in total: inline them so no CSS request blocks first paint.
		inlineStyleThreshold: 20000,
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: '200.html',
			strict: false
		})
	},
	vitePlugin: {
		dynamicCompileOptions: ({ filename }) =>
			filename.includes('node_modules') ? undefined : { runes: true }
	}
};

export default config;

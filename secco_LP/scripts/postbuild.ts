// Runs after `vite build`. SECCO_NOINDEX=1 (stage2) tells search engines to
// stay away from every page and file of the build.

import { appendFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

if (process.env.SECCO_NOINDEX === '1') {
	const headers = join(dirname(fileURLToPath(import.meta.url)), '..', 'build', '_headers');
	appendFileSync(headers, '\n/*\n  X-Robots-Tag: noindex, nofollow\n');
	console.log('postbuild: X-Robots-Tag noindex added to build/_headers');
}

// Runs before `vite dev` / `vite build`:
//   1. resizes the photos in content/media into WebP variants under static/media
//      and writes the manifest the pages read (src/lib/content/media.gen.json);
//   2. when the content sections are on (SECCO_CONTENT=1), copies the CMS admin
//      page to static/admin.
// Everything written here is git-ignored and safe to delete.

import { cpSync, existsSync, mkdirSync, readdirSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import sharp from 'sharp';
import { variantPath, variantWidths, type MediaManifest } from '../src/lib/content/images.ts';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const SRC = join(root, 'content/media');
const OUT = join(root, 'static/media');
const MANIFEST = join(root, 'src/lib/content/media.gen.json');
const RASTER = /\.(jpe?g|png|webp|avif|tiff?)$/i;
const QUALITY = 82;

function walk(dir: string): string[] {
	if (!existsSync(dir)) return [];
	return readdirSync(dir, { withFileTypes: true }).flatMap((e) =>
		e.isDirectory() ? walk(join(dir, e.name)) : [join(dir, e.name)]
	);
}

const manifest: MediaManifest = {};
const expected = new Set<string>();
let written = 0;

const contentOn = process.env.SECCO_CONTENT === '1';

// With the content sections off nothing refers to the photos: ship none.
for (const file of contentOn ? walk(SRC).sort() : []) {
	if (!RASTER.test(file)) continue;
	const publicPath = '/media/' + relative(SRC, file).split('\\').join('/');
	// .rotate() applies the EXIF orientation, so read the size after it.
	const { info } = await sharp(file).rotate().toBuffer({ resolveWithObject: true });
	const widths = variantWidths(info.width);
	const full = widths[widths.length - 1];
	manifest[publicPath] = {
		width: full,
		height: Math.round((info.height * full) / info.width),
		variants: widths.map((width) => ({ width, src: variantPath(publicPath, width) }))
	};
	for (const width of widths) {
		const out = join(root, 'static', variantPath(publicPath, width));
		expected.add(out);
		if (existsSync(out) && statSync(out).mtimeMs >= statSync(file).mtimeMs) continue;
		mkdirSync(dirname(out), { recursive: true });
		await sharp(file).rotate().resize({ width, withoutEnlargement: true }).webp({ quality: QUALITY }).toFile(out);
		written++;
	}
}

// Variants of photos that were removed or renamed.
for (const file of walk(OUT)) if (!expected.has(file)) rmSync(file);

writeFileSync(MANIFEST, JSON.stringify(manifest, null, '\t') + '\n');

const admin = join(root, 'static/admin');
rmSync(admin, { recursive: true, force: true });
if (contentOn) cpSync(join(root, 'cms'), admin, { recursive: true });

console.log(
	`prebuild: ${Object.keys(manifest).length} photos, ${expected.size} variants (${written} written), content sections ${contentOn ? 'on' : 'off'}`
);

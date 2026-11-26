import { describe, expect, it } from 'vitest';
import { imgAttrs, variantPath, variantWidths, type MediaManifest } from '../../src/lib/content/images';
import { renderMarkdown } from '../../src/lib/content/markdown';

const media: MediaManifest = {
	'/media/salon.jpg': {
		width: 1600,
		height: 1067,
		variants: [480, 800, 1200, 1600].map((width) => ({ width, src: `/media/salon-${width}.webp` }))
	}
};

describe('variants', () => {
	it('never upscales', () => {
		expect(variantWidths(3000)).toEqual([480, 800, 1200, 1920]);
		expect(variantWidths(1920)).toEqual([480, 800, 1200, 1920]);
		expect(variantWidths(1000)).toEqual([480, 800, 1000]);
		expect(variantWidths(300)).toEqual([300]);
	});
	it('names variant files next to the original', () => {
		expect(variantPath('/media/dom.1/salon.JPG', 800)).toBe('/media/dom.1/salon-800.webp');
	});
	it('builds img attributes', () => {
		expect(imgAttrs(media['/media/salon.jpg'])).toEqual({
			src: '/media/salon-1200.webp',
			srcset: '/media/salon-480.webp 480w, /media/salon-800.webp 800w, /media/salon-1200.webp 1200w, /media/salon-1600.webp 1600w',
			width: 1600,
			height: 1067
		});
		expect(imgAttrs({ width: 300, height: 200, variants: [{ width: 300, src: '/a-300.webp' }] }).src).toBe('/a-300.webp');
	});
});

describe('renderMarkdown', () => {
	it('renders text', () => {
		expect(renderMarkdown('## Etap 1\n\nOpis **ważny**.', media)).toBe('<h2>Etap 1</h2>\n<p>Opis <strong>ważny</strong>.</p>\n');
	});
	it('renders site images as responsive, sized and lazy', () => {
		const html = renderMarkdown('![Salon "z" widokiem](/media/salon.jpg)', media);
		expect(html).toContain('src="/media/salon-1200.webp"');
		expect(html).toContain('srcset="/media/salon-480.webp 480w,');
		expect(html).toContain('width="1600" height="1067"');
		expect(html).toContain('alt="Salon &quot;z&quot; widokiem"');
		expect(html).toContain('loading="lazy"');
		expect(html).not.toContain('<p>');
		expect(html).not.toContain('figcaption');
	});
	it('uses the image title as caption', () => {
		expect(renderMarkdown('![a](/media/salon.jpg "Salon <po> remoncie")', media)).toContain('<figcaption>Salon &lt;po&gt; remoncie</figcaption>');
	});
	it('links images to the largest file when zoom is on', () => {
		expect(renderMarkdown('![a](/media/salon.jpg)', media, { zoom: true })).toContain('<a href="/media/salon-1600.webp" data-zoom><img');
	});
	it('leaves unknown images as plain tags', () => {
		expect(renderMarkdown('![a](https://example.com/x.jpg)', media)).toBe('<figure><img src="https://example.com/x.jpg" alt="a" loading="lazy" decoding="async" /></figure>\n');
	});
	it('opens external links in a new tab, internal ones in place', () => {
		expect(renderMarkdown('[x](https://example.com)', media)).toContain('<a href="https://example.com" target="_blank" rel="noopener noreferrer">x</a>');
		expect(renderMarkdown('[form](/form)', media)).toContain('<a href="/form">form</a>');
	});
});

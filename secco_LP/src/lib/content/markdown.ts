// Markdown -> HTML for article, project and page bodies. Content comes from
// the repository (trusted editors), so raw HTML in Markdown is allowed.

import { Marked } from 'marked';
import { imgAttrs, largest, type MediaManifest } from './images';

export const esc = (s: string) =>
	s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

export interface RenderOptions {
	/** `sizes` attribute for body images. Default: the 720 px reading column. */
	sizes?: string;
	/** Wrap images in a link to the largest file (the lightbox hooks into it). */
	zoom?: boolean;
}

const DEFAULT_SIZES = '(max-width: 768px) calc(100vw - 48px), 720px';

export function renderMarkdown(markdown: string, media: MediaManifest, opts: RenderOptions = {}): string {
	const marked = new Marked({
		renderer: {
			image({ href, title, text }) {
				const info = media[href];
				const caption = title ? `<figcaption>${esc(title)}</figcaption>` : '';
				if (!info) {
					// External image, or a file the build did not process: plain tag.
					return `<figure><img src="${esc(href)}" alt="${esc(text)}" loading="lazy" decoding="async" />${caption}</figure>`;
				}
				const a = imgAttrs(info);
				const img = `<img src="${a.src}" srcset="${a.srcset}" sizes="${opts.sizes ?? DEFAULT_SIZES}" width="${a.width}" height="${a.height}" alt="${esc(text)}" loading="lazy" decoding="async" />`;
				const inner = opts.zoom ? `<a href="${largest(info).src}" data-zoom>${img}</a>` : img;
				return `<figure>${inner}${caption}</figure>`;
			},
			link({ href, title, tokens }) {
				const external = /^https?:\/\//.test(href) && !href.startsWith('https://www.seccostudio.com');
				const attrs = (title ? ` title="${esc(title)}"` : '') + (external ? ' target="_blank" rel="noopener noreferrer"' : '');
				return `<a href="${esc(href)}"${attrs}>${this.parser.parseInline(tokens)}</a>`;
			}
		}
	});
	// A lone image sits in its own paragraph: <p><figure> is invalid HTML, unwrap it.
	return (marked.parse(markdown, { async: false }) as string).replace(/<p>(<figure>[\s\S]*?<\/figure>)<\/p>/g, '$1');
}

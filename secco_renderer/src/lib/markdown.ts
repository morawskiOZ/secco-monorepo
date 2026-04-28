import { Marked } from 'marked';

const marked = new Marked();

const renderer = {
	image({ href, title, text }: { href: string; title: string | null; text: string }): string {
		const alt = text || '';
		const titleAttr = title ? ` title="${title}"` : '';
		const isR2 = href.includes('assets.seccostudio.com') || href.includes('.r2.dev');

		if (isR2) {
			return `<figure class="content-image">
				<picture>
					<source
						media="(max-width: 480px)"
						srcset="/cdn-cgi/image/w=480,quality=90,format=auto/${href}"
					/>
					<source
						media="(max-width: 768px)"
						srcset="/cdn-cgi/image/w=768,quality=90,format=auto/${href}"
					/>
					<source
						media="(max-width: 1200px)"
						srcset="/cdn-cgi/image/w=1200,quality=90,format=auto/${href}"
					/>
					<img
						src="/cdn-cgi/image/w=1920,quality=95,format=auto/${href}"
						alt="${alt}"${titleAttr}
						loading="lazy"
						decoding="async"
					/>
				</picture>
				${alt ? `<figcaption>${alt}</figcaption>` : ''}
			</figure>`;
		}

		return `<figure class="content-image">
			<img src="${href}" alt="${alt}"${titleAttr} loading="lazy" decoding="async" />
			${alt ? `<figcaption>${alt}</figcaption>` : ''}
		</figure>`;
	}
};

marked.use({ renderer });

export function renderMarkdown(markdown: string): string {
	return marked.parse(markdown) as string;
}

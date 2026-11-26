<script lang="ts">
	import { media } from '$lib/content';
	import { imgAttrs } from '$lib/content/images';
	import type { Image } from '$lib/content/model';

	interface Props {
		image: Image;
		/** How wide the image is shown, e.g. "(max-width: 768px) 100vw, 580px". */
		sizes: string;
		/** First image on the page: load at once instead of lazily. */
		eager?: boolean;
	}
	let { image, sizes, eager = false }: Props = $props();

	const a = $derived.by(() => {
		const info = media[image.src];
		// A typo in a content file should stop the build, not ship a broken image.
		if (!info && image.src.startsWith('/media/')) {
			throw new Error(`Image not found in content/media: ${image.src}`);
		}
		return info ? imgAttrs(info) : null;
	});
</script>

{#if a}
	<img
		src={a.src}
		srcset={a.srcset}
		{sizes}
		width={a.width}
		height={a.height}
		alt={image.alt}
		loading={eager ? 'eager' : 'lazy'}
		fetchpriority={eager ? 'high' : undefined}
		decoding="async"
	/>
{:else}
	<img src={image.src} alt={image.alt} loading={eager ? 'eager' : 'lazy'} decoding="async" />
{/if}

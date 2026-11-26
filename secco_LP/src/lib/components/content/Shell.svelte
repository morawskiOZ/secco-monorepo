<script lang="ts">
	import type { Snippet } from 'svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import Footer from '$lib/components/landing/Footer.svelte';
	import { SITE_NAME, SITE_URL } from '$lib/site';
	import { media } from '$lib/content';
	import { imgAttrs } from '$lib/content/images';

	interface Props {
		/** Path of this page, e.g. "/projekty/dom-pod-lasem". */
		path: string;
		title: string;
		description: string;
		/** Content path of the image used when the page is shared. */
		image?: string;
		ogType?: 'website' | 'article';
		draft?: boolean;
		/** Narrow reading column instead of the full-width grid. */
		narrow?: boolean;
		children: Snippet;
	}
	let { path, title, description, image, ogType = 'website', draft = false, narrow = false, children }: Props = $props();

	const url = $derived(SITE_URL + path);
	const ogImage = $derived(SITE_URL + (image && media[image] ? imgAttrs(media[image]).src : '/images/hero.webp'));
</script>

<svelte:head>
	<title>{title} — {SITE_NAME}</title>
	<meta name="description" content={description} />
	<link rel="canonical" href={url} />
	<meta property="og:type" content={ogType} />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:url" content={url} />
	<meta property="og:site_name" content={SITE_NAME} />
	<meta property="og:locale" content="pl_PL" />
	<meta property="og:image" content={ogImage} />
	{#if draft}<meta name="robots" content="noindex" />{/if}
</svelte:head>

<div class="page">
	<Navbar current={path} />
	<main class="main" class:narrow>
		{#if draft}<p class="draft">Szkic — widoczny tylko w wersji testowej</p>{/if}
		{@render children()}
	</main>
	<Footer />
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
		background-color: var(--color-bg-page);
	}

	.main {
		flex: 1;
		width: 100%;
		max-width: var(--max-width);
		margin: 0 auto;
		padding: var(--spacing-2xl) var(--spacing-2xl) var(--spacing-3xl);
	}

	.main.narrow {
		max-width: calc(720px + 2 * var(--spacing-2xl));
	}

	.draft {
		width: fit-content;
		background-color: var(--color-accent);
		color: var(--color-white);
		font-size: var(--font-size-xs);
		letter-spacing: 0.5px;
		text-transform: uppercase;
		padding: 2px 10px;
		border-radius: var(--radius-sm);
		margin-bottom: var(--spacing-md);
	}

	@media (max-width: 768px) {
		.main {
			padding: var(--spacing-xl) var(--spacing-lg) var(--spacing-2xl);
		}
	}
</style>

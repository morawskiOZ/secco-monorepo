<script lang="ts">
	import Shell from '$lib/components/content/Shell.svelte';
	import PageTitle from '$lib/components/content/PageTitle.svelte';
	import Picture from '$lib/components/content/Picture.svelte';
	import Prose from '$lib/components/content/Prose.svelte';
	import { media } from '$lib/content';
	import { largest } from '$lib/content/images';
	import lightbox from '$lib/content/lightbox.js?raw';

	let { data } = $props();
	const p = $derived(data.project);
	const facts = $derived([p.location, p.area, p.year].filter(Boolean));
	const zoomHref = (src: string) => (media[src] ? largest(media[src]).src : src);
</script>

<Shell
	path="/projekty/{p.slug}"
	title={p.title}
	description={p.summary || `Projekt wnętrza: ${p.title} — Secco Studio.`}
	image={p.cover?.src}
	ogType="article"
	draft={p.draft}
>
	<a class="back" href="/projekty">← Wróć do projektów</a>
	<PageTitle title={p.title} lead={p.summary} />
	{#if facts.length > 0}<p class="facts">{facts.join(' · ')}</p>{/if}

	{#if p.cover}
		<a class="hero" href={zoomHref(p.cover.src)} data-zoom>
			<Picture image={p.cover} sizes="(max-width: 1296px) calc(100vw - 48px), 1200px" eager />
		</a>
	{/if}

	{#if data.html}<div class="text"><Prose html={data.html} /></div>{/if}

	{#if p.gallery.length > 0}
		<ul class="gallery">
			{#each p.gallery as image}
				<li>
					<a href={zoomHref(image.src)} data-zoom>
						<Picture {image} sizes="(max-width: 768px) calc(100vw - 48px), 588px" />
					</a>
				</li>
			{/each}
		</ul>
	{/if}

	<p class="cta"><a href="/form">Wypełnij formularz wyceny</a></p>

	<dialog class="lightbox" aria-label="Powiększone zdjęcie">
		<img alt="" />
		<button class="lightbox-close" aria-label="Zamknij">✕</button>
		<button class="lightbox-prev" aria-label="Poprzednie zdjęcie">‹</button>
		<button class="lightbox-next" aria-label="Następne zdjęcie">›</button>
		<span class="lightbox-count"></span>
	</dialog>
	{@html `<script>${lightbox}</script>`}
</Shell>

<style>
	.back {
		display: inline-block;
		font-size: var(--font-size-sm);
		margin-bottom: var(--spacing-lg);
	}

	.facts {
		color: var(--color-text-light);
		font-size: var(--font-size-sm);
		letter-spacing: 0.5px;
		text-transform: uppercase;
		margin: calc(-1 * var(--spacing-md)) 0 var(--spacing-xl);
	}

	.hero {
		display: block;
		margin-bottom: var(--spacing-2xl);
	}

	.hero :global(img) {
		width: 100%;
		max-height: 80vh;
		object-fit: cover;
		border-radius: var(--radius-md);
	}

	.text {
		max-width: 720px;
		margin: 0 auto var(--spacing-2xl);
	}

	.gallery {
		list-style: none;
		columns: 2;
		column-gap: var(--spacing-lg);
	}

	.gallery li {
		break-inside: avoid;
		margin-bottom: var(--spacing-lg);
	}

	.gallery :global(img) {
		width: 100%;
		border-radius: var(--radius-md);
	}

	.cta {
		text-align: center;
		margin-top: var(--spacing-2xl);
	}

	.cta a {
		display: inline-block;
		background-color: var(--color-accent);
		color: var(--color-white);
		padding: var(--spacing-md) var(--spacing-2xl);
		border-radius: var(--radius-md);
		font-weight: 600;
	}

	.cta a:hover {
		opacity: 0.9;
		text-decoration: none;
	}

	.lightbox {
		border: 0;
		padding: 0;
		margin: auto;
		background: transparent;
		max-width: 100vw;
		max-height: 100vh;
		overflow: visible;
	}

	.lightbox::backdrop {
		background-color: rgba(20, 20, 20, 0.92);
	}

	.lightbox img {
		max-width: 96vw;
		max-height: 92vh;
		width: auto;
		height: auto;
	}

	.lightbox button {
		position: fixed;
		width: 48px;
		height: 48px;
		border: 0;
		border-radius: var(--radius-full);
		background-color: rgba(0, 0, 0, 0.5);
		color: var(--color-white);
		font-size: 1.5rem;
		line-height: 1;
		cursor: pointer;
	}

	.lightbox-close {
		top: 12px;
		right: 12px;
	}

	.lightbox-prev {
		left: 12px;
		top: calc(50% - 24px);
	}

	.lightbox-next {
		right: 12px;
		top: calc(50% - 24px);
	}

	.lightbox-count {
		position: fixed;
		bottom: 12px;
		left: 0;
		right: 0;
		text-align: center;
		color: var(--color-white);
		font-size: var(--font-size-sm);
	}

	@media (max-width: 768px) {
		.gallery {
			columns: 1;
		}
	}
</style>

<script lang="ts">
	import Shell from '$lib/components/content/Shell.svelte';
	import PageTitle from '$lib/components/content/PageTitle.svelte';
	import Picture from '$lib/components/content/Picture.svelte';
	import Prose from '$lib/components/content/Prose.svelte';
	import { formatDate, tagSlug } from '$lib/content/model';
	import { SITE_NAME, SITE_URL } from '$lib/site';

	let { data } = $props();
	const a = $derived(data.article);
	const jsonLd = $derived({
		'@context': 'https://schema.org',
		'@type': 'BlogPosting',
		headline: a.title,
		description: a.summary,
		datePublished: a.date,
		mainEntityOfPage: `${SITE_URL}/tresci/${a.slug}`,
		author: { '@type': 'Organization', name: SITE_NAME, url: SITE_URL },
		publisher: { '@type': 'Organization', name: SITE_NAME, url: SITE_URL }
	});
</script>

<svelte:head>
	<meta property="article:published_time" content={a.date} />
	{@html `<script type="application/ld+json">${JSON.stringify(jsonLd).replace(/</g, '\\u003c')}</script>`}
</svelte:head>

<Shell
	path="/tresci/{a.slug}"
	title={a.title}
	description={a.summary || `${a.title} — artykuł Secco Studio.`}
	image={a.cover?.src}
	ogType="article"
	draft={a.draft}
	narrow
>
	<a class="back" href="/tresci">← Wróć do treści</a>
	<article>
		<PageTitle title={a.title} lead={a.summary} />
		<p class="meta">
			<time datetime={a.date}>{formatDate(a.date)}</time>
			{#each a.tags as tag}<a href="/tresci/tag/{tagSlug(tag)}">{tag}</a>{/each}
		</p>
		{#if a.cover}
			<div class="cover"><Picture image={a.cover} sizes="(max-width: 816px) calc(100vw - 48px), 720px" eager /></div>
		{/if}
		<Prose html={data.html} />
	</article>
	<p class="cta">Planujesz remont lub urządzasz nowe wnętrze? <a href="/form">Wypełnij formularz wyceny</a>.</p>
</Shell>

<style>
	.back {
		display: inline-block;
		font-size: var(--font-size-sm);
		margin-bottom: var(--spacing-lg);
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--spacing-sm) var(--spacing-md);
		font-size: var(--font-size-sm);
		color: var(--color-text-light);
		margin: calc(-1 * var(--spacing-md)) 0 var(--spacing-xl);
	}

	.meta a {
		text-decoration: underline;
	}

	.cover {
		margin-bottom: var(--spacing-xl);
	}

	.cover :global(img) {
		width: 100%;
		max-height: 480px;
		object-fit: cover;
		border-radius: var(--radius-md);
	}

	.cta {
		margin-top: var(--spacing-2xl);
		padding-top: var(--spacing-lg);
		border-top: 1px solid var(--color-bg-primary);
	}

	.cta a {
		text-decoration: underline;
	}
</style>

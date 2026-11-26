<script lang="ts">
	import Picture from './Picture.svelte';
	import { formatDate, type Article, type Tag } from '$lib/content/model';

	interface Props {
		articles: Article[];
		tags: Tag[];
		/** Slug of the tag being shown, if any. */
		activeTag?: string;
	}
	let { articles, tags, activeTag }: Props = $props();
</script>

{#if tags.length > 0}
	<ul class="tags" aria-label="Tematy">
		<li><a href="/tresci" aria-current={activeTag ? undefined : 'page'}>Wszystkie</a></li>
		{#each tags as tag}
			<li><a href="/tresci/tag/{tag.slug}" aria-current={activeTag === tag.slug ? 'page' : undefined}>{tag.name}</a></li>
		{/each}
	</ul>
{/if}

{#if articles.length === 0}
	<p class="empty">Wkrótce pojawią się tu nasze artykuły.</p>
{:else}
	<ul class="grid">
		{#each articles as article, i}
			<li>
				<a class="card" href="/tresci/{article.slug}">
					{#if article.cover}
						<div class="cover">
							<Picture image={article.cover} sizes="(max-width: 640px) calc(100vw - 48px), (max-width: 1024px) 50vw, 368px" eager={i === 0} />
						</div>
					{/if}
					<h2>{article.title}{#if article.draft}<span class="badge">szkic</span>{/if}</h2>
					{#if article.summary}<p class="summary">{article.summary}</p>{/if}
					<time datetime={article.date}>{formatDate(article.date)}</time>
				</a>
			</li>
		{/each}
	</ul>
{/if}

<style>
	ul {
		list-style: none;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: var(--spacing-sm);
		margin-bottom: var(--spacing-xl);
	}

	.tags a {
		display: block;
		font-size: var(--font-size-sm);
		color: var(--color-text);
		border: 1px solid var(--color-bg-primary);
		border-radius: var(--radius-full);
		padding: 6px 16px;
	}

	.tags a:hover,
	.tags a[aria-current='page'] {
		background-color: var(--color-bg-primary);
		text-decoration: none;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--spacing-xl);
	}

	.card {
		display: block;
		color: var(--color-text);
	}

	.card:hover {
		text-decoration: none;
	}

	.cover {
		aspect-ratio: 4 / 3;
		overflow: hidden;
		border-radius: var(--radius-md);
		margin-bottom: var(--spacing-md);
		background-color: var(--color-bg-primary);
	}

	.cover :global(img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform var(--transition-normal);
	}

	.card:hover .cover :global(img) {
		transform: scale(1.03);
	}

	h2 {
		font-size: var(--font-size-lg);
		font-weight: 600;
		line-height: 1.3;
		margin-bottom: var(--spacing-sm);
	}

	.card:hover h2 {
		color: var(--color-accent);
	}

	.summary {
		color: var(--color-text-light);
		font-size: var(--font-size-sm);
		margin-bottom: var(--spacing-sm);
		display: -webkit-box;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	time {
		font-size: var(--font-size-xs);
		color: var(--color-text-light);
	}

	.badge {
		font-size: var(--font-size-xs);
		font-weight: 400;
		text-transform: uppercase;
		background-color: var(--color-accent);
		color: var(--color-white);
		border-radius: var(--radius-sm);
		padding: 1px 8px;
		margin-left: var(--spacing-sm);
		vertical-align: middle;
	}

	.empty {
		color: var(--color-text-light);
	}

	@media (max-width: 1024px) {
		.grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 640px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>

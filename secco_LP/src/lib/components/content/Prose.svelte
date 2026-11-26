<script lang="ts">
	/** Rendered Markdown (see $lib/content/markdown). */
	let { html }: { html: string } = $props();
</script>

<div class="prose">{@html html}</div>

<style>
	.prose {
		line-height: 1.8;
		color: var(--color-text);
	}

	.prose :global(h2) {
		font-size: var(--font-size-xl);
		font-weight: 600;
		line-height: 1.3;
		margin: var(--spacing-2xl) 0 var(--spacing-md);
	}

	.prose :global(h3) {
		font-size: var(--font-size-lg);
		font-weight: 600;
		line-height: 1.3;
		margin: var(--spacing-xl) 0 var(--spacing-sm);
	}

	.prose :global(:is(h2, h3):first-child) {
		margin-top: 0;
	}

	.prose :global(p),
	.prose :global(ul),
	.prose :global(ol),
	.prose :global(blockquote) {
		margin-bottom: var(--spacing-md);
	}

	.prose :global(ul),
	.prose :global(ol) {
		padding-left: var(--spacing-lg);
	}

	.prose :global(li) {
		margin-bottom: var(--spacing-xs);
	}

	.prose :global(a) {
		text-decoration: underline;
	}

	.prose :global(blockquote) {
		border-left: 3px solid var(--color-bg-primary);
		padding-left: var(--spacing-md);
		font-style: italic;
	}

	.prose :global(hr) {
		border: 0;
		border-top: 1px solid var(--color-bg-primary);
		margin: var(--spacing-xl) 0;
	}

	.prose :global(figure) {
		margin: var(--spacing-xl) 0;
	}

	.prose :global(figure img) {
		width: 100%;
		border-radius: var(--radius-md);
	}

	.prose :global(figcaption) {
		font-size: var(--font-size-sm);
		color: var(--color-text-light);
		text-align: center;
		margin-top: var(--spacing-sm);
	}
</style>

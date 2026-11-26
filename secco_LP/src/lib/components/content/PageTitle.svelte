<script lang="ts">
	let { title, lead = '' }: { title: string; lead?: string } = $props();
</script>

<header>
	<h1>{title}</h1>
	{#if lead}<p>{lead}</p>{/if}
</header>

<style>
	header {
		margin-bottom: var(--spacing-xl);
	}

	h1 {
		font-size: var(--font-size-2xl);
		font-weight: 700;
		line-height: 1.2;
		letter-spacing: 0.02em;
	}

	p {
		margin-top: var(--spacing-sm);
		font-size: var(--font-size-lg);
		color: var(--color-text-light);
		max-width: 720px;
	}
</style>

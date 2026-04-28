<script lang="ts">
	import type { FormConfig } from '$lib/config/questions';

	let { config }: { config: FormConfig } = $props();
</script>

<div class="thanks">
	<h1 class="thanks-title">{config.thankYou.title}</h1>
	<p class="thanks-text">{config.thankYou.text}</p>
	<a href="/" class="thanks-link">Wróć na stronę główną</a>
</div>

<style>
	.thanks {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		min-height: 60vh;
		padding: var(--spacing-xl);
	}

	.thanks-title {
		font-size: var(--font-size-3xl);
		color: var(--color-text);
		margin-bottom: var(--spacing-xl);
	}

	.thanks-text {
		font-size: var(--font-size-lg);
		color: var(--color-text-light);
		max-width: 560px;
		line-height: 1.8;
		margin-bottom: var(--spacing-2xl);
	}

	.thanks-link {
		color: var(--color-accent);
		font-size: var(--font-size-base);
	}
</style>

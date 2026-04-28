<script lang="ts">
	import type { FormConfig } from '$lib/config/questions';

	let { config }: { config: FormConfig } = $props();
</script>

<div class="gdpr-notice" role="note" aria-label="Informacja o przetwarzaniu danych">
	<p>
		{config.gdprText}
		Szczegóły w <a href={config.privacyPolicyUrl} target="_blank" rel="noopener noreferrer">Polityce Prywatności</a>.
	</p>
</div>

<style>
	.gdpr-notice {
		margin-top: var(--spacing-xl);
		padding: var(--spacing-md);
		background-color: rgba(0, 0, 0, 0.03);
		border-radius: var(--radius-md);
	}

	.gdpr-notice p {
		font-size: var(--font-size-xs);
		color: var(--color-text-light);
		line-height: 1.6;
	}

	.gdpr-notice a {
		color: var(--color-accent);
		text-decoration: underline;
	}
</style>

<script lang="ts">
	let {
		showBack,
		showNext,
		showSubmit,
		submitting = false,
		onback,
		onnext,
		onsubmit
	}: {
		showBack: boolean;
		showNext: boolean;
		showSubmit: boolean;
		submitting?: boolean;
		onback: () => void;
		onnext: () => void;
		onsubmit: () => void;
	} = $props();
</script>

<div class="nav-buttons">
	{#if showBack}
		<button type="button" class="nav-btn nav-btn-back" onclick={onback} aria-label="Poprzednie pytanie">
			<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
				<path d="M15 18l-6-6 6-6" />
			</svg>
		</button>
	{/if}
	<div class="nav-spacer"></div>
	{#if showNext}
		<button type="button" class="nav-btn nav-btn-next" onclick={onnext} aria-label="Następne pytanie">
			OK
			<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
				<path d="M9 18l6-6-6-6" />
			</svg>
		</button>
	{/if}
	{#if showSubmit}
		<button type="button" class="nav-btn nav-btn-submit" onclick={onsubmit} disabled={submitting} aria-label="Wyślij formularz">
			{#if submitting}
				Wysyłanie...
			{:else}
				Wyślij
			{/if}
		</button>
	{/if}
</div>

<style>
	.nav-buttons {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
		margin-top: var(--spacing-2xl);
	}

	.nav-spacer {
		flex: 1;
	}

	.nav-btn {
		display: inline-flex;
		align-items: center;
		gap: var(--spacing-xs);
		padding: var(--spacing-sm) var(--spacing-xl);
		border: none;
		border-radius: var(--radius-md);
		font-family: var(--font-family);
		font-size: var(--font-size-base);
		cursor: pointer;
		transition: opacity var(--transition-fast);
	}

	.nav-btn:hover:not(:disabled) {
		opacity: 0.85;
	}

	.nav-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.nav-btn-back {
		background: transparent;
		color: var(--color-text-light);
		padding: var(--spacing-sm) var(--spacing-md);
	}

	.nav-btn-next {
		background-color: var(--color-accent);
		color: var(--color-white);
		font-weight: 600;
	}

	.nav-btn-submit {
		background-color: var(--color-accent);
		color: var(--color-white);
		font-weight: 600;
		padding: var(--spacing-md) var(--spacing-2xl);
	}
</style>

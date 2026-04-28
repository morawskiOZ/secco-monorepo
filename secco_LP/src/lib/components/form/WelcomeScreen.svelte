<script lang="ts">
	import type { FormConfig } from '$lib/config/questions';

	let { config, onstart }: { config: FormConfig; onstart: () => void } = $props();
</script>

<div class="welcome">
	<div class="welcome-image">
		<img
			src="/images/hero-form.webp"
			alt="Wnętrze zaprojektowane przez Secco Studio"
			width="500"
			height="1202"
			loading="eager"
		/>
	</div>
	<div class="welcome-content">
		<p class="welcome-text">{config.welcome.text}</p>
		<button type="button" class="welcome-btn" onclick={onstart}>Rozpocznij</button>
	</div>
</div>

<style>
	.welcome {
		display: grid;
		grid-template-columns: auto 1fr;
		min-height: 100vh;
		align-items: center;
		gap: var(--spacing-3xl);
		padding: var(--spacing-2xl);
		max-width: 1300px;
		margin: 0 auto;
	}

	.welcome-image {
		display: flex;
	}

	.welcome-image img {
		max-height: 85vh;
		width: auto;
		display: block;
		border-radius: var(--radius-lg);
	}

	.welcome-content {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		text-align: center;
		padding: var(--spacing-xl);
	}

	.welcome-text {
		font-size: var(--font-size-2xl);
		color: var(--color-text);
		line-height: 1.5;
		margin-bottom: var(--spacing-xl);
		max-width: 520px;
	}

	.welcome-btn {
		background-color: var(--color-accent);
		color: var(--color-white);
		border: none;
		padding: var(--spacing-md) var(--spacing-3xl);
		border-radius: var(--radius-md);
		font-family: var(--font-family);
		font-size: var(--font-size-lg);
		font-weight: 600;
		cursor: pointer;
		transition: opacity var(--transition-fast);
	}

	.welcome-btn:hover {
		opacity: 0.9;
	}

	@media (max-width: 1024px) {
		.welcome-text {
			font-size: var(--font-size-xl);
		}
	}

	@media (max-width: 768px) {
		.welcome {
			grid-template-columns: 1fr;
			align-items: center;
			justify-content: center;
			padding: var(--spacing-xl) var(--spacing-2xl);
			gap: 0;
		}

		.welcome-image {
			display: none;
		}

		.welcome-content {
			padding: 0;
		}

		.welcome-text {
			font-size: var(--font-size-lg);
			margin-bottom: var(--spacing-xl);
		}

		.welcome-btn {
			align-self: stretch;
			text-align: center;
			padding: var(--spacing-lg) var(--spacing-2xl);
			font-size: var(--font-size-lg);
		}
	}
</style>

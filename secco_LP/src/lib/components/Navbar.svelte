<script lang="ts">
	import { articles, page, projects } from '$lib/content';

	let { current = '' }: { current?: string } = $props();

	// A section is linked once it has something to show.
	const links = [
		{ href: '/projekty', label: 'Projekty', show: projects.length > 0 },
		{ href: '/tresci', label: 'Treści', show: articles.length > 0 },
		{ href: '/proces-projektowy', label: 'Proces', show: page('proces-projektowy') !== null }
	].filter((l) => l.show);
</script>

{#snippet items()}
	{#each links as link}
		<li>
			<a href={link.href} aria-current={current.startsWith(link.href) ? 'page' : undefined}>{link.label}</a>
		</li>
	{/each}
	<li><a class="cta" href="/form">Formularz wyceny</a></li>
{/snippet}

<header class="navbar">
	<nav class="inner" aria-label="Główna nawigacja">
		<a class="logo" href="/">secco.studio</a>
		<ul class="links">{@render items()}</ul>
		<!-- No JavaScript: <details> opens and closes the mobile menu. -->
		<details class="mobile">
			<summary aria-label="Menu"><span class="bars" aria-hidden="true"></span></summary>
			<ul>{@render items()}</ul>
		</details>
	</nav>
</header>

<style>
	.navbar {
		--navbar-height: 64px;
		position: sticky;
		top: 0;
		z-index: 100;
		background-color: var(--color-bg-primary);
		box-shadow: 0 1px 0 rgba(0, 0, 0, 0.06);
	}

	.inner {
		max-width: var(--max-width);
		margin: 0 auto;
		height: var(--navbar-height);
		padding: 0 var(--spacing-2xl);
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.logo {
		font-size: var(--font-size-lg);
		font-weight: 600;
		color: var(--color-text);
		letter-spacing: 0.02em;
	}

	.logo:hover {
		text-decoration: none;
	}

	ul {
		list-style: none;
	}

	.links {
		display: flex;
		align-items: center;
		gap: var(--spacing-xl);
	}

	a:not(.logo) {
		font-size: 0.9rem;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: var(--color-text);
		padding: 6px 0;
		border-bottom: 2px solid transparent;
		transition: color var(--transition-fast);
	}

	a:not(.logo):hover {
		color: var(--color-accent);
		text-decoration: none;
	}

	a[aria-current='page'] {
		border-bottom-color: var(--color-accent);
	}

	a.cta {
		background-color: var(--color-accent);
		color: var(--color-white);
		border-radius: var(--radius-sm);
		padding: 8px 20px;
		border: 0;
	}

	a.cta:hover {
		color: var(--color-white);
		opacity: 0.9;
	}

	.mobile {
		display: none;
	}

	summary {
		list-style: none;
		cursor: pointer;
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	summary::-webkit-details-marker {
		display: none;
	}

	.bars,
	.bars::before,
	.bars::after {
		display: block;
		width: 24px;
		height: 2px;
		background-color: var(--color-text);
		transition: transform var(--transition-fast);
	}

	.bars {
		position: relative;
	}

	.bars::before,
	.bars::after {
		content: '';
		position: absolute;
		left: 0;
	}

	.bars::before {
		top: -7px;
	}

	.bars::after {
		top: 7px;
	}

	details[open] .bars {
		background-color: transparent;
	}

	details[open] .bars::before {
		transform: translateY(7px) rotate(45deg);
	}

	details[open] .bars::after {
		transform: translateY(-7px) rotate(-45deg);
	}

	@media (max-width: 768px) {
		.navbar {
			--navbar-height: 56px;
		}

		.inner {
			padding: 0 var(--spacing-lg);
		}

		.links {
			display: none;
		}

		.mobile {
			display: block;
		}

		.mobile ul {
			position: absolute;
			top: var(--navbar-height);
			left: 0;
			right: 0;
			background-color: var(--color-bg-primary);
			box-shadow: var(--shadow-md);
			padding: var(--spacing-lg);
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: var(--spacing-lg);
		}

		.mobile a:not(.cta) {
			font-size: var(--font-size-base);
		}
	}
</style>

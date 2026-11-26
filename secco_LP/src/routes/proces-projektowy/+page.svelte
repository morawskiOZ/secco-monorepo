<script lang="ts">
	import Shell from '$lib/components/content/Shell.svelte';
	import PageTitle from '$lib/components/content/PageTitle.svelte';
	import Prose from '$lib/components/content/Prose.svelte';

	let { data } = $props();
	const p = $derived(data.process);
</script>

<Shell
	path="/proces-projektowy"
	title={p.title}
	description={p.summary || 'Jak wygląda współpraca z Secco Studio — proces projektowania wnętrz krok po kroku.'}
	draft={p.draft}
	narrow
>
	<PageTitle title={p.title} lead={p.summary} />
	<Prose html={data.html} />
	<p class="cta"><a href="/form">Wypełnij formularz wyceny</a></p>
</Shell>

<style>
	.cta {
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
</style>

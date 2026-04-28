<script>
	import FormEngine from '$lib/components/form/FormEngine.svelte';
	import { preferencesConfig } from '$lib/config/preferences-questions';
</script>

<svelte:head>
	<title>Ankieta preferencji — Secco Studio</title>
	<meta name="description" content="Ankieta preferencji klienta Secco Studio." />
	<meta name="robots" content="noindex" />
</svelte:head>

<FormEngine config={preferencesConfig} submitUrl="/api/submit-preferences" />

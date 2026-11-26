<script lang="ts">
	import Shell from '$lib/components/content/Shell.svelte';
	import PageTitle from '$lib/components/content/PageTitle.svelte';
	import ArticleList from '$lib/components/content/ArticleList.svelte';

	let { data } = $props();
</script>

<Shell
	path="/tresci"
	title="Treści — porady i inspiracje wnętrzarskie"
	description="Porady, inspiracje i kulisy pracy architekta wnętrz — artykuły Secco Studio z Wrocławia."
	image={data.articles[0]?.cover?.src}
>
	<PageTitle title="Treści" lead="Porady, inspiracje i kulisy naszej pracy." />
	<ArticleList articles={data.articles} tags={data.tags} />
</Shell>

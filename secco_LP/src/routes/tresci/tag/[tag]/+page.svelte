<script lang="ts">
	import Shell from '$lib/components/content/Shell.svelte';
	import PageTitle from '$lib/components/content/PageTitle.svelte';
	import ArticleList from '$lib/components/content/ArticleList.svelte';

	let { data } = $props();
</script>

<Shell
	path="/tresci/tag/{data.tag.slug}"
	title="{data.tag.name} — Treści"
	description="Artykuły Secco Studio na temat: {data.tag.name}."
	image={data.articles[0]?.cover?.src}
>
	<PageTitle title="Treści: {data.tag.name}" />
	<ArticleList articles={data.articles} tags={data.tags} activeTag={data.tag.slug} />
</Shell>

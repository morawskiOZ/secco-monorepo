<script lang="ts">
	import Shell from '$lib/components/content/Shell.svelte';
	import PageTitle from '$lib/components/content/PageTitle.svelte';
	import Picture from '$lib/components/content/Picture.svelte';

	let { data } = $props();
</script>

<Shell
	path="/projekty"
	title="Projekty wnętrz"
	description="Wybrane realizacje Secco Studio — projekty wnętrz domów i mieszkań we Wrocławiu i okolicach."
	image={data.projects[0]?.cover?.src}
>
	<PageTitle title="Projekty" lead="Wybrane realizacje — domy i mieszkania, które zaprojektowaliśmy." />

	{#if data.projects.length === 0}
		<p class="empty">Wkrótce pojawią się tu nasze realizacje.</p>
	{:else}
		<ul class="grid">
			{#each data.projects as project, i}
				<li>
					<a class="card" href="/projekty/{project.slug}">
						<div class="cover">
							{#if project.cover}
								<Picture image={project.cover} sizes="(max-width: 768px) calc(100vw - 48px), 580px" eager={i === 0} />
							{/if}
						</div>
						<h2>{project.title}{#if project.draft}<span class="badge">szkic</span>{/if}</h2>
						{#if project.summary}<p>{project.summary}</p>{/if}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</Shell>

<style>
	.grid {
		list-style: none;
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--spacing-xl) var(--spacing-lg);
	}

	.card {
		display: block;
		color: var(--color-text);
	}

	.card:hover {
		text-decoration: none;
	}

	.cover {
		aspect-ratio: 3 / 2;
		overflow: hidden;
		border-radius: var(--radius-md);
		margin-bottom: var(--spacing-md);
		background-color: var(--color-bg-primary);
	}

	.cover :global(img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform var(--transition-normal);
	}

	.card:hover .cover :global(img) {
		transform: scale(1.03);
	}

	h2 {
		font-size: var(--font-size-lg);
		font-weight: 600;
		line-height: 1.3;
	}

	.card:hover h2 {
		color: var(--color-accent);
	}

	p {
		color: var(--color-text-light);
		font-size: var(--font-size-sm);
		margin-top: var(--spacing-xs);
	}

	.badge {
		font-size: var(--font-size-xs);
		font-weight: 400;
		text-transform: uppercase;
		background-color: var(--color-accent);
		color: var(--color-white);
		border-radius: var(--radius-sm);
		padding: 1px 8px;
		margin-left: var(--spacing-sm);
		vertical-align: middle;
	}

	.empty {
		color: var(--color-text-light);
	}

	@media (max-width: 768px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>

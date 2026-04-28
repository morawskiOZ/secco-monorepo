<script lang="ts">
  import { page } from '$app/state';
  import { getContent, type Content } from '$lib/api/content';
  import ContentEditor from '$lib/components/ContentEditor.svelte';

  const id = $derived(Number(page.params.id));

  let content = $state<Content | null>(null);
  let loading = $state(true);
  let error = $state('');

  $effect(() => {
    loadContent(id);
  });

  async function loadContent(contentId: number) {
    loading = true;
    error = '';
    try {
      content = await getContent(contentId);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Nie udało się załadować projektu';
    } finally {
      loading = false;
    }
  }
</script>

<div class="page-header">
  <h1 class="page-title">Edytuj projekt</h1>
  <a href="/projekty" class="btn btn-secondary">Powrót do listy</a>
</div>

{#if loading}
  <div class="loading-container">
    <div class="spinner"></div>
  </div>
{:else if error}
  <div class="card">
    <p class="error-text">{error}</p>
    <button class="btn btn-secondary" onclick={() => loadContent(id)}>Spróbuj ponownie</button>
  </div>
{:else if content}
  {#key content.id}
    <ContentEditor
      contentType="project"
      existingContent={content}
      backUrl="/projekty"
      showSlug={true}
      showSummary={true}
      showSortOrder={true}
    />
  {/key}
{/if}

<style>
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--spacing-xl);
  }

  .page-header .page-title {
    margin-bottom: 0;
  }

  .loading-container {
    display: flex;
    justify-content: center;
    padding: var(--spacing-3xl);
  }

  .error-text {
    color: var(--color-error);
    margin-bottom: var(--spacing-md);
  }
</style>

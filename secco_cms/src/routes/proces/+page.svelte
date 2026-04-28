<script lang="ts">
  import { listContent, type Content } from '$lib/api/content';
  import ContentEditor from '$lib/components/ContentEditor.svelte';

  let content = $state<Content | null>(null);
  let loading = $state(true);
  let error = $state('');
  let isNew = $state(false);

  $effect(() => {
    loadProcess();
  });

  async function loadProcess() {
    loading = true;
    error = '';
    try {
      const items = await listContent('process');
      if (items.length > 0) {
        content = items[0];
        isNew = false;
      } else {
        content = null;
        isNew = true;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Nie udało się załadować procesu';
    } finally {
      loading = false;
    }
  }
</script>

<div class="page-header">
  <h1 class="page-title">Proces projektowy</h1>
</div>

{#if loading}
  <div class="loading-container">
    <div class="spinner"></div>
  </div>
{:else if error}
  <div class="card">
    <p class="error-text">{error}</p>
    <button class="btn btn-secondary" onclick={loadProcess}>Spróbuj ponownie</button>
  </div>
{:else}
  {#key content?.id ?? 'new'}
    <ContentEditor
      contentType="process"
      existingContent={content}
      backUrl="/proces"
      showSlug={false}
      showSummary={false}
      showSortOrder={false}
      showDelete={false}
      defaultTitle="Proces projektowy"
    />
  {/key}
{/if}

<style>
  .page-header {
    margin-bottom: var(--spacing-xl);
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

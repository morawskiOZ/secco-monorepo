<script lang="ts">
  import { listContent, deleteContent, publishContent, draftContent, type Content } from '$lib/api/content';

  let items = $state<Content[]>([]);
  let loading = $state(true);
  let error = $state('');

  $effect(() => {
    loadArticles();
  });

  async function loadArticles() {
    loading = true;
    error = '';
    try {
      items = await listContent('article');
    } catch (err) {
      error = 'Nie udało się załadować artykułów';
    } finally {
      loading = false;
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Czy na pewno chcesz usunąć ten artykuł?')) return;
    try {
      await deleteContent(id);
      items = items.filter(i => i.id !== id);
    } catch {
      alert('Błąd usuwania');
    }
  }

  async function toggleStatus(item: Content) {
    try {
      const updated = item.status === 'published'
        ? await draftContent(item.id)
        : await publishContent(item.id);
      items = items.map(i => i.id === updated.id ? updated : i);
    } catch {
      alert('Błąd zmiany statusu');
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('pl-PL');
  }
</script>

<div class="page-header">
  <h1 class="page-title">Artykuły</h1>
  <a href="/tresci/new" class="btn btn-primary">Nowy artykuł</a>
</div>

{#if loading}
  <div class="dashboard-loading">
    <div class="spinner"></div>
  </div>
{:else if error}
  <div class="card">
    <p class="error-text">{error}</p>
    <button class="btn btn-secondary" onclick={loadArticles}>Spróbuj ponownie</button>
  </div>
{:else if items.length === 0}
  <div class="empty-state">
    <p>Brak artykułów</p>
    <a href="/tresci/new" class="btn btn-primary">Utwórz pierwszy artykuł</a>
  </div>
{:else}
  <div class="card">
    <table class="data-table">
      <thead>
        <tr>
          <th>Tytuł</th>
          <th>Slug</th>
          <th>Status</th>
          <th>Data</th>
          <th>Akcje</th>
        </tr>
      </thead>
      <tbody>
        {#each items as item (item.id)}
          <tr>
            <td><a href="/tresci/{item.id}">{item.title}</a></td>
            <td>{item.slug}</td>
            <td>
              <span class="badge" class:badge-published={item.status === 'published'} class:badge-draft={item.status === 'draft'}>
                {item.status === 'published' ? 'Opublikowany' : 'Szkic'}
              </span>
            </td>
            <td>{formatDate(item.updated_at)}</td>
            <td>
              <div class="actions">
                <a href="/tresci/{item.id}" class="btn btn-secondary btn-sm">Edytuj</a>
                <button class="btn btn-sm" class:btn-primary={item.status === 'draft'} class:btn-secondary={item.status === 'published'} onclick={() => toggleStatus(item)}>
                  {item.status === 'published' ? 'Cofnij' : 'Publikuj'}
                </button>
                <button class="btn btn-danger btn-sm" onclick={() => handleDelete(item.id)}>Usuń</button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
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

  .dashboard-loading {
    display: flex;
    justify-content: center;
    padding: var(--spacing-3xl);
  }

  .error-text {
    color: var(--color-error);
    margin-bottom: var(--spacing-md);
  }
</style>

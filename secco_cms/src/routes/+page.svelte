<script lang="ts">
  import { listContent, type Content } from '$lib/api/content';
  import { listImages } from '$lib/api/images';
  import { getDeployStatus } from '$lib/api/deploy';

  let articles = $state<Content[]>([]);
  let projects = $state<Content[]>([]);
  let imageCount = $state(0);
  let recentItems = $state<Content[]>([]);
  let deployStatus = $state<{ status: string; last_deploy: string | null } | null>(null);
  let loading = $state(true);

  $effect(() => {
    loadData();
  });

  async function loadData() {
    loading = true;
    try {
      const [a, p, imgs, deploy] = await Promise.all([
        listContent('article').catch(() => [] as Content[]),
        listContent('project').catch(() => [] as Content[]),
        listImages().catch(() => []),
        getDeployStatus().catch(() => null)
      ]);
      articles = a;
      projects = p;
      imageCount = imgs.length;
      deployStatus = deploy;

      // Build recent items from all content, sorted by updated_at
      const allContent = [...a, ...p];
      allContent.sort((x, y) => new Date(y.updated_at).getTime() - new Date(x.updated_at).getTime());
      recentItems = allContent.slice(0, 5);
    } finally {
      loading = false;
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('pl-PL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function contentUrl(item: Content): string {
    if (item.type === 'article') return `/tresci/${item.id}`;
    if (item.type === 'project') return `/projekty/${item.id}`;
    return '/proces';
  }

  function typeLabel(type: string): string {
    if (type === 'article') return 'Artykuł';
    if (type === 'project') return 'Projekt';
    return 'Proces';
  }
</script>

<h1 class="page-title">Panel zarządzania</h1>

{#if loading}
  <div class="dashboard-loading">
    <div class="spinner"></div>
  </div>
{:else}
  <div class="dashboard-cards">
    <div class="card summary-card">
      <span class="summary-count">{articles.length}</span>
      <span class="summary-label">Artykułów</span>
    </div>
    <div class="card summary-card">
      <span class="summary-count">{projects.length}</span>
      <span class="summary-label">Projektów</span>
    </div>
    <div class="card summary-card">
      <span class="summary-count">{imageCount}</span>
      <span class="summary-label">Zdjęć</span>
    </div>
  </div>

  <div class="dashboard-actions">
    <h2 class="section-title">Szybkie akcje</h2>
    <div class="actions-grid">
      <a href="/tresci/new" class="btn btn-primary">Nowy artykuł</a>
      <a href="/projekty/new" class="btn btn-primary">Nowy projekt</a>
      <a href="/proces" class="btn btn-secondary">Edytuj proces</a>
    </div>
  </div>

  {#if recentItems.length > 0}
    <div class="dashboard-recent card">
      <h2 class="section-title">Ostatnie zmiany</h2>
      <table class="data-table">
        <thead>
          <tr>
            <th>Tytuł</th>
            <th>Typ</th>
            <th>Status</th>
            <th>Ostatnia zmiana</th>
          </tr>
        </thead>
        <tbody>
          {#each recentItems as item (item.id)}
            <tr>
              <td><a href={contentUrl(item)}>{item.title}</a></td>
              <td>{typeLabel(item.type)}</td>
              <td>
                <span class="badge" class:badge-published={item.status === 'published'} class:badge-draft={item.status === 'draft'}>
                  {item.status === 'published' ? 'Opublikowany' : 'Szkic'}
                </span>
              </td>
              <td>{formatDate(item.updated_at)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <div class="dashboard-deploy card">
    <h2 class="section-title">Ostatni deploy</h2>
    {#if deployStatus}
      <div class="deploy-info">
        <div class="deploy-row">
          <span class="deploy-label">Status:</span>
          <span class="deploy-value">{deployStatus.status}</span>
        </div>
        {#if deployStatus.last_deploy}
          <div class="deploy-row">
            <span class="deploy-label">Ostatni deploy:</span>
            <span class="deploy-value">{formatDate(deployStatus.last_deploy)}</span>
          </div>
        {:else}
          <div class="deploy-row">
            <span class="deploy-label">Ostatni deploy:</span>
            <span class="deploy-value deploy-none">Brak</span>
          </div>
        {/if}
      </div>
    {:else}
      <p class="deploy-placeholder">Informacja o deploymencie będzie dostępna po wdrożeniu backendu.</p>
    {/if}
  </div>
{/if}

<style>
  .dashboard-loading {
    display: flex;
    justify-content: center;
    padding: var(--spacing-3xl);
  }

  .dashboard-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: var(--spacing-md);
    margin-bottom: var(--spacing-xl);
  }

  .summary-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: var(--spacing-xl);
  }

  .summary-count {
    font-size: var(--font-size-3xl);
    font-weight: 700;
    color: var(--color-accent);
  }

  .summary-label {
    font-size: var(--font-size-sm);
    color: var(--color-accent-secondary);
    margin-top: var(--spacing-xs);
  }

  .section-title {
    font-size: var(--font-size-lg);
    font-weight: 600;
    margin-bottom: var(--spacing-md);
  }

  .dashboard-actions {
    margin-bottom: var(--spacing-xl);
  }

  .actions-grid {
    display: flex;
    gap: var(--spacing-sm);
    flex-wrap: wrap;
  }

  .dashboard-recent {
    margin-bottom: var(--spacing-xl);
  }

  .dashboard-deploy {
    margin-bottom: var(--spacing-xl);
  }

  .deploy-placeholder {
    color: var(--color-accent-secondary);
    font-size: var(--font-size-sm);
  }

  .deploy-info {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
  }

  .deploy-row {
    display: flex;
    gap: var(--spacing-sm);
    font-size: var(--font-size-sm);
  }

  .deploy-label {
    color: var(--color-accent-secondary);
    font-weight: 600;
    min-width: 120px;
  }

  .deploy-value {
    color: var(--color-text);
  }

  .deploy-none {
    color: var(--color-accent-secondary);
    font-style: italic;
  }
</style>

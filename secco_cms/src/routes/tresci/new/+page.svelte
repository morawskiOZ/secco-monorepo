<script lang="ts">
  import ContentEditor from '$lib/components/ContentEditor.svelte';
</script>

<div class="page-header">
  <h1 class="page-title">Nowy artykuł</h1>
  <a href="/tresci" class="btn btn-secondary">Powrót do listy</a>
</div>

<ContentEditor
  contentType="article"
  backUrl="/tresci"
  showSlug={true}
  showSummary={true}
  showSortOrder={false}
  showDelete={false}
/>

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
</style>

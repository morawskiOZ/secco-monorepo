<script lang="ts">
  import { listImages, deleteImage, updateImageAlt, type Image } from '$lib/api/images';
  import ImageUpload from '$lib/components/ImageUpload.svelte';

  let images = $state<Image[]>([]);
  let loading = $state(true);
  let error = $state('');
  let searchQuery = $state('');
  let selectedImage = $state<Image | null>(null);
  let editAlt = $state('');
  let savingAlt = $state(false);
  let copyMessage = $state('');

  const filteredImages = $derived(
    searchQuery
      ? images.filter(img =>
          img.filename.toLowerCase().includes(searchQuery.toLowerCase()) ||
          img.alt_text.toLowerCase().includes(searchQuery.toLowerCase())
        )
      : images
  );

  $effect(() => {
    loadImages();
  });

  async function loadImages() {
    loading = true;
    error = '';
    try {
      images = await listImages();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Nie udało się załadować zdjęć';
    } finally {
      loading = false;
    }
  }

  function handleUpload(uploaded: Image[]) {
    images = [...uploaded, ...images];
  }

  function selectImage(img: Image) {
    selectedImage = img;
    editAlt = img.alt_text || '';
  }

  function closeDetails() {
    selectedImage = null;
  }

  async function saveAlt() {
    if (!selectedImage) return;
    savingAlt = true;
    try {
      const updated = await updateImageAlt(selectedImage.id, editAlt);
      images = images.map(img => img.id === updated.id ? updated : img);
      selectedImage = updated;
    } catch {
      // Silently fail on alt text save
    } finally {
      savingAlt = false;
    }
  }

  async function handleDelete(img: Image) {
    if (!confirm(`Czy na pewno chcesz usunąć "${img.filename}"?`)) return;
    try {
      await deleteImage(img.id);
      images = images.filter(i => i.id !== img.id);
      if (selectedImage?.id === img.id) {
        selectedImage = null;
      }
    } catch {
      alert('Błąd usuwania zdjęcia');
    }
  }

  async function copyUrl(url: string) {
    try {
      await navigator.clipboard.writeText(url);
      copyMessage = 'Skopiowano!';
      setTimeout(() => { copyMessage = ''; }, 2000);
    } catch {
      copyMessage = 'Błąd kopiowania';
      setTimeout(() => { copyMessage = ''; }, 2000);
    }
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('pl-PL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function handleDetailsKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      closeDetails();
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains('details-backdrop')) {
      closeDetails();
    }
  }
</script>

<svelte:window onkeydown={handleDetailsKeydown} />

<div class="page-header">
  <h1 class="page-title">Zdjęcia</h1>
</div>

<div class="card upload-card">
  <ImageUpload onupload={handleUpload} />
</div>

<div class="search-bar">
  <input
    type="text"
    bind:value={searchQuery}
    placeholder="Szukaj po nazwie pliku..."
    class="search-input"
  />
</div>

{#if loading}
  <div class="loading-container">
    <div class="spinner"></div>
  </div>
{:else if error}
  <div class="card">
    <p class="error-text">{error}</p>
    <button class="btn btn-secondary" onclick={loadImages}>Spróbuj ponownie</button>
  </div>
{:else if filteredImages.length === 0}
  <div class="empty-state">
    <p>{searchQuery ? 'Brak wyników wyszukiwania' : 'Brak zdjęć. Prześlij nowe zdjęcia powyżej.'}</p>
  </div>
{:else}
  <div class="images-grid">
    {#each filteredImages as img (img.id)}
      <button
        type="button"
        class="image-card"
        onclick={() => selectImage(img)}
      >
        <div class="image-thumb">
          <img src={img.public_url} alt={img.alt_text || img.filename} loading="lazy" />
        </div>
        <div class="image-name">{img.filename}</div>
      </button>
    {/each}
  </div>
{/if}

{#if selectedImage}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="details-backdrop" onclick={handleBackdropClick} role="dialog" aria-modal="true">
    <div class="details-panel card">
      <div class="details-header">
        <h3>Szczegóły zdjęcia</h3>
        <button type="button" class="details-close" onclick={closeDetails} aria-label="Zamknij">&times;</button>
      </div>

      <div class="details-preview">
        <img src={selectedImage.public_url} alt={selectedImage.alt_text || selectedImage.filename} />
      </div>

      <div class="details-info">
        <div class="detail-row">
          <span class="detail-label">Plik:</span>
          <span>{selectedImage.filename}</span>
        </div>
        {#if selectedImage.width && selectedImage.height}
          <div class="detail-row">
            <span class="detail-label">Wymiary:</span>
            <span>{selectedImage.width} x {selectedImage.height}</span>
          </div>
        {/if}
        <div class="detail-row">
          <span class="detail-label">Rozmiar:</span>
          <span>{formatSize(selectedImage.size_bytes)}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Data:</span>
          <span>{formatDate(selectedImage.created_at)}</span>
        </div>
      </div>

      <div class="details-alt">
        <label for="details-alt-input">Tekst alternatywny</label>
        <input
          id="details-alt-input"
          type="text"
          bind:value={editAlt}
          onblur={saveAlt}
          placeholder="Opis zdjęcia..."
          disabled={savingAlt}
        />
      </div>

      <div class="details-actions">
        <button type="button" class="btn btn-secondary btn-sm" onclick={() => copyUrl(selectedImage!.public_url)}>
          {copyMessage || 'Kopiuj URL'}
        </button>
        <button type="button" class="btn btn-danger btn-sm" onclick={() => handleDelete(selectedImage!)}>
          Usuń
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-header {
    margin-bottom: var(--spacing-xl);
  }

  .upload-card {
    margin-bottom: var(--spacing-md);
  }

  .search-bar {
    margin-bottom: var(--spacing-md);
  }

  .search-input {
    width: 100%;
    max-width: 400px;
    padding: var(--spacing-sm) var(--spacing-md);
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
  }

  .search-input:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(46, 49, 146, 0.15);
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

  .images-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: var(--spacing-md);
  }

  .image-card {
    background: var(--color-white);
    border: 1px solid #e0e0e0;
    border-radius: var(--radius-md);
    overflow: hidden;
    cursor: pointer;
    padding: 0;
    transition: box-shadow var(--transition-fast);
    text-align: left;
  }

  .image-card:hover {
    box-shadow: var(--shadow-md);
  }

  .image-thumb {
    aspect-ratio: 1;
    overflow: hidden;
    background: #f5f5f5;
  }

  .image-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .image-name {
    padding: var(--spacing-xs) var(--spacing-sm);
    font-size: var(--font-size-xs);
    color: var(--color-text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Details modal */
  .details-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--spacing-md);
  }

  .details-panel {
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
  }

  .details-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--spacing-md);
  }

  .details-header h3 {
    font-size: var(--font-size-lg);
    font-weight: 600;
  }

  .details-close {
    background: none;
    border: none;
    font-size: var(--font-size-2xl);
    color: var(--color-accent-secondary);
    cursor: pointer;
    line-height: 1;
    padding: 0;
  }

  .details-close:hover {
    color: var(--color-text);
  }

  .details-preview {
    margin-bottom: var(--spacing-md);
    border-radius: var(--radius-sm);
    overflow: hidden;
    background: #f5f5f5;
  }

  .details-preview img {
    width: 100%;
    max-height: 300px;
    object-fit: contain;
  }

  .details-info {
    margin-bottom: var(--spacing-md);
  }

  .detail-row {
    display: flex;
    gap: var(--spacing-sm);
    padding: var(--spacing-xs) 0;
    font-size: var(--font-size-sm);
  }

  .detail-label {
    color: var(--color-accent-secondary);
    font-weight: 600;
    min-width: 80px;
  }

  .details-alt {
    margin-bottom: var(--spacing-md);
  }

  .details-alt label {
    display: block;
    font-size: var(--font-size-sm);
    font-weight: 600;
    margin-bottom: var(--spacing-xs);
  }

  .details-alt input {
    width: 100%;
    padding: var(--spacing-sm) var(--spacing-md);
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
  }

  .details-alt input:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(46, 49, 146, 0.15);
  }

  .details-actions {
    display: flex;
    gap: var(--spacing-sm);
  }
</style>

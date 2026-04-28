<script lang="ts">
  import { listImages, type Image } from '$lib/api/images';
  import ImageUpload from './ImageUpload.svelte';

  let {
    open = false,
    onselect,
    onclose
  }: {
    open?: boolean;
    onselect?: (url: string, alt: string) => void;
    onclose?: () => void;
  } = $props();

  let images = $state<Image[]>([]);
  let loading = $state(false);
  let selected = $state<Image | null>(null);
  let altText = $state('');

  $effect(() => {
    if (open) {
      loadImages();
      selected = null;
      altText = '';
    }
  });

  async function loadImages() {
    loading = true;
    try {
      images = await listImages();
    } catch {
      images = [];
    } finally {
      loading = false;
    }
  }

  function selectImage(img: Image) {
    selected = img;
    altText = img.alt_text || '';
  }

  function handleInsert() {
    if (!selected) return;
    onselect?.(selected.public_url, altText);
    onclose?.();
  }

  function handleCancel() {
    onclose?.();
  }

  function handleBackdropClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains('modal-backdrop')) {
      onclose?.();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onclose?.();
    }
  }

  function handleUpload(uploaded: Image[]) {
    images = [...uploaded, ...images];
    if (uploaded.length > 0) {
      selectImage(uploaded[0]);
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="modal-backdrop" onclick={handleBackdropClick} role="dialog" aria-modal="true">
    <div class="modal-content">
      <div class="modal-header">
        <h2 class="modal-title">Wybierz obraz</h2>
        <button type="button" class="modal-close" onclick={handleCancel} aria-label="Zamknij">&times;</button>
      </div>

      <div class="modal-body">
        <div class="upload-section">
          <ImageUpload onupload={handleUpload} />
        </div>

        {#if loading}
          <div class="images-loading">
            <div class="spinner"></div>
          </div>
        {:else if images.length === 0}
          <p class="no-images">Brak obrazów. Prześlij nowy obraz powyżej.</p>
        {:else}
          <div class="images-grid">
            {#each images as img (img.id)}
              <button
                type="button"
                class="image-thumb"
                class:selected={selected?.id === img.id}
                onclick={() => selectImage(img)}
              >
                <img src={img.public_url} alt={img.alt_text || img.filename} loading="lazy" />
              </button>
            {/each}
          </div>
        {/if}

        {#if selected}
          <div class="selected-details">
            <label class="alt-label" for="picker-alt">
              Tekst alternatywny
              <input
                id="picker-alt"
                type="text"
                bind:value={altText}
                placeholder="Opis obrazu..."
                class="alt-input"
              />
            </label>
          </div>
        {/if}
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" onclick={handleCancel}>Anuluj</button>
        <button type="button" class="btn btn-primary" onclick={handleInsert} disabled={!selected}>
          Wstaw
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--spacing-md);
  }

  .modal-content {
    background: var(--color-white);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    width: 100%;
    max-width: 800px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--spacing-md) var(--spacing-lg);
    border-bottom: 1px solid #e0e0e0;
  }

  .modal-title {
    font-size: var(--font-size-lg);
    font-weight: 600;
  }

  .modal-close {
    background: none;
    border: none;
    font-size: var(--font-size-2xl);
    color: var(--color-accent-secondary);
    cursor: pointer;
    line-height: 1;
    padding: 0;
  }

  .modal-close:hover {
    color: var(--color-text);
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-lg);
  }

  .upload-section {
    margin-bottom: var(--spacing-lg);
  }

  .images-loading {
    display: flex;
    justify-content: center;
    padding: var(--spacing-xl);
  }

  .no-images {
    text-align: center;
    color: var(--color-accent-secondary);
    padding: var(--spacing-xl);
  }

  .images-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: var(--spacing-sm);
  }

  .image-thumb {
    aspect-ratio: 1;
    border: 2px solid transparent;
    border-radius: var(--radius-sm);
    overflow: hidden;
    cursor: pointer;
    padding: 0;
    background: #f5f5f5;
    transition: border-color var(--transition-fast);
  }

  .image-thumb:hover {
    border-color: var(--color-accent-secondary);
  }

  .image-thumb.selected {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(46, 49, 146, 0.3);
  }

  .image-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .selected-details {
    margin-top: var(--spacing-md);
    padding: var(--spacing-md);
    background: #f8f8f8;
    border-radius: var(--radius-sm);
  }

  .alt-label {
    display: block;
    font-size: var(--font-size-sm);
    font-weight: 600;
    color: var(--color-text);
  }

  .alt-input {
    display: block;
    width: 100%;
    margin-top: var(--spacing-xs);
    padding: var(--spacing-sm) var(--spacing-md);
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
  }

  .alt-input:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(46, 49, 146, 0.15);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--spacing-sm);
    padding: var(--spacing-md) var(--spacing-lg);
    border-top: 1px solid #e0e0e0;
  }
</style>

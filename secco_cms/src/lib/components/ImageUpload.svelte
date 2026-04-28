<script lang="ts">
  import { uploadImage, type Image } from '$lib/api/images';

  let {
    onupload
  }: {
    onupload?: (images: Image[]) => void;
  } = $props();

  let dragover = $state(false);
  let files = $state<{ file: File; status: 'pending' | 'uploading' | 'done' | 'error'; error?: string; result?: Image }[]>([]);
  let fileInput: HTMLInputElement | undefined = $state();

  const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp'];
  const MAX_SIZE = 20 * 1024 * 1024; // 20MB

  function validateFile(file: File): string | null {
    if (!ALLOWED_TYPES.includes(file.type)) {
      return `Nieprawidłowy format: ${file.type}. Dozwolone: JPEG, PNG, WebP`;
    }
    if (file.size > MAX_SIZE) {
      return `Plik za duży: ${(file.size / 1024 / 1024).toFixed(1)}MB. Maksimum: 20MB`;
    }
    return null;
  }

  function addFiles(newFiles: FileList | File[]) {
    const entries = Array.from(newFiles).map(file => {
      const error = validateFile(file);
      return {
        file,
        status: error ? 'error' as const : 'pending' as const,
        error: error ?? undefined
      };
    });
    files = [...files, ...entries];
    uploadPending();
  }

  async function uploadPending() {
    const pending = files.filter(f => f.status === 'pending');
    const uploaded: Image[] = [];

    for (const entry of pending) {
      const idx = files.indexOf(entry);
      if (idx === -1) continue;

      files[idx] = { ...files[idx], status: 'uploading' };
      files = [...files];

      try {
        const result = await uploadImage(entry.file);
        files[idx] = { ...files[idx], status: 'done', result };
        files = [...files];
        uploaded.push(result);
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Błąd przesyłania';
        files[idx] = { ...files[idx], status: 'error', error: message };
        files = [...files];
      }
    }

    if (uploaded.length > 0) {
      onupload?.(uploaded);
    }
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    dragover = false;
    if (e.dataTransfer?.files) {
      addFiles(e.dataTransfer.files);
    }
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    dragover = true;
  }

  function handleDragLeave() {
    dragover = false;
  }

  function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files) {
      addFiles(input.files);
      input.value = '';
    }
  }

  function clearCompleted() {
    files = files.filter(f => f.status !== 'done' && f.status !== 'error');
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
</script>

<div
  class="upload-zone"
  class:dragover
  ondrop={handleDrop}
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  role="button"
  tabindex="0"
>
  <p class="upload-text">Przeciągnij pliki tutaj lub</p>
  <button type="button" class="btn btn-primary btn-sm" onclick={() => fileInput?.click()}>
    Wybierz pliki
  </button>
  <p class="upload-hint">JPEG, PNG, WebP - max 20MB</p>
  <input
    bind:this={fileInput}
    type="file"
    accept="image/jpeg,image/png,image/webp"
    multiple
    hidden
    onchange={handleFileSelect}
  />
</div>

{#if files.length > 0}
  <div class="upload-list">
    {#each files as entry, i (i)}
      <div class="upload-item" class:error={entry.status === 'error'} class:done={entry.status === 'done'}>
        <span class="upload-filename">{entry.file.name}</span>
        <span class="upload-size">{formatSize(entry.file.size)}</span>
        {#if entry.status === 'uploading'}
          <span class="upload-status uploading">Przesyłanie...</span>
        {:else if entry.status === 'done'}
          <span class="upload-status success">Gotowe</span>
        {:else if entry.status === 'error'}
          <span class="upload-status error-text">{entry.error}</span>
        {:else}
          <span class="upload-status pending">Oczekuje...</span>
        {/if}
      </div>
    {/each}
    {#if files.some(f => f.status === 'done' || f.status === 'error')}
      <button type="button" class="btn btn-secondary btn-sm clear-btn" onclick={clearCompleted}>
        Wyczyść listę
      </button>
    {/if}
  </div>
{/if}

<style>
  .upload-zone {
    border: 2px dashed #d0d0d0;
    border-radius: var(--radius-md);
    padding: var(--spacing-xl);
    text-align: center;
    transition: border-color var(--transition-fast), background var(--transition-fast);
    cursor: pointer;
  }

  .upload-zone.dragover {
    border-color: var(--color-accent);
    background: rgba(46, 49, 146, 0.04);
  }

  .upload-zone:hover {
    border-color: var(--color-accent-secondary);
  }

  .upload-text {
    color: var(--color-text);
    margin-bottom: var(--spacing-sm);
  }

  .upload-hint {
    font-size: var(--font-size-xs);
    color: var(--color-accent-secondary);
    margin-top: var(--spacing-sm);
  }

  .upload-list {
    margin-top: var(--spacing-md);
  }

  .upload-item {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-sm) var(--spacing-md);
    border-bottom: 1px solid #f0f0f0;
    font-size: var(--font-size-sm);
  }

  .upload-item.done {
    background: #e8f5e9;
  }

  .upload-item.error {
    background: #ffebee;
  }

  .upload-filename {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--color-text);
  }

  .upload-size {
    color: var(--color-accent-secondary);
    white-space: nowrap;
  }

  .upload-status {
    white-space: nowrap;
    font-weight: 600;
  }

  .upload-status.uploading {
    color: var(--color-accent);
  }

  .upload-status.success {
    color: var(--color-success);
  }

  .upload-status.error-text {
    color: var(--color-error);
    font-weight: 400;
    font-size: var(--font-size-xs);
  }

  .upload-status.pending {
    color: var(--color-accent-secondary);
  }

  .clear-btn {
    margin-top: var(--spacing-sm);
  }
</style>

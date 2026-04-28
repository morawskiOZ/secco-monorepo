<script lang="ts">
  import { goto } from '$app/navigation';
  import {
    createContent,
    updateContent,
    publishContent,
    draftContent,
    deleteContent,
    type Content,
    type ContentInput
  } from '$lib/api/content';
  import { htmlToMarkdown, markdownToHtml } from '$lib/utils/markdown';
  import Editor from './Editor.svelte';
  import Preview from './Preview.svelte';
  import ImagePicker from './ImagePicker.svelte';

  let {
    contentType,
    existingContent = null,
    showSlug = true,
    showSummary = true,
    showSortOrder = false,
    showDelete = true,
    backUrl = '/',
    defaultTitle = ''
  }: {
    contentType: 'article' | 'project' | 'process';
    existingContent?: Content | null;
    showSlug?: boolean;
    showSummary?: boolean;
    showSortOrder?: boolean;
    showDelete?: boolean;
    backUrl?: string;
    defaultTitle?: string;
  } = $props();

  // Snapshot initial values from props (component is re-created via {#key} when content changes)
  function getInitial(c: Content | null, defTitle: string) {
    return {
      title: c?.title ?? defTitle,
      slug: c?.slug ?? '',
      summary: c?.summary ?? '',
      sortOrder: c?.sort_order ?? 0,
      coverImage: c?.cover_image ?? '',
      status: (c?.status ?? 'draft') as 'draft' | 'published',
      body: c?.body ?? '',
      id: c?.id ?? null,
      html: c ? markdownToHtml(c.body || '') : ''
    };
  }
  const initial = getInitial(existingContent, defaultTitle);

  let title = $state(initial.title);
  let slug = $state(initial.slug);
  let summary = $state(initial.summary);
  let sortOrder = $state(initial.sortOrder);
  let coverImage = $state(initial.coverImage);
  let status = $state<'draft' | 'published'>(initial.status);
  let bodyMarkdown = $state(initial.body);
  let bodyHtml = $state('');
  let showPreview = $state(false);
  let saving = $state(false);
  let error = $state('');
  let saveMessage = $state('');
  let hasUnsaved = $state(false);
  let contentId = $state<number | null>(initial.id);
  let imagePickerOpen = $state(false);
  let coverPickerOpen = $state(false);
  let autoSaveTimer: ReturnType<typeof setTimeout> | undefined;
  let editorComponent: Editor | undefined = $state();

  const initialHtml = initial.html;

  function generateSlug(text: string): string {
    return text
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[łŁ]/g, 'l')
      .replace(/[ąĄ]/g, 'a')
      .replace(/[ćĆ]/g, 'c')
      .replace(/[ęĘ]/g, 'e')
      .replace(/[ńŃ]/g, 'n')
      .replace(/[óÓ]/g, 'o')
      .replace(/[śŚ]/g, 's')
      .replace(/[źŹżŻ]/g, 'z')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')
      .slice(0, 80);
  }

  function handleTitleInput() {
    if (!contentId && showSlug) {
      slug = generateSlug(title);
    }
    markUnsaved();
  }

  function handleEditorChange(html: string) {
    bodyHtml = html;
    bodyMarkdown = htmlToMarkdown(html);
    markUnsaved();
    scheduleAutoSave();
  }

  function markUnsaved() {
    hasUnsaved = true;
  }

  function scheduleAutoSave() {
    if (autoSaveTimer) clearTimeout(autoSaveTimer);
    if (!contentId) return; // Don't auto-save new content
    autoSaveTimer = setTimeout(() => {
      save(false);
    }, 5000);
  }

  function buildInput(): ContentInput {
    const data: ContentInput = {
      type: contentType,
      title,
      body: bodyMarkdown
    };
    if (showSlug) data.slug = slug;
    if (showSummary) data.summary = summary;
    if (showSortOrder) data.sort_order = sortOrder;
    if (coverImage) data.cover_image = coverImage;
    return data;
  }

  async function save(redirect: boolean) {
    if (saving) return;
    saving = true;
    error = '';
    saveMessage = '';

    try {
      const data = buildInput();
      if (contentId) {
        await updateContent(contentId, data);
        saveMessage = 'Zapisano';
        hasUnsaved = false;
      } else {
        const created = await createContent(data);
        contentId = created.id;
        status = created.status;
        hasUnsaved = false;
        saveMessage = 'Utworzono';
        if (redirect) {
          const editUrl = contentType === 'article'
            ? `/tresci/${created.id}`
            : contentType === 'project'
              ? `/projekty/${created.id}`
              : backUrl;
          goto(editUrl);
          return;
        }
      }
      setTimeout(() => { saveMessage = ''; }, 3000);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Błąd zapisu';
    } finally {
      saving = false;
    }
  }

  async function handlePublish() {
    if (!contentId) {
      await save(false);
      if (!contentId) return;
    }
    try {
      const updated = await publishContent(contentId);
      status = updated.status;
      saveMessage = 'Opublikowano';
      hasUnsaved = false;
      setTimeout(() => { saveMessage = ''; }, 3000);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Błąd publikacji';
    }
  }

  async function handleDraft() {
    if (!contentId) return;
    try {
      const updated = await draftContent(contentId);
      status = updated.status;
      saveMessage = 'Cofnięto do szkicu';
      setTimeout(() => { saveMessage = ''; }, 3000);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Błąd zmiany statusu';
    }
  }

  async function handleDelete() {
    if (!contentId) return;
    if (!confirm('Czy na pewno chcesz usunąć tę treść? Ta operacja jest nieodwracalna.')) return;
    try {
      await deleteContent(contentId);
      goto(backUrl);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Błąd usuwania';
    }
  }

  function openImagePicker() {
    imagePickerOpen = true;
  }

  function openCoverPicker() {
    coverPickerOpen = true;
  }

  function handleImageSelect(url: string, alt: string) {
    editorComponent?.insertImage(url, alt);
    imagePickerOpen = false;
  }

  function handleCoverSelect(url: string, _alt: string) {
    coverImage = url;
    coverPickerOpen = false;
    markUnsaved();
  }

  function removeCoverImage() {
    coverImage = '';
    markUnsaved();
  }
</script>

<div class="editor-page">
  <div class="editor-form">
    <div class="form-fields card">
      <div class="form-group">
        <label for="editor-title">Tytuł</label>
        <input
          id="editor-title"
          type="text"
          bind:value={title}
          oninput={handleTitleInput}
          placeholder="Tytuł..."
          class="title-input"
        />
      </div>

      {#if showSlug}
        <div class="form-group">
          <label for="editor-slug">Slug</label>
          <input
            id="editor-slug"
            type="text"
            bind:value={slug}
            oninput={markUnsaved}
            placeholder="slug-artykulu"
            class="slug-input"
          />
        </div>
      {/if}

      {#if showSummary}
        <div class="form-group">
          <label for="editor-summary">Podsumowanie</label>
          <textarea
            id="editor-summary"
            bind:value={summary}
            oninput={markUnsaved}
            placeholder="Krótkie podsumowanie..."
            rows="2"
          ></textarea>
        </div>
      {/if}

      {#if showSortOrder}
        <div class="form-group">
          <label for="editor-sort">Kolejność</label>
          <input
            id="editor-sort"
            type="number"
            bind:value={sortOrder}
            oninput={markUnsaved}
            min="0"
          />
        </div>
      {/if}

      <div class="form-group">
        <span class="field-label">Obraz okładkowy</span>
        <div class="cover-image-field">
          {#if coverImage}
            <div class="cover-preview">
              <img src={coverImage} alt="Okładka" />
              <button type="button" class="btn btn-danger btn-sm" onclick={removeCoverImage}>
                Usuń
              </button>
            </div>
          {/if}
          <button type="button" class="btn btn-secondary btn-sm" onclick={openCoverPicker}>
            {coverImage ? 'Zmień obraz' : 'Wybierz obraz'}
          </button>
        </div>
      </div>

      <div class="form-meta">
        <span class="badge" class:badge-published={status === 'published'} class:badge-draft={status === 'draft'}>
          {status === 'published' ? 'Opublikowany' : 'Szkic'}
        </span>
        {#if hasUnsaved}
          <span class="unsaved-indicator">Niezapisane zmiany</span>
        {/if}
        {#if saveMessage}
          <span class="save-message">{saveMessage}</span>
        {/if}
      </div>
    </div>

    <div class="editor-area">
      <div class="editor-toolbar-row">
        <button
          type="button"
          class="btn btn-sm"
          class:btn-primary={!showPreview}
          class:btn-secondary={showPreview}
          onclick={() => { showPreview = false; }}
        >Edytor</button>
        <button
          type="button"
          class="btn btn-sm"
          class:btn-primary={showPreview}
          class:btn-secondary={!showPreview}
          onclick={() => { showPreview = true; }}
        >Podgląd</button>
      </div>

      <div class="editor-pane" class:hidden={showPreview}>
        <Editor
          bind:this={editorComponent}
          content={initialHtml}
          onchange={handleEditorChange}
          onimageinsert={openImagePicker}
        />
      </div>
      <div class="preview-pane" class:hidden={!showPreview}>
        <Preview content={bodyMarkdown} />
      </div>
    </div>

    {#if error}
      <div class="error-banner">{error}</div>
    {/if}

    <div class="action-buttons">
      <button
        type="button"
        class="btn btn-secondary"
        onclick={() => save(true)}
        disabled={saving}
      >
        {saving ? 'Zapisywanie...' : 'Zapisz szkic'}
      </button>

      {#if status === 'draft'}
        <button type="button" class="btn btn-primary" onclick={handlePublish} disabled={saving}>
          Publikuj
        </button>
      {:else}
        <button type="button" class="btn btn-secondary" onclick={handleDraft} disabled={saving}>
          Cofnij do szkicu
        </button>
      {/if}

      {#if showDelete && contentId}
        <button type="button" class="btn btn-danger" onclick={handleDelete} disabled={saving}>
          Usuń
        </button>
      {/if}
    </div>
  </div>
</div>

<ImagePicker
  open={imagePickerOpen}
  onselect={handleImageSelect}
  onclose={() => { imagePickerOpen = false; }}
/>

<ImagePicker
  open={coverPickerOpen}
  onselect={handleCoverSelect}
  onclose={() => { coverPickerOpen = false; }}
/>

<style>
  .editor-page {
    max-width: 1000px;
  }

  .editor-form {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-md);
  }

  .form-fields {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-md);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
  }

  .form-group label,
  .form-group .field-label {
    font-size: var(--font-size-sm);
    font-weight: 600;
    color: var(--color-text);
  }

  .title-input {
    padding: var(--spacing-sm) var(--spacing-md);
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xl);
    font-weight: 600;
  }

  .title-input:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(46, 49, 146, 0.15);
  }

  .slug-input,
  .form-group textarea,
  .form-group input[type="number"] {
    padding: var(--spacing-sm) var(--spacing-md);
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
  }

  .slug-input:focus,
  .form-group textarea:focus,
  .form-group input[type="number"]:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(46, 49, 146, 0.15);
  }

  .form-group textarea {
    resize: vertical;
  }

  .form-group input[type="number"] {
    max-width: 120px;
  }

  .cover-image-field {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
    align-items: flex-start;
  }

  .cover-preview {
    display: flex;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }

  .cover-preview img {
    width: 160px;
    height: 100px;
    object-fit: cover;
    border-radius: var(--radius-sm);
    border: 1px solid #e0e0e0;
  }

  .form-meta {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    flex-wrap: wrap;
  }

  .unsaved-indicator {
    font-size: var(--font-size-xs);
    color: #e65100;
    font-weight: 600;
  }

  .save-message {
    font-size: var(--font-size-xs);
    color: var(--color-success);
    font-weight: 600;
  }

  .editor-area {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .editor-toolbar-row {
    display: flex;
    gap: var(--spacing-xs);
  }

  .editor-pane,
  .preview-pane {
    display: block;
  }

  .hidden {
    display: none;
  }

  .error-banner {
    padding: var(--spacing-sm) var(--spacing-md);
    background: #ffebee;
    color: var(--color-error);
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
  }

  .action-buttons {
    display: flex;
    gap: var(--spacing-sm);
    flex-wrap: wrap;
  }
</style>

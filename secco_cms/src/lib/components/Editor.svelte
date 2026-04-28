<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Editor } from '@tiptap/core';
  import StarterKit from '@tiptap/starter-kit';
  import Image from '@tiptap/extension-image';
  import Link from '@tiptap/extension-link';
  import Placeholder from '@tiptap/extension-placeholder';

  let {
    content = '',
    onchange,
    onimageinsert
  }: {
    content?: string;
    onchange?: (html: string) => void;
    onimageinsert?: () => void;
  } = $props();

  let element: HTMLDivElement | undefined = $state();
  let editor: Editor | undefined = $state();

  onMount(() => {
    if (!element) return;
    editor = new Editor({
      element,
      extensions: [
        StarterKit,
        Image.configure({ inline: false, allowBase64: false }),
        Link.configure({ openOnClick: false }),
        Placeholder.configure({ placeholder: 'Zacznij pisać...' })
      ],
      content,
      onUpdate: ({ editor: e }) => {
        onchange?.(e.getHTML());
      },
      onTransaction: () => {
        // Force Svelte reactivity update
        editor = editor;
      }
    });
  });

  onDestroy(() => {
    editor?.destroy();
  });

  export function insertImage(url: string, alt: string) {
    editor?.chain().focus().setImage({ src: url, alt }).run();
  }

  function toggleHeading(level: 1 | 2 | 3) {
    editor?.chain().focus().toggleHeading({ level }).run();
  }

  function toggleBold() {
    editor?.chain().focus().toggleBold().run();
  }

  function toggleItalic() {
    editor?.chain().focus().toggleItalic().run();
  }

  function toggleBulletList() {
    editor?.chain().focus().toggleBulletList().run();
  }

  function toggleOrderedList() {
    editor?.chain().focus().toggleOrderedList().run();
  }

  function toggleBlockquote() {
    editor?.chain().focus().toggleBlockquote().run();
  }

  function setLink() {
    if (!editor) return;
    const previousUrl = editor.getAttributes('link').href as string | undefined;
    const url = window.prompt('URL linku:', previousUrl ?? '');
    if (url === null) return;
    if (url === '') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run();
      return;
    }
    editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
  }

  function addHorizontalRule() {
    editor?.chain().focus().setHorizontalRule().run();
  }

  function handleImageClick() {
    onimageinsert?.();
  }
</script>

<div class="editor-wrapper">
  {#if editor}
    <div class="editor-toolbar">
      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('heading', { level: 1 })}
        onclick={() => toggleHeading(1)}
        title="Nagłówek 1"
      >H1</button>
      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('heading', { level: 2 })}
        onclick={() => toggleHeading(2)}
        title="Nagłówek 2"
      >H2</button>
      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('heading', { level: 3 })}
        onclick={() => toggleHeading(3)}
        title="Nagłówek 3"
      >H3</button>

      <span class="toolbar-separator"></span>

      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('bold')}
        onclick={toggleBold}
        title="Pogrubienie"
      ><b>B</b></button>
      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('italic')}
        onclick={toggleItalic}
        title="Kursywa"
      ><i>I</i></button>

      <span class="toolbar-separator"></span>

      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('bulletList')}
        onclick={toggleBulletList}
        title="Lista punktowana"
      >UL</button>
      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('orderedList')}
        onclick={toggleOrderedList}
        title="Lista numerowana"
      >OL</button>

      <span class="toolbar-separator"></span>

      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('blockquote')}
        onclick={toggleBlockquote}
        title="Cytat"
      >Quote</button>
      <button
        type="button"
        class="toolbar-btn"
        class:active={editor.isActive('link')}
        onclick={setLink}
        title="Link"
      >Link</button>
      <button
        type="button"
        class="toolbar-btn"
        onclick={handleImageClick}
        title="Wstaw obraz"
      >Img</button>
      <button
        type="button"
        class="toolbar-btn"
        onclick={addHorizontalRule}
        title="Linia pozioma"
      >HR</button>
    </div>
  {/if}

  <div class="editor-content" bind:this={element}></div>
</div>

<style>
  .editor-wrapper {
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .editor-toolbar {
    position: sticky;
    top: 0;
    z-index: 10;
    display: flex;
    flex-wrap: wrap;
    gap: var(--spacing-xs);
    padding: var(--spacing-sm);
    background: var(--color-white);
    border-bottom: 1px solid #e0e0e0;
  }

  .toolbar-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 32px;
    height: 32px;
    padding: 0 var(--spacing-sm);
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-sm);
    background: var(--color-white);
    color: var(--color-text);
    font-size: var(--font-size-xs);
    font-weight: 600;
    cursor: pointer;
    transition: background var(--transition-fast), color var(--transition-fast);
  }

  .toolbar-btn:hover {
    background: #f0f0f0;
  }

  .toolbar-btn.active {
    background: var(--color-accent);
    color: var(--color-white);
    border-color: var(--color-accent);
  }

  .toolbar-separator {
    width: 1px;
    height: 24px;
    background: #e0e0e0;
    align-self: center;
    margin: 0 var(--spacing-xs);
  }

  .editor-content {
    min-height: 400px;
    padding: var(--spacing-md);
  }

  .editor-content :global(.tiptap) {
    outline: none;
    min-height: 380px;
    line-height: 1.8;
    color: var(--color-text);
  }

  .editor-content :global(.tiptap p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    float: left;
    color: var(--color-accent-secondary);
    pointer-events: none;
    height: 0;
  }

  /* Match ContentRenderer article-body styles */
  .editor-content :global(.tiptap h1) {
    font-size: var(--font-size-2xl);
    font-weight: 700;
    margin-top: var(--spacing-2xl);
    margin-bottom: var(--spacing-md);
    color: var(--color-text);
  }

  .editor-content :global(.tiptap h2) {
    font-size: var(--font-size-xl);
    font-weight: 600;
    margin-top: var(--spacing-2xl);
    margin-bottom: var(--spacing-md);
    color: var(--color-text);
  }

  .editor-content :global(.tiptap h3) {
    font-size: var(--font-size-lg);
    font-weight: 600;
    margin-top: var(--spacing-xl);
    margin-bottom: var(--spacing-sm);
    color: var(--color-text);
  }

  .editor-content :global(.tiptap p) {
    margin-bottom: var(--spacing-lg);
    color: var(--color-text-light);
  }

  .editor-content :global(.tiptap ul),
  .editor-content :global(.tiptap ol) {
    margin-bottom: var(--spacing-lg);
    padding-left: var(--spacing-xl);
    color: var(--color-text-light);
  }

  .editor-content :global(.tiptap li) {
    margin-bottom: var(--spacing-xs);
  }

  .editor-content :global(.tiptap li p) {
    margin-bottom: 0;
  }

  .editor-content :global(.tiptap blockquote) {
    border-left: 3px solid var(--color-accent);
    margin: var(--spacing-lg) 0;
    padding: var(--spacing-md) var(--spacing-xl);
    background-color: rgba(46, 49, 146, 0.04);
    border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  }

  .editor-content :global(.tiptap blockquote p) {
    margin-bottom: 0;
    font-style: italic;
  }

  .editor-content :global(.tiptap code) {
    font-family: monospace;
    font-size: 0.9em;
    background-color: rgba(0, 0, 0, 0.06);
    padding: 2px 6px;
    border-radius: var(--radius-sm);
  }

  .editor-content :global(.tiptap pre) {
    margin-bottom: var(--spacing-lg);
    padding: var(--spacing-md);
    background-color: rgba(0, 0, 0, 0.06);
    border-radius: var(--radius-md);
    overflow-x: auto;
  }

  .editor-content :global(.tiptap pre code) {
    background: none;
    padding: 0;
  }

  .editor-content :global(.tiptap a) {
    color: var(--color-accent);
    text-decoration: underline;
  }

  .editor-content :global(.tiptap img) {
    max-width: 100%;
    height: auto;
    border-radius: var(--radius-md);
    margin: var(--spacing-xl) 0;
  }

  .editor-content :global(.tiptap hr) {
    border: none;
    border-top: 1px solid rgba(0, 0, 0, 0.1);
    margin: var(--spacing-2xl) 0;
  }
</style>

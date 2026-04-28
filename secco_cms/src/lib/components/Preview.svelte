<script lang="ts">
  import { renderMarkdown } from '@secco/render';

  let { content = '' }: { content?: string } = $props();

  const html = $derived(content ? renderMarkdown(content) : '');
</script>

<div class="preview-container">
  <div class="preview-label">Podgląd</div>
  <div class="article-body">
    {#if html}
      {@html html}
    {:else}
      <p class="preview-empty">Brak treści do wyświetlenia</p>
    {/if}
  </div>
</div>

<style>
  .preview-container {
    border: 1px solid #e0e0e0;
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .preview-label {
    padding: var(--spacing-sm) var(--spacing-md);
    background: #f5f5f5;
    border-bottom: 1px solid #e0e0e0;
    font-size: var(--font-size-xs);
    font-weight: 600;
    color: var(--color-accent-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .article-body {
    padding: var(--spacing-md);
    min-height: 400px;
    line-height: 1.8;
    color: var(--color-text);
  }

  .preview-empty {
    color: var(--color-accent-secondary);
    font-style: italic;
  }

  .article-body :global(h1) {
    font-size: var(--font-size-2xl);
    font-weight: 700;
    margin-top: var(--spacing-2xl);
    margin-bottom: var(--spacing-md);
    color: var(--color-text);
  }

  .article-body :global(h2) {
    font-size: var(--font-size-xl);
    font-weight: 600;
    margin-top: var(--spacing-2xl);
    margin-bottom: var(--spacing-md);
    color: var(--color-text);
  }

  .article-body :global(h3) {
    font-size: var(--font-size-lg);
    font-weight: 600;
    margin-top: var(--spacing-xl);
    margin-bottom: var(--spacing-sm);
    color: var(--color-text);
  }

  .article-body :global(h4) {
    font-size: var(--font-size-base);
    font-weight: 600;
    margin-top: var(--spacing-xl);
    margin-bottom: var(--spacing-sm);
    color: var(--color-text);
  }

  .article-body :global(p) {
    margin-bottom: var(--spacing-lg);
    color: var(--color-text-light);
  }

  .article-body :global(ul),
  .article-body :global(ol) {
    margin-bottom: var(--spacing-lg);
    padding-left: var(--spacing-xl);
    color: var(--color-text-light);
  }

  .article-body :global(li) {
    margin-bottom: var(--spacing-xs);
  }

  .article-body :global(li p) {
    margin-bottom: 0;
  }

  .article-body :global(blockquote) {
    border-left: 3px solid var(--color-accent);
    margin: var(--spacing-lg) 0;
    padding: var(--spacing-md) var(--spacing-xl);
    background-color: rgba(46, 49, 146, 0.04);
    border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  }

  .article-body :global(blockquote p) {
    margin-bottom: 0;
    font-style: italic;
  }

  .article-body :global(code) {
    font-family: monospace;
    font-size: 0.9em;
    background-color: rgba(0, 0, 0, 0.06);
    padding: 2px 6px;
    border-radius: var(--radius-sm);
  }

  .article-body :global(pre) {
    margin-bottom: var(--spacing-lg);
    padding: var(--spacing-md);
    background-color: rgba(0, 0, 0, 0.06);
    border-radius: var(--radius-md);
    overflow-x: auto;
  }

  .article-body :global(pre code) {
    background: none;
    padding: 0;
  }

  .article-body :global(a) {
    color: var(--color-accent);
    text-decoration: underline;
  }

  .article-body :global(a:hover) {
    opacity: 0.8;
  }

  .article-body :global(.content-image) {
    margin: var(--spacing-xl) 0;
  }

  .article-body :global(.content-image img) {
    width: 100%;
    height: auto;
    border-radius: var(--radius-md);
  }

  .article-body :global(.content-image figcaption) {
    font-size: var(--font-size-sm);
    color: var(--color-accent-secondary);
    text-align: center;
    margin-top: var(--spacing-sm);
  }

  .article-body :global(hr) {
    border: none;
    border-top: 1px solid rgba(0, 0, 0, 0.1);
    margin: var(--spacing-2xl) 0;
  }

  .article-body :global(img) {
    max-width: 100%;
    height: auto;
    border-radius: var(--radius-md);
  }
</style>

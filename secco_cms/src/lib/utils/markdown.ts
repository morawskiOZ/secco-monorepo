import TurndownService from 'turndown';
import { Marked } from 'marked';

const turndown = new TurndownService({
  headingStyle: 'atx',
  bulletListMarker: '-',
  codeBlockStyle: 'fenced'
});

const marked = new Marked();

/**
 * Convert HTML (from TipTap editor) to Markdown for storage.
 */
export function htmlToMarkdown(html: string): string {
  if (!html || html === '<p></p>') return '';
  return turndown.turndown(html);
}

/**
 * Convert Markdown to HTML for loading into TipTap editor.
 * Uses plain marked (no custom image renderer) so TipTap gets standard HTML.
 */
export function markdownToHtml(markdown: string): string {
  if (!markdown) return '';
  return marked.parse(markdown) as string;
}

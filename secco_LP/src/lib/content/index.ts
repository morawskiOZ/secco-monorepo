// Loads the Markdown files in /content at build time. Pages are prerendered,
// so none of this runs in the browser.

import { CONTENT_ENABLED, SHOW_DRAFTS } from '$lib/flags';
import type { MediaManifest } from './images';
import mediaManifest from './media.gen.json';
import { parseArticle, parsePage, parseProject, sortArticles, sortProjects, visible, type Page } from './model';

export const media: MediaManifest = mediaManifest;

const raw = (files: Record<string, unknown>) =>
	Object.entries(files).map(([path, text]) => ({
		slug: path.split('/').pop()!.replace(/\.md$/, ''),
		text: text as string
	}));

// The flag is a build-time constant: with content off, the files are not bundled at all.
const articleFiles = CONTENT_ENABLED ? import.meta.glob('/content/articles/*.md', { query: '?raw', import: 'default', eager: true }) : {};
const projectFiles = CONTENT_ENABLED ? import.meta.glob('/content/projects/*.md', { query: '?raw', import: 'default', eager: true }) : {};
const pageFiles = CONTENT_ENABLED ? import.meta.glob('/content/pages/*.md', { query: '?raw', import: 'default', eager: true }) : {};

export const articles = sortArticles(visible(raw(articleFiles).map((f) => parseArticle(f.slug, f.text)), SHOW_DRAFTS));
export const projects = sortProjects(visible(raw(projectFiles).map((f) => parseProject(f.slug, f.text)), SHOW_DRAFTS));

const pages = new Map(raw(pageFiles).map((f) => [f.slug, parsePage(f.slug, f.text)]));

export function page(name: string): Page | null {
	const p = pages.get(name);
	return p && (SHOW_DRAFTS || !p.draft) ? p : null;
}

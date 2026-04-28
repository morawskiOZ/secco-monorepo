export interface Content {
  id: number;
  type: 'article' | 'project' | 'process';
  slug: string;
  title: string;
  summary: string;
  body: string;
  cover_image: string;
  status: 'draft' | 'published';
  sort_order: number;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface ContentInput {
  type: 'article' | 'project' | 'process';
  title: string;
  slug?: string;
  summary?: string;
  body?: string;
  cover_image?: string;
  sort_order?: number;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text();
    let message: string;
    try {
      const json = JSON.parse(text);
      message = json.error || json.message || text;
    } catch {
      message = text;
    }
    throw new Error(message);
  }
  return res.json();
}

export async function listContent(type?: string, status?: string): Promise<Content[]> {
  const params = new URLSearchParams();
  if (type) params.set('type', type);
  if (status) params.set('status', status);
  const query = params.toString();
  const url = '/api/content' + (query ? `?${query}` : '');
  const res = await fetch(url, { credentials: 'same-origin' });
  return handleResponse<Content[]>(res);
}

export async function getContent(id: number): Promise<Content> {
  const res = await fetch(`/api/content/${id}`, { credentials: 'same-origin' });
  return handleResponse<Content>(res);
}

export async function createContent(data: ContentInput): Promise<Content> {
  const res = await fetch('/api/content', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    body: JSON.stringify(data)
  });
  return handleResponse<Content>(res);
}

export async function updateContent(id: number, data: ContentInput): Promise<Content> {
  const res = await fetch(`/api/content/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    body: JSON.stringify(data)
  });
  return handleResponse<Content>(res);
}

export async function deleteContent(id: number): Promise<void> {
  const res = await fetch(`/api/content/${id}`, {
    method: 'DELETE',
    credentials: 'same-origin'
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || 'Failed to delete content');
  }
}

export async function publishContent(id: number): Promise<Content> {
  const res = await fetch(`/api/content/${id}/publish`, {
    method: 'PUT',
    credentials: 'same-origin'
  });
  return handleResponse<Content>(res);
}

export async function draftContent(id: number): Promise<Content> {
  const res = await fetch(`/api/content/${id}/draft`, {
    method: 'PUT',
    credentials: 'same-origin'
  });
  return handleResponse<Content>(res);
}

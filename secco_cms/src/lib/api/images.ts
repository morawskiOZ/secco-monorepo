export interface Image {
  id: number;
  r2_key: string;
  filename: string;
  alt_text: string;
  width: number;
  height: number;
  size_bytes: number;
  content_type: string;
  public_url: string;
  created_at: string;
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

export async function listImages(): Promise<Image[]> {
  const res = await fetch('/api/images', { credentials: 'same-origin' });
  return handleResponse<Image[]>(res);
}

export async function deleteImage(id: number): Promise<void> {
  const res = await fetch(`/api/images/${id}`, {
    method: 'DELETE',
    credentials: 'same-origin'
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || 'Failed to delete image');
  }
}

export async function uploadImage(file: File, altText?: string): Promise<Image> {
  const formData = new FormData();
  formData.append('file', file);
  if (altText) {
    formData.append('alt_text', altText);
  }
  const res = await fetch('/api/images/upload', {
    method: 'POST',
    credentials: 'same-origin',
    body: formData
  });
  return handleResponse<Image>(res);
}

export async function updateImageAlt(id: number, altText: string): Promise<Image> {
  const res = await fetch(`/api/images/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    body: JSON.stringify({ alt_text: altText })
  });
  return handleResponse<Image>(res);
}

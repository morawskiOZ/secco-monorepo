# Content API — Reference

## Overview

The Content API is part of the secco_forms Go backend. It provides public read endpoints for the website and protected CRUD endpoints for the CMS.

## Authentication

Protected endpoints require the `X-API-Key` header:
```
X-API-Key: <CONTENT_API_KEY env var value>
```

Preview endpoints require a `token` query parameter:
```
?token=<PREVIEW_TOKEN env var value>
```

## Public Endpoints (Website)

### List Articles
```
GET /api/content/tresci
```

Query parameters:
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `limit` | int | 12 | Items per page |

Response `200`:
```json
{
  "items": [
    {
      "id": 1,
      "type": "article",
      "slug": "jak-wybrac-styl",
      "title": "Jak wybrać styl wnętrzarski?",
      "summary": "Krótki przewodnik...",
      "cover_image": "https://imagedelivery.net/hash/id/w=400,h=300,fit=cover,quality=85",
      "published_at": "2026-03-19T10:00:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "limit": 12
}
```

Only returns `status=published` items, ordered by `published_at DESC`.

---

### Get Article by Slug
```
GET /api/content/tresci/:slug
```

Response `200`:
```json
{
  "id": 1,
  "type": "article",
  "slug": "jak-wybrac-styl",
  "title": "Jak wybrać styl wnętrzarski?",
  "summary": "Krótki przewodnik...",
  "body": "# Jak wybrać styl\n\nMarkdown content...",
  "cover_image": "https://imagedelivery.net/hash/id/large",
  "published_at": "2026-03-19T10:00:00Z",
  "created_at": "2026-03-18T15:00:00Z",
  "updated_at": "2026-03-19T09:30:00Z"
}
```

Response `404`:
```json
{ "error": "not found" }
```

---

### List Projects
```
GET /api/content/projekty
```

Same query params as articles. Returns `status=published` items ordered by `sort_order ASC`.

---

### Get Project by Slug
```
GET /api/content/projekty/:slug
```

Same format as article response.

---

### Get Process Page
```
GET /api/content/proces
```

Returns the single published `process` type entry. Response `200` same format. Response `404` if no published process entry exists.

---

### Preview Draft Content
```
GET /api/content/preview/:type/:slug?token=PREVIEW_TOKEN
```

- `:type` — `article`, `project`, or `process`
- Returns content regardless of `status` (drafts visible)
- Requires valid `token` query parameter
- Response format same as public get endpoints

Response `403`:
```json
{ "error": "invalid preview token" }
```

---

## Protected Endpoints (CMS)

All require `X-API-Key` header.

### List All Content
```
GET /api/admin/content
```

Query parameters:
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | (all) | Filter by `article`, `project`, or `process` |
| `status` | string | (all) | Filter by `draft` or `published` |
| `page` | int | 1 | Page number |
| `limit` | int | 50 | Items per page |

Returns all content including drafts.

---

### Get Content by ID
```
GET /api/admin/content/:id
```

Returns full content including body, regardless of status.

---

### Create Content
```
POST /api/admin/content
Content-Type: application/json
```

Request body:
```json
{
  "type": "article",
  "slug": "nowy-artykul",
  "title": "Nowy artykuł",
  "summary": "Opis...",
  "body": "# Treść\n\nMarkdown...",
  "cover_image": "https://imagedelivery.net/hash/id/large",
  "sort_order": 0
}
```

Validation:
- `type` required, must be `article`, `project`, or `process`
- `slug` required, must be unique, lowercase alphanumeric + hyphens
- `title` required
- `body` required
- For `process` type: only one entry allowed

Response `201`:
```json
{
  "id": 5,
  "type": "article",
  "slug": "nowy-artykul",
  "status": "draft",
  "created_at": "2026-03-19T12:00:00Z"
}
```

Response `400`:
```json
{ "error": "slug already exists" }
```

---

### Update Content
```
PUT /api/admin/content/:id
Content-Type: application/json
```

Request body (all fields optional, only provided fields are updated):
```json
{
  "title": "Updated title",
  "body": "Updated markdown..."
}
```

Response `200`:
```json
{
  "id": 5,
  "updated_at": "2026-03-19T13:00:00Z"
}
```

---

### Delete Content
```
DELETE /api/admin/content/:id
```

Response `204` (no body).

Cascade: also removes entries in `content_images` junction table.

---

### Publish Content
```
PUT /api/admin/content/:id/publish
```

Sets `status=published` and `published_at=NOW()` (if not already set).

Response `200`:
```json
{
  "id": 5,
  "status": "published",
  "published_at": "2026-03-19T14:00:00Z"
}
```

---

### Unpublish Content
```
PUT /api/admin/content/:id/draft
```

Sets `status=draft`. Does not clear `published_at`.

Response `200`:
```json
{
  "id": 5,
  "status": "draft"
}
```

---

## Image Endpoints (Protected)

### Upload Image
```
POST /api/admin/images/upload
Content-Type: multipart/form-data
X-API-Key: <key>
```

Form fields:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | file | yes | Image file (JPEG, PNG, WebP) |
| `alt_text` | string | no | Alt text for accessibility |

Validation:
- Max file size: 20MB
- Allowed types: `image/jpeg`, `image/png`, `image/webp`

Response `201`:
```json
{
  "id": 1,
  "cloudflare_id": "abc123-def456",
  "filename": "salon.jpg",
  "alt_text": "Nowoczesny salon w stylu skandynawskim",
  "width": 4000,
  "height": 3000,
  "size_bytes": 5242880,
  "delivery_url": "https://imagedelivery.net/<hash>/abc123-def456",
  "created_at": "2026-03-19T10:00:00Z"
}
```

The `delivery_url` is the base URL. Append a variant string for transformed images:
```
{delivery_url}/w=800,quality=90
```

---

### List Images
```
GET /api/admin/images
X-API-Key: <key>
```

Query parameters:
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `limit` | int | 50 | Items per page |
| `search` | string | | Filter by filename |

Response `200`:
```json
{
  "items": [
    {
      "id": 1,
      "cloudflare_id": "abc123",
      "filename": "salon.jpg",
      "alt_text": "Salon",
      "width": 4000,
      "height": 3000,
      "size_bytes": 5242880,
      "delivery_url": "https://imagedelivery.net/hash/abc123",
      "created_at": "2026-03-19T10:00:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 50
}
```

---

### Delete Image
```
DELETE /api/admin/images/:id
X-API-Key: <key>
```

Deletes the image from both Cloudflare and the local database. Also removes `content_images` junction entries.

Response `204` (no body).

---

## Error Responses

All errors follow this format:
```json
{
  "error": "description of the error"
}
```

| Status | Meaning |
|--------|---------|
| 400 | Bad request (validation error, missing fields) |
| 401 | Missing or invalid API key |
| 403 | Invalid preview token |
| 404 | Content or image not found |
| 409 | Conflict (duplicate slug) |
| 413 | File too large |
| 415 | Unsupported media type |
| 429 | Rate limited |
| 500 | Internal server error |

## SEO Meta Tag Injection

The Go backend intercepts requests to content pages (`/tresci/*`, `/projekty/*`, `/proces-projektowy`) and injects meta tags into the HTML before serving.

For a request to `/tresci/jak-wybrac-styl`:
1. Go looks up the article by slug
2. Reads the `200.html` (SPA fallback) template
3. Replaces the `<head>` section with:
   ```html
   <title>Jak wybrać styl wnętrzarski? — Secco Studio</title>
   <meta name="description" content="Krótki przewodnik...">
   <meta property="og:title" content="Jak wybrać styl wnętrzarski?">
   <meta property="og:description" content="Krótki przewodnik...">
   <meta property="og:image" content="https://imagedelivery.net/hash/id/w=1200,h=630,fit=cover,quality=85">
   <meta property="og:url" content="https://seccostudio.com/tresci/jak-wybrac-styl">
   <meta property="og:type" content="article">
   ```
4. Serves the modified HTML
5. SvelteKit hydrates and renders the full content client-side

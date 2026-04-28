# Cloudflare Images — Setup & Integration Guide

## Overview

Cloudflare Images handles image storage, optimization, and CDN delivery for secco_forms and secco_cms. Images uploaded through the CMS are stored on Cloudflare's infrastructure and served via their global CDN with automatic format conversion (WebP/AVIF).

## Prerequisites

- Cloudflare account (free plan is sufficient for setup, Images is a paid add-on)
- Cloudflare Images subscription ($5/month for 100K images stored + $1/100K images served)

## Step 1: Enable Cloudflare Images

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Navigate to **Images** in the left sidebar
3. Click **Subscribe** if not already subscribed
4. Note your **Account ID** (visible in the dashboard URL or in the right sidebar of any zone)
5. Note your **Account Hash** (visible in **Images > Overview** under "Delivery URL")
   - Format: `https://imagedelivery.net/<ACCOUNT_HASH>/...`

## Step 2: Create API Token

1. Go to **My Profile > API Tokens** (https://dash.cloudflare.com/profile/api-tokens)
2. Click **Create Token**
3. Use **Custom Token** template:
   - **Token name:** `secco-cms-images`
   - **Permissions:**
     - Account > Cloudflare Images > Edit
   - **Account Resources:**
     - Include > Your account
4. Click **Continue to summary** > **Create Token**
5. **Copy the token immediately** — it won't be shown again

## Step 3: Enable Flexible Variants

Flexible variants allow on-the-fly image transformations via URL parameters (no pre-configured variants needed).

1. Go to **Images > Variants**
2. Toggle **Flexible variants** to ON
3. This enables URL-based transformations like:
   ```
   https://imagedelivery.net/<hash>/<id>/w=800,quality=90
   ```

## Step 4: Configure Environment Variables

### secco_forms (.env)
```env
CLOUDFLARE_ACCOUNT_ID=<your-account-id>
CLOUDFLARE_IMAGES_API_TOKEN=<your-api-token>
CLOUDFLARE_ACCOUNT_HASH=<your-account-hash>
```

### secco_cms (.env)
```env
CLOUDFLARE_ACCOUNT_ID=<same-account-id>
CLOUDFLARE_IMAGES_API_TOKEN=<same-api-token>
CLOUDFLARE_ACCOUNT_HASH=<same-account-hash>
```

## API Usage

### Upload an Image

```bash
curl -X POST "https://api.cloudflare.com/client/v4/accounts/{account_id}/images/v1" \
  -H "Authorization: Bearer {api_token}" \
  -F "file=@/path/to/image.jpg" \
  -F "metadata={\"key\":\"value\"}"
```

Response:
```json
{
  "result": {
    "id": "abc123-def456",
    "filename": "image.jpg",
    "uploaded": "2026-03-19T10:00:00Z",
    "variants": [
      "https://imagedelivery.net/<hash>/abc123-def456/public"
    ]
  },
  "success": true
}
```

### Delivery URL Format

```
https://imagedelivery.net/<ACCOUNT_HASH>/<IMAGE_ID>/<VARIANT>
```

### Flexible Variant Parameters

| Parameter | Values | Description |
|-----------|--------|-------------|
| `w` | 1-10000 | Width in pixels |
| `h` | 1-10000 | Height in pixels |
| `fit` | `scale-down`, `contain`, `cover`, `crop`, `pad` | Resize behavior |
| `quality` | 1-100 | JPEG/WebP quality |
| `format` | `auto`, `webp`, `avif`, `json` | Output format (`auto` uses Accept header) |
| `gravity` | `auto`, `center`, `top`, `bottom`, `left`, `right` | Crop anchor point |

### Recommended Variants for Secco Studio

| Use Case | Variant String | Notes |
|----------|---------------|-------|
| Thumbnail (cards) | `w=400,h=300,fit=cover,quality=85` | Blog/project listing cards |
| Medium (in-content) | `w=800,quality=90` | Images within article body |
| Large (gallery) | `w=1200,quality=90` | Project gallery tiles |
| Hero (cover) | `w=1920,quality=95` | Full-width cover images |
| Full (lightbox) | `w=2400,quality=95` | Full-size viewing in lightbox |
| OG (social) | `w=1200,h=630,fit=cover,quality=85` | OpenGraph/social sharing |

### Delete an Image

```bash
curl -X DELETE "https://api.cloudflare.com/client/v4/accounts/{account_id}/images/v1/{image_id}" \
  -H "Authorization: Bearer {api_token}"
```

## Image Quality Guidelines

For an interior design portfolio, image quality is paramount:

1. **Always upload original files** — CF stores the original and generates variants on-the-fly
2. **Use quality=90+ for portfolio images** — the quality difference between 85 and 95 is visible on high-resolution displays
3. **Use quality=85 for thumbnails** — at small sizes, quality differences are imperceptible
4. **Prefer JPEG originals** for photographs — smaller upload size than PNG, better for photos
5. **Let CF handle format conversion** — `format=auto` serves WebP/AVIF to supporting browsers
6. **Specify dimensions** to prevent layout shift (CLS) — the image upload response includes width/height

## Rate Limits

- **Upload:** 100 images per 10 minutes per account
- **Delivery:** No rate limit (CDN)
- **API calls:** 1200 requests per 5 minutes per account

## Cost Estimation

For a small interior design portfolio:
- **Storage:** 500 images × ~5MB avg = ~2.5GB → well within 100K image limit ($5/month)
- **Delivery:** ~10K page views/month × ~5 images/page = ~50K deliveries → within $1 tier
- **Estimated monthly cost:** $6-10/month

## Troubleshooting

| Issue | Solution |
|-------|----------|
| 403 on upload | Check API token has `Images:Edit` permission |
| Variant not working | Ensure Flexible Variants is enabled in dashboard |
| Blurry images | Increase quality parameter (use 90+) |
| Slow first load | Normal — CF generates variants on first request, cached after |
| CORS errors | CF Images handles CORS automatically; check if your domain is set up |

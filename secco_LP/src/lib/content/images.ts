// Responsive image variants. The build script (scripts/prebuild.ts) writes the
// files and a manifest; pages turn manifest entries into srcset attributes.

export const VARIANT_WIDTHS = [480, 800, 1200, 1920];

export interface Variant {
	width: number;
	src: string;
}

export interface MediaInfo {
	width: number;
	height: number;
	variants: Variant[];
}

/** Keyed by the public path used in content, e.g. "/media/salon.jpg". */
export type MediaManifest = Record<string, MediaInfo>;

/** Standard widths below the original, plus the original width itself (never upscaled). */
export function variantWidths(originalWidth: number): number[] {
	const max = VARIANT_WIDTHS[VARIANT_WIDTHS.length - 1];
	const smaller = VARIANT_WIDTHS.filter((w) => w < originalWidth);
	return originalWidth <= max ? [...smaller, originalWidth] : [...smaller];
}

/** "/media/a/salon.jpg", 800 -> "/media/a/salon-800.webp" */
export function variantPath(src: string, width: number): string {
	return src.replace(/\.[^./]+$/, '') + `-${width}.webp`;
}

export interface ImgAttrs {
	src: string;
	srcset: string;
	width: number;
	height: number;
}

/** Attributes for an <img>; `fallbackWidth` picks the plain `src` for browsers without srcset. */
export function imgAttrs(info: MediaInfo, fallbackWidth = 1200): ImgAttrs {
	const v = info.variants;
	const fallback = v.find((x) => x.width >= fallbackWidth) ?? v[v.length - 1];
	return {
		src: fallback.src,
		srcset: v.map((x) => `${x.src} ${x.width}w`).join(', '),
		width: info.width,
		height: info.height
	};
}

export function largest(info: MediaInfo): Variant {
	return info.variants[info.variants.length - 1];
}

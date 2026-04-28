export interface Article {
	slug: string;
	title: string;
	summary: string;
	body: string;
	cover_image: string;
	tags: string[];
	published_at: string;
}

export interface Project {
	slug: string;
	title: string;
	summary: string;
	body: string;
	cover_image: string;
	sort_order: number;
	published_at: string;
}

export interface ProcessPage {
	title: string;
	body: string;
}

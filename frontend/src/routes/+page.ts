import type { PageLoad } from './$types';

interface Tag {
	id: string;
	tag: string;
	description: string | null;
	viewCount: number;
	flags: number;
	fanslyCreatedAt: Date;
	createdAt: Date;
	updatedAt: Date;
}

interface TagsResponse {
	tags: Tag[];
	pagination: {
		page: number;
		limit: number;
		totalCount: number;
		totalPages: number;
	};
}

export const load: PageLoad = async ({ fetch, url }) => {
	const page = url.searchParams.get('page') || '1';
	const search = url.searchParams.get('search') || '';
	const sortBy = url.searchParams.get('sortBy') || 'viewCount';
	const sortOrder = url.searchParams.get('sortOrder') || 'desc';

	try {
		const params = new URLSearchParams({
			page,
			limit: '20',
			search,
			sortBy,
			sortOrder
		});

		const response = await fetch(`http://localhost:3000/api/tags?${params}`);

		if (!response.ok) {
			throw new Error('Failed to fetch tags');
		}

		const data: TagsResponse = await response.json();

		return {
			tags: data.tags,
			pagination: data.pagination,
			search,
			sortBy,
			sortOrder
		};
	} catch (error) {
		console.error('Error loading tags:', error);
		return {
			tags: [],
			pagination: {
				page: 1,
				limit: 20,
				totalCount: 0,
				totalPages: 0
			},
			search,
			sortBy,
			sortOrder
		};
	}
};

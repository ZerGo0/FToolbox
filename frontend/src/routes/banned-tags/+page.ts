import { PUBLIC_API_URL } from '$env/static/public';
import type { PageLoad } from './$types';

interface Tag {
  id: string;
  tag: string;
  viewCount: number;
  postCount: number;
  heat: number;
  fanslyCreatedAt: Date;
  lastCheckedAt: Date | null;
  createdAt: Date;
  updatedAt: Date;
  rank?: number | null;
  isDeleted?: boolean;
  deletedDetectedAt?: number | null;
}

interface BannedTagsResponse {
  tags: Tag[];
  pagination: {
    page: number;
    limit: number;
    totalCount: number;
    totalPages: number;
  };
  statistics: {
    totalBanned: number;
    bannedLast24h: number;
    bannedLast7d: number;
    bannedLast30d: number;
  };
}

export const load: PageLoad = async ({ fetch, url }) => {
  const page = url.searchParams.get('page') || '1';
  const search = url.searchParams.get('search') || '';
  const sortBy = url.searchParams.get('sortBy') || 'deletedDetectedAt';
  const sortOrder = url.searchParams.get('sortOrder') || 'desc';

  try {
    const params = new URLSearchParams({
      page,
      limit: '20',
      search,
      sortBy,
      sortOrder
    });

    // Fetch banned tags data
    const response = await fetch(`${PUBLIC_API_URL}/api/tags/banned?${params}`);

    if (!response.ok) {
      throw new Error('Failed to fetch banned tags');
    }

    const data: BannedTagsResponse = await response.json();

    return {
      tags: data.tags,
      pagination: data.pagination,
      statistics: data.statistics,
      search,
      sortBy,
      sortOrder
    };
  } catch (error) {
    console.error('Error loading banned tags:', error);
    return {
      tags: [],
      pagination: {
        page: 1,
        limit: 20,
        totalCount: 0,
        totalPages: 0
      },
      statistics: {
        totalBanned: 0,
        bannedLast24h: 0,
        bannedLast7d: 0,
        bannedLast30d: 0
      },
      search,
      sortBy,
      sortOrder
    };
  }
};

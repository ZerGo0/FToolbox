import { PUBLIC_API_URL } from '$env/static/public';
import type { PageLoad } from './$types';

interface TagHistory {
  id: number;
  tagId: string;
  viewCount: number;
  change: number;
  changePercent: number;
  createdAt: Date;
  updatedAt: Date;
}

interface Tag {
  id: string;
  tag: string;
  viewCount: number;
  fanslyCreatedAt: Date;
  lastCheckedAt: Date | null;
  createdAt: Date;
  updatedAt: Date;
  rank?: number | null;
  history?: TagHistory[];
  isDeleted?: boolean;
  deletedDetectedAt?: number | null;
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
  console.log('hello world');
  const page = url.searchParams.get('page') || '1';
  const search = url.searchParams.get('search') || '';
  const sortBy = url.searchParams.get('sortBy') || 'rank';
  const sortOrder = url.searchParams.get('sortOrder') || 'asc';
  const includeHistory = url.searchParams.get('includeHistory') || 'true';

  // Default to last 7 days if no dates provided
  const now = new Date();
  const sevenDaysAgo = new Date();
  sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);

  // Set end date to end of day
  const endOfDay = new Date(now);
  endOfDay.setHours(23, 59, 59, 999);

  const historyStartDate = url.searchParams.get('historyStartDate') || sevenDaysAgo.toISOString();
  const historyEndDate = url.searchParams.get('historyEndDate') || endOfDay.toISOString();

  try {
    const params = new URLSearchParams({
      page,
      limit: '20',
      search,
      sortBy,
      sortOrder,
      includeHistory,
      historyStartDate,
      historyEndDate
    });

    const response = await fetch(`${PUBLIC_API_URL}/api/tags?${params}`);

    if (!response.ok) {
      throw new Error('Failed to fetch tags');
    }

    const data: TagsResponse = await response.json();

    return {
      tags: data.tags,
      pagination: data.pagination,
      search,
      sortBy,
      sortOrder,
      includeHistory: includeHistory === 'true',
      historyStartDate,
      historyEndDate
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
      sortOrder,
      includeHistory: includeHistory === 'true',
      historyStartDate,
      historyEndDate
    };
  }
};

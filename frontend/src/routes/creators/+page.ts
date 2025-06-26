import { PUBLIC_API_URL } from '$env/static/public';
import type { PageLoad } from './$types';

interface CreatorHistory {
  id: number;
  creatorId: string;
  mediaLikes: number;
  postLikes: number;
  followers: number;
  imageCount: number;
  videoCount: number;
  mediaLikesChange: number;
  postLikesChange: number;
  followersChange: number;
  createdAt: number;
  updatedAt: number;
}

interface Creator {
  id: string;
  username: string;
  displayName?: string | null;
  mediaLikes: number;
  postLikes: number;
  followers: number;
  imageCount: number;
  videoCount: number;
  rank?: number | null;
  lastCheckedAt: number | null;
  isDeleted?: boolean;
  deletedDetectedAt?: number | null;
  createdAt: number;
  updatedAt: number;
  history?: CreatorHistory[];
}

interface CreatorsResponse {
  creators: Creator[];
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
  const sortBy = url.searchParams.get('sortBy') || 'followers';
  const sortOrder = url.searchParams.get('sortOrder') || 'desc';
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

    const response = await fetch(`${PUBLIC_API_URL}/api/creators?${params}`);

    if (!response.ok) {
      throw new Error('Failed to fetch creators');
    }

    const data: CreatorsResponse = await response.json();

    return {
      creators: data.creators,
      pagination: data.pagination,
      search,
      sortBy,
      sortOrder,
      includeHistory: includeHistory === 'true',
      historyStartDate,
      historyEndDate
    };
  } catch (error) {
    console.error('Error loading creators:', error);
    return {
      creators: [],
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

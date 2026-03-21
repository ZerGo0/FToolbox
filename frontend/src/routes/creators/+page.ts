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

interface CreatorStatistics {
  totalFollowers: number;
  followersChange24h: number;
  followersChangePercent24h: number;
  totalMediaLikes: number;
  mediaLikesChange24h: number;
  mediaLikesChangePercent24h: number;
  totalPostLikes: number;
  postLikesChange24h: number;
  postLikesChangePercent24h: number;
  calculatedAt: number | null;
}

function createDefaultCreatorStatistics(): CreatorStatistics {
  return {
    totalFollowers: 0,
    followersChange24h: 0,
    followersChangePercent24h: 0,
    totalMediaLikes: 0,
    mediaLikesChange24h: 0,
    mediaLikesChangePercent24h: 0,
    totalPostLikes: 0,
    postLikesChange24h: 0,
    postLikesChangePercent24h: 0,
    calculatedAt: null
  };
}

function createEmptyPagination() {
  return {
    page: 1,
    limit: 20,
    totalCount: 0,
    totalPages: 0
  };
}

async function fetchCreatorStatistics(fetchFn: typeof fetch): Promise<CreatorStatistics> {
  try {
    const response = await fetchFn(`${PUBLIC_API_URL}/api/creators/statistics`);
    if (!response.ok) {
      return createDefaultCreatorStatistics();
    }

    return response.json();
  } catch (statsError) {
    console.error('Error loading creator statistics:', statsError);
    return createDefaultCreatorStatistics();
  }
}

export const load: PageLoad = async ({ fetch, url }) => {
  const page = url.searchParams.get('page') || '1';
  const search = url.searchParams.get('search') || '';
  const sortBy = 'rank';
  const sortOrderParam = url.searchParams.get('sortOrder') || 'asc';
  const sortOrder = sortOrderParam === 'desc' ? 'desc' : 'asc';
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

    const [response, statistics] = await Promise.all([
      fetch(`${PUBLIC_API_URL}/api/creators?${params}`),
      fetchCreatorStatistics(fetch)
    ]);

    if (!response.ok) {
      throw new Error('Failed to fetch creators');
    }

    const data: CreatorsResponse = await response.json();

    return {
      creators: data.creators,
      pagination: data.pagination,
      statistics,
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
      pagination: createEmptyPagination(),
      statistics: createDefaultCreatorStatistics(),
      search,
      sortBy,
      sortOrder,
      includeHistory: includeHistory === 'true',
      historyStartDate,
      historyEndDate
    };
  }
};

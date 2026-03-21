import { PUBLIC_API_URL } from '$env/static/public';
import type { PageLoad } from './$types';

interface TagHistory {
  id: number;
  tagId: string;
  viewCount: number;
  change: number;
  changePercent: number;
  postCount: number;
  ratio: number;
  postCountChange: number;
  createdAt: Date;
  updatedAt: Date;
}

interface Tag {
  id: string;
  tag: string;
  viewCount: number;
  postCount: number;
  ratio: number;
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

interface TagStatistics {
  totalViewCount: number;
  totalPostCount: number;
  change24h: number;
  changePercent24h: number;
  postChange24h: number;
  postChangePercent24h: number;
  calculatedAt: number | null;
}

interface TagLoadPayload {
  tags: Tag[];
  pagination: TagsResponse['pagination'];
  statistics: TagStatistics;
}

function createDefaultTagStatistics(): TagStatistics {
  return {
    totalViewCount: 0,
    totalPostCount: 0,
    change24h: 0,
    changePercent24h: 0,
    postChange24h: 0,
    postChangePercent24h: 0,
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

function getDefaultHistoryRange(url: URL) {
  const now = new Date();
  const sevenDaysAgo = new Date();
  sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);

  const endOfDay = new Date(now);
  endOfDay.setHours(23, 59, 59, 999);

  return {
    historyStartDate: url.searchParams.get('historyStartDate') || sevenDaysAgo.toISOString(),
    historyEndDate: url.searchParams.get('historyEndDate') || endOfDay.toISOString()
  };
}

function createLoadResult(input: {
  tags: Tag[];
  pagination: TagsResponse['pagination'];
  statistics: TagStatistics;
  search: string;
  sortBy: 'rank' | 'ratio';
  sortOrder: 'asc' | 'desc';
  includeHistory: string;
  historyStartDate: string;
  historyEndDate: string;
}) {
  return {
    tags: input.tags,
    pagination: input.pagination,
    statistics: input.statistics,
    search: input.search,
    sortBy: input.sortBy,
    sortOrder: input.sortOrder,
    includeHistory: input.includeHistory === 'true',
    historyStartDate: input.historyStartDate,
    historyEndDate: input.historyEndDate
  };
}

async function fetchTagStatistics(fetchFn: typeof fetch): Promise<TagStatistics> {
  try {
    const response = await fetchFn(`${PUBLIC_API_URL}/api/tags/statistics`);
    if (!response.ok) {
      return createDefaultTagStatistics();
    }

    return response.json();
  } catch (statsError) {
    console.error('Error loading tag statistics:', statsError);
    return createDefaultTagStatistics();
  }
}

async function fetchTagLoadPayload(
  fetchFn: typeof fetch,
  params: URLSearchParams
): Promise<TagLoadPayload> {
  const [tagsResponse, statistics] = await Promise.all([
    fetchFn(`${PUBLIC_API_URL}/api/tags?${params}`),
    fetchTagStatistics(fetchFn)
  ]);

  if (!tagsResponse.ok) {
    throw new Error('Failed to fetch tags');
  }

  const data: TagsResponse = await tagsResponse.json();
  return {
    tags: data.tags,
    pagination: data.pagination,
    statistics
  };
}

export const load: PageLoad = async ({ fetch, url }) => {
  const page = url.searchParams.get('page') || '1';
  const search = url.searchParams.get('search') || '';
  const tags = url.searchParams.get('tags') || '';
  const sortByParam = url.searchParams.get('sortBy') || 'rank';
  const sortBy = sortByParam === 'ratio' ? 'ratio' : 'rank';
  const sortOrderParam = url.searchParams.get('sortOrder') || 'asc';
  const sortOrder = sortOrderParam === 'desc' ? 'desc' : 'asc';
  const includeHistory = url.searchParams.get('includeHistory') || 'true';
  const { historyStartDate, historyEndDate } = getDefaultHistoryRange(url);

  try {
    const params = new URLSearchParams({
      page,
      limit: '20',
      search,
      tags,
      sortBy,
      sortOrder,
      includeHistory,
      historyStartDate,
      historyEndDate
    });

    const payload = await fetchTagLoadPayload(fetch, params);

    return createLoadResult({
      tags: payload.tags,
      pagination: payload.pagination,
      statistics: payload.statistics,
      search,
      sortBy,
      sortOrder,
      includeHistory,
      historyStartDate,
      historyEndDate
    });
  } catch (error) {
    console.error('Error loading tags:', error);
    return createLoadResult({
      tags: [],
      pagination: createEmptyPagination(),
      statistics: createDefaultTagStatistics(),
      search,
      sortBy,
      sortOrder,
      includeHistory,
      historyStartDate,
      historyEndDate
    });
  }
};

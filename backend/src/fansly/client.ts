import { Logger } from '../utils/logger';

// Generic API response type
interface FanslyResponse<T> {
  success: boolean;
  response?: T;
}

// Tag interfaces
interface FanslyTag {
  id: string;
  tag: string;
  description: string;
  viewCount: number;
  flags: number;
  createdAt: number;
}

interface TagResponseData {
  mediaOfferSuggestionTag?: FanslyTag;
  aggregationData?: Record<string, unknown>;
}

// Post interfaces
interface FanslyAttachment {
  postId: string;
  pos: number;
  contentType: number;
  contentId: string;
}

interface FanslyPost {
  id: string;
  accountId: string;
  content: string;
  createdAt: number;
  attachments: FanslyAttachment[];
  // Extended fields from the API
  fypFlags?: number;
  inReplyTo?: string | null;
  inReplyToRoot?: string | null;
  replyPermissionFlags?: unknown | null;
  expiresAt?: number | null;
  likeCount?: number;
  replyCount?: number;
  wallIds?: string[];
  mediaLikeCount?: number;
  totalTipAmount?: number;
  attachmentTipAmount?: number;
  accountMentions?: Array<{
    start: number;
    end: number;
    handle: string;
    accountId: string;
  }>;
  postReplyPermissionFlags?: Array<{
    id: string;
    postId: string;
    type: number;
    flags: number;
    metadata: string;
  }>;
  tipAmount?: number;
}

interface PostsResponseData {
  mediaOfferSuggestions?: unknown[];
  aggregationData?: {
    accounts?: unknown[];
    accountMedia?: unknown[];
    accountMediaBundles?: unknown[];
    posts?: FanslyPost[];
    tips?: unknown[];
    tipGoals?: unknown[];
    stories?: unknown[];
  };
}

export class FanslyClient {
  private logger: Logger;
  private baseUrl = 'https://apiv3.fansly.com/api/v1';

  constructor() {
    this.logger = new Logger('FanslyClient');
  }

  async getTag(tagName: string): Promise<FanslyTag | null> {
    try {
      const url = `${this.baseUrl}/contentdiscovery/media/tag?tag=${encodeURIComponent(tagName)}&ngsw-bypass=true`;
      const response = await fetch(url);

      if (!response.ok) {
        this.logger.error(`Failed to fetch tag: ${tagName}, status: ${response.status}`);
        return null;
      }

      const data = (await response.json()) as FanslyResponse<TagResponseData>;

      if (data.success && data.response?.mediaOfferSuggestionTag) {
        return data.response.mediaOfferSuggestionTag;
      }

      return null;
    } catch (error) {
      this.logger.error(`Error fetching tag ${tagName}:`, error);
      return null;
    }
  }

  // Method for getting full tag response (used by routes/tags.ts)
  async getTagResponse(tagName: string): Promise<FanslyResponse<TagResponseData> | null> {
    try {
      const url = `${this.baseUrl}/contentdiscovery/media/tag?tag=${encodeURIComponent(tagName)}&ngsw-bypass=true`;
      const response = await fetch(url);

      if (!response.ok) {
        this.logger.error(`Failed to fetch tag: ${tagName}, status: ${response.status}`);
        return null;
      }

      return (await response.json()) as FanslyResponse<TagResponseData>;
    } catch (error) {
      this.logger.error(`Error fetching tag ${tagName}:`, error);
      return null;
    }
  }

  async getPostsForTag(
    tagId: string,
    limit: number = 25,
    offset: number = 0
  ): Promise<FanslyPost[]> {
    try {
      const url = `${this.baseUrl}/contentdiscovery/media/suggestionsnew?before=0&after=0&tagIds=${tagId}&limit=${limit}&offset=${offset}&ngsw-bypass=true`;
      const response = await fetch(url);

      if (!response.ok) {
        this.logger.error(`Failed to fetch posts for tag: ${tagId}, status: ${response.status}`);
        return [];
      }

      const data = (await response.json()) as FanslyResponse<PostsResponseData>;

      if (data.success && data.response?.aggregationData?.posts) {
        return data.response.aggregationData.posts;
      }

      return [];
    } catch (error) {
      this.logger.error(`Error fetching posts for tag ${tagId}:`, error);
      return [];
    }
  }

  // Method for getting full posts response
  async getPostsResponse(
    tagId: string,
    limit: number = 25,
    offset: number = 0
  ): Promise<FanslyResponse<PostsResponseData> | null> {
    try {
      const url = `${this.baseUrl}/contentdiscovery/media/suggestionsnew?before=0&after=0&tagIds=${tagId}&limit=${limit}&offset=${offset}&ngsw-bypass=true`;
      const response = await fetch(url);

      if (!response.ok) {
        this.logger.error(`Failed to fetch posts for tag: ${tagId}, status: ${response.status}`);
        return null;
      }

      return (await response.json()) as FanslyResponse<PostsResponseData>;
    } catch (error) {
      this.logger.error(`Error fetching posts for tag ${tagId}:`, error);
      return null;
    }
  }

  extractTagsFromContent(content: string): string[] {
    const tagRegex = /#(\w+)/g;
    const matches = content.match(tagRegex);

    if (!matches) return [];

    // Remove the # and convert to lowercase
    return matches.map((tag) => tag.substring(1).toLowerCase());
  }
}

export const fanslyClient = new FanslyClient();

// Export types for external use
export type { FanslyResponse, FanslyTag, FanslyPost, TagResponseData, PostsResponseData };

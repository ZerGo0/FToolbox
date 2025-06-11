import { Logger } from '../utils/logger';

interface FanslyTag {
  id: string;
  tag: string;
  description: string;
  viewCount: number;
  flags: number;
  createdAt: number;
}

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

      const data = (await response.json()) as {
        success: boolean;
        response?: {
          mediaOfferSuggestionTag?: FanslyTag;
        };
      };

      if (data.success && data.response?.mediaOfferSuggestionTag) {
        return data.response.mediaOfferSuggestionTag;
      }

      return null;
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

      const data = (await response.json()) as {
        success: boolean;
        response?: {
          aggregationData?: {
            posts?: FanslyPost[];
          };
        };
      };

      if (data.success && data.response?.aggregationData?.posts) {
        return data.response.aggregationData.posts;
      }

      return [];
    } catch (error) {
      this.logger.error(`Error fetching posts for tag ${tagId}:`, error);
      return [];
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

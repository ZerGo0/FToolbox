interface TagResponse {
  success: boolean;
  response: {
    mediaOfferSuggestionTag: {
      id: string;
      tag: string;
      description: string;
      viewCount: number;
      flags: number;
      createdAt: number;
    };
    aggregationData: Record<string, unknown>;
  };
}

interface Post {
  id: string;
  accountId: string;
  content: string;
  fypFlags: number;
  inReplyTo: string | null;
  inReplyToRoot: string | null;
  replyPermissionFlags: unknown | null;
  createdAt: number;
  expiresAt: number | null;
  attachments: Array<{
    postId: string;
    pos: number;
    contentType: number;
    contentId: string;
  }>;
  likeCount: number;
  replyCount: number;
  wallIds: string[];
  mediaLikeCount: number;
  totalTipAmount: number;
  attachmentTipAmount: number;
  accountMentions: Array<{
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

interface PostsResponse {
  success: boolean;
  response: {
    mediaOfferSuggestions: unknown[];
    aggregationData: {
      accounts: unknown[];
      accountMedia: unknown[];
      accountMediaBundles: unknown[];
      posts: Post[];
      tips: unknown[];
      tipGoals: unknown[];
      stories: unknown[];
    };
  };
}

export class FanslyAPI {
  private baseUrl = 'https://apiv3.fansly.com/api/v1';

  async getTagViewCount(tag: string): Promise<TagResponse | null> {
    try {
      const response = await fetch(
        `${this.baseUrl}/contentdiscovery/media/tag?tag=${encodeURIComponent(tag)}&ngsw-bypass=true`
      );

      if (!response.ok) {
        console.error(`Failed to fetch tag data: ${response.status}`);
        return null;
      }

      return (await response.json()) as TagResponse;
    } catch (error) {
      console.error('Error fetching tag data:', error);
      return null;
    }
  }

  async getPostsForTag(tagId: string, limit = 25, offset = 0): Promise<PostsResponse | null> {
    try {
      const response = await fetch(
        `${this.baseUrl}/contentdiscovery/media/suggestionsnew?before=0&after=0&tagIds=${tagId}&limit=${limit}&offset=${offset}&ngsw-bypass=true`
      );

      if (!response.ok) {
        console.error(`Failed to fetch posts: ${response.status}`);
        return null;
      }

      return (await response.json()) as PostsResponse;
    } catch (error) {
      console.error('Error fetching posts:', error);
      return null;
    }
  }

  extractTagsFromContent(content: string): string[] {
    const tagRegex = /#[a-zA-Z0-9_]+/g;
    const matches = content.match(tagRegex);
    return matches ? matches.map((tag) => tag.substring(1)) : [];
  }
}

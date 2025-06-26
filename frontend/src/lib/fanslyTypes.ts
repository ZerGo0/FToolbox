export interface FanslyAccount {
  id: string;
  username: string;
  displayName?: string;
  flags: number;
  version: number;
  followCount: number;
  subscriberCount: number;
  permissions: {
    accountPermissionFlags: {
      flags: number;
    };
  };
  timelineStats: {
    accountId: string;
    imageCount: number;
    videoCount: number;
    bundleCount: number;
    bundleImageCount: number;
    bundleVideoCount: number;
    fetchedAt: number;
  };
  accountMediaLikes: number;
  mediaStoryState?: {
    accountId: string;
    status: number;
    storyCount: number;
    version: number;
    createdAt: number;
    updatedAt: number;
    hasActiveStories: boolean;
  };
  statusId: number;
  lastSeenAt: number;
  profileAccessFlags: number;
  profileFlags: number;
  about: string;
  location: string;
  profileSocials: Array<{
    providerId: string;
    handle: string;
  }>;
  profileBadges: Array<{
    accountId: string;
    badgeId: string;
    badgeType: number;
    badgeDescription: string;
    displayFlags: number;
    metadata: string;
    createdAt: number;
  }>;
  pinnedPosts?: Array<{
    postId: string;
    accountId: string;
    pos: number;
    createdAt: number;
  }>;
  postLikes: number;
  streaming: {
    accountId: string;
    channel?: {
      id: string;
      accountId: string;
      playbackUrl: string;
      chatRoomId: string;
      status: number;
      version: number;
      createdAt: number;
      updatedAt: unknown;
      stream: {
        id: string;
        historyId?: string;
        channelId: string;
        accountId: string;
        title: string;
        status: number;
        viewerCount: number;
        version: number;
        createdAt: number;
        updatedAt: unknown;
        lastFetchedAt?: number;
        startedAt?: number;
        permissions: {
          permissionFlags: Array<{
            id: string;
            streamId: string;
            type: number;
            flags: number;
            price: number;
            metadata: string;
          }>;
        };
      };
      arn: unknown;
      ingestEndpoint: unknown;
    };
    enabled: boolean;
  };
  subscriptionTiers: Array<{
    id: string;
    accountId: string;
    name: string;
    color: string;
    pos: number;
    price: number;
    maxSubscribers: number;
    subscriptionBenefits: Array<string>;
    includedTierIds: Array<string>;
    plans: Array<{
      id: string;
      status: number;
      billingCycle: number;
      price: number;
      useAmounts: number;
      promos: Array<{
        id: string;
        status: number;
        price: number;
        duration: number;
        maxUses: number;
        maxUsesBefore?: number;
        newSubscribersOnly: number;
        description: string;
        startsAt: number;
        endsAt: number;
        uses: number;
      }>;
      uses: number;
    }>;
    maxSubscribersReached?: boolean;
  }>;
  walls: Array<{
    id: string;
    accountId: string;
    pos?: number;
    name: string;
    description: string;
    private: number;
    metadata: string;
    defaultWall?: boolean;
    mainWall?: boolean;
  }>;
  hasMainWall: boolean;
  avatar?: {
    id: string;
    type: number;
    status: number;
    accountId: string;
    mimetype: string;
    flags: number;
    location: string;
    width: number;
    height: number;
    metadata: string;
    updatedAt: number;
    createdAt: number;
    variants: Array<{
      id: string;
      type: number;
      status: number;
      mimetype: string;
      flags: number;
      location: string;
      width: number;
      height: number;
      metadata?: string;
      updatedAt: number;
      locations: Array<{
        locationId: string;
        location: string;
      }>;
    }>;
    variantHash: unknown;
    locations: Array<{
      locationId: string;
      location: string;
    }>;
  };
  banner?: {
    id: string;
    type: number;
    status: number;
    accountId: string;
    mimetype: string;
    flags: number;
    location: string;
    width: number;
    height: number;
    metadata: string;
    updatedAt: number;
    createdAt: number;
    variants: Array<{
      id: string;
      type: number;
      status: number;
      mimetype: string;
      flags: number;
      location: string;
      width: number;
      height: number;
      metadata?: string;
      updatedAt: number;
      locations: Array<{
        locationId: string;
        location: string;
      }>;
    }>;
    variantHash: unknown;
    locations: Array<{
      locationId: string;
      location: string;
    }>;
  };
  profileAccess: boolean;
  createdAt?: number;
}

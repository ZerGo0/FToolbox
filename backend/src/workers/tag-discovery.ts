import { db } from '../db';
import { tags } from '../db/schema';
import { eq } from 'drizzle-orm';
import { fanslyClient } from '../fansly/client';
import { logger } from '../utils/logger';
import type { Worker } from './manager';

class TagDiscoveryWorker implements Worker {
  name = 'tag-discovery';
  interval: number;

  // Popular tags to search for new tags
  private seedTags = [
    'teen',
    'young',
    'petite',
    'blonde',
    'brunette',
    'amateur',
    'cute',
    'pussy',
    'ass',
    'boobs',
    'fyp',
    'nsfw',
    'nude',
    'naked',
    'sexy',
  ];

  constructor() {
    this.interval = parseInt(process.env.WORKER_DISCOVERY_INTERVAL || '30000'); // Default: 30 seconds for continuous discovery
  }

  async run(): Promise<void> {
    logger.info('Starting tag discovery process');
    const rateLimitDelay = 60000 / parseInt(process.env.FANSLY_API_RATE_LIMIT || '60'); // Default: 60 requests per minute

    try {
      const discoveredTags = new Set<string>();
      let newTags = 0;
      let errors = 0;

      // Randomly select 2-3 seed tags to process each run to avoid overwhelming the API
      const shuffled = [...this.seedTags].sort(() => 0.5 - Math.random());
      const selectedTags = shuffled.slice(0, Math.floor(Math.random() * 2) + 2);

      logger.info(`Processing ${selectedTags.length} seed tags for discovery`);

      // Fetch posts from selected seed tags
      for (const seedTag of selectedTags) {
        try {
          logger.info(`Fetching posts for seed tag: ${seedTag}`);

          // First get the tag ID
          const tagData = await fanslyClient.getTag(seedTag);
          if (!tagData) {
            logger.warn(`Seed tag not found: ${seedTag}`);
            continue;
          }

          // Fetch posts for this tag
          const posts = await fanslyClient.getPostsForTag(tagData.id, 20); // Reduced to 20 for continuous operation

          // Extract tags from post content
          for (const post of posts) {
            const tagsInPost = fanslyClient.extractTagsFromContent(post.content);
            tagsInPost.forEach((tag) => discoveredTags.add(tag));
          }

          logger.info(`Found ${discoveredTags.size} unique tags from ${seedTag}`);

          // Rate limiting delay
          await new Promise((resolve) => setTimeout(resolve, rateLimitDelay));
        } catch (error) {
          logger.error(`Error processing seed tag ${seedTag}:`, error);
          errors++;
        }
      }

      // Check which tags are new and add them to database
      for (const tagName of discoveredTags) {
        try {
          // Check if tag already exists
          const existingTag = await db.select().from(tags).where(eq(tags.tag, tagName)).get();

          if (!existingTag) {
            // Fetch tag data from Fansly
            const tagData = await fanslyClient.getTag(tagName);

            if (tagData) {
              // Add new tag to database
              await db.insert(tags).values({
                id: tagData.id,
                tag: tagData.tag,
                viewCount: tagData.viewCount,
                fanslyCreatedAt: new Date(tagData.createdAt),
              });

              logger.info(`Discovered new tag: ${tagName} (${tagData.viewCount} views)`);
              newTags++;
            }

            // Rate limiting delay
            await new Promise((resolve) => setTimeout(resolve, rateLimitDelay));
          }
        } catch (error) {
          logger.error(`Error adding tag ${tagName}:`, error);
          errors++;
        }
      }

      logger.info(
        `Tag discovery completed. Discovered: ${discoveredTags.size} unique tags, Added: ${newTags} new tags, Errors: ${errors}`
      );
    } catch (error) {
      logger.error('Tag discovery process failed:', error);
      throw error;
    }
  }
}

export const tagDiscoveryWorker = new TagDiscoveryWorker();

import { db } from '../db';
import { tags, tagHistory } from '../db/schema';
import { eq } from 'drizzle-orm';
import { fanslyClient } from '../fansly/client';
import { logger } from '../utils/logger';
import type { Worker } from './manager';

class TagUpdaterWorker implements Worker {
  name = 'tag-updater';
  interval: number;

  constructor() {
    this.interval = parseInt(process.env.WORKER_UPDATE_INTERVAL || String(24 * 60 * 60 * 1000)); // Default: 24 hours
  }

  async run(): Promise<void> {
    logger.info('Starting tag update process');
    const rateLimitDelay = 60000 / parseInt(process.env.FANSLY_API_RATE_LIMIT || '60'); // Default: 60 requests per minute

    try {
      // Fetch all tracked tags
      const trackedTags = await db.select().from(tags).where(eq(tags.isTracked, true));
      logger.info(`Found ${trackedTags.length} tracked tags to update`);

      let updated = 0;
      let errors = 0;

      for (const tag of trackedTags) {
        try {
          // Fetch updated data from Fansly
          const tagData = await fanslyClient.getTag(tag.tag);

          if (tagData) {
            const currentViewCount = tagData.viewCount;
            const previousViewCount = tag.viewCount;

            // Update tag data
            await db
              .update(tags)
              .set({
                viewCount: currentViewCount,
                lastCheckedAt: new Date(),
              })
              .where(eq(tags.id, tag.id));

            // Create history record if view count changed
            if (currentViewCount !== previousViewCount) {
              await db.insert(tagHistory).values({
                tagId: tag.id,
                viewCount: currentViewCount,
                change: currentViewCount - previousViewCount,
                recordedAt: new Date(),
              });

              logger.info(
                `Updated tag "${tag.tag}": ${previousViewCount} -> ${currentViewCount} (${currentViewCount > previousViewCount ? '+' : ''}${currentViewCount - previousViewCount})`
              );
            } else {
              logger.info(`No change for tag "${tag.tag}": ${currentViewCount} views`);
            }

            updated++;
          } else {
            logger.warn(`Failed to fetch data for tag: ${tag.tag}`);
            errors++;
          }

          // Rate limiting delay
          await new Promise((resolve) => setTimeout(resolve, rateLimitDelay));
        } catch (error) {
          logger.error(`Error updating tag "${tag.tag}":`, error);
          errors++;
        }
      }

      logger.info(`Tag update completed. Updated: ${updated}, Errors: ${errors}`);
    } catch (error) {
      logger.error('Tag update process failed:', error);
      throw error;
    }
  }
}

export const tagUpdaterWorker = new TagUpdaterWorker();

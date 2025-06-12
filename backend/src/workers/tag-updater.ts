import { eq, sql } from 'drizzle-orm';
import { db } from '../db';
import { tagHistory, tags } from '../db/schema';
import { fanslyClient } from '../fansly/client';
import { logger } from '../utils/logger';
import type { Worker } from './manager';

class TagUpdaterWorker implements Worker {
  name = 'tag-updater';
  interval: number;

  constructor() {
    this.interval = parseInt(process.env.WORKER_UPDATE_INTERVAL || '10000'); // Default: 60 seconds for continuous running
  }

  async run(): Promise<void> {
    logger.info('Starting tag update process');
    const rateLimitDelay = 60000 / parseInt(process.env.FANSLY_API_RATE_LIMIT || '60'); // Default: 60 requests per minute
    const twentyFourHoursAgo = new Date(Date.now() - 24 * 60 * 60 * 1000);
    const twentyFourHoursAgoUnix = Math.floor(twentyFourHoursAgo.getTime() / 1000);

    try {
      // Fetch tags that need updating based on their latest history entry
      const tagsToUpdate = await db
        .select({
          id: tags.id,
          tag: tags.tag,
          viewCount: tags.viewCount,
          lastCheckedAt: tags.lastCheckedAt,
          lastHistoryCreatedAt: sql<number | null>`(
            SELECT created_at 
            FROM tag_history 
            WHERE tag_history.tag_id = ${tags.id} 
            ORDER BY created_at DESC 
            LIMIT 1
          )`.as('lastHistoryCreatedAt'),
        })
        .from(tags)
        .where(
          sql`
            NOT EXISTS (
              SELECT 1 
              FROM tag_history 
              WHERE tag_history.tag_id = ${tags.id} 
                AND tag_history.created_at >= ${twentyFourHoursAgoUnix}
            )
          `
        )
        .limit(10); // Process in batches to avoid long-running operations

      if (tagsToUpdate.length === 0) {
        logger.info('No tags need updating at this time');
        return;
      }

      logger.info(`Found ${tagsToUpdate.length} tags that need updating`);

      // Log the tags and their last history timestamps for validation
      for (const tag of tagsToUpdate) {
        const historyAge = tag.lastHistoryCreatedAt
          ? `${Math.floor((Date.now() / 1000 - tag.lastHistoryCreatedAt) / (60 * 60))} hours ago`
          : 'never';
        const historyDate = tag.lastHistoryCreatedAt
          ? new Date(tag.lastHistoryCreatedAt * 1000).toISOString()
          : 'no history';
        logger.info(`Tag "${tag.tag}" - Last history: ${historyAge} (${historyDate})`);
      }

      let updated = 0;
      let errors = 0;

      for (const tag of tagsToUpdate) {
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

            // Create history record
            await db.insert(tagHistory).values({
              tagId: tag.id,
              viewCount: currentViewCount,
              change: currentViewCount - previousViewCount,
            });

            logger.info(
              `Updated tag "${tag.tag}": ${previousViewCount} -> ${currentViewCount} (${currentViewCount > previousViewCount ? '+' : ''}${currentViewCount - previousViewCount})`
            );

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

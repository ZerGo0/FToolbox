import { db } from '../db';
import { tags } from '../db/schema';
import { eq, desc } from 'drizzle-orm';
import { logger } from '../utils/logger';
import type { Worker } from './manager';

class RankCalculatorWorker implements Worker {
  name = 'rank-calculator';
  interval: number;

  constructor() {
    // Run every hour by default
    this.interval = parseInt(process.env.WORKER_RANK_INTERVAL || String(60 * 60 * 1000));
  }

  async run(): Promise<void> {
    logger.info('Starting rank calculation process');

    try {
      // Get all tags ordered by viewCount descending
      const allTags = await db
        .select({ id: tags.id, tag: tags.tag, viewCount: tags.viewCount })
        .from(tags)
        .orderBy(desc(tags.viewCount));

      logger.info(`Calculating ranks for ${allTags.length} tags`);

      // Update each tag with its rank
      let previousViewCount = -1;
      let currentRank = 0;

      for (let i = 0; i < allTags.length; i++) {
        const tag = allTags[i];
        if (!tag) continue;

        // Handle ties - tags with same view count get same rank
        if (tag.viewCount !== previousViewCount) {
          currentRank = i + 1;
        }

        await db.update(tags).set({ rank: currentRank }).where(eq(tags.id, tag.id));

        previousViewCount = tag.viewCount;
      }

      logger.info(`Rank calculation completed for ${allTags.length} tags`);
    } catch (error) {
      logger.error('Rank calculation process failed:', error);
      throw error;
    }
  }
}

export const rankCalculatorWorker = new RankCalculatorWorker();

import { sql } from 'drizzle-orm';
import { db } from '../db';
import { tagHistory } from '../db/schema';
import { logger } from '../utils/logger';

async function cleanupDuplicateHistory() {
  logger.info('Starting duplicate history cleanup...');

  try {
    // Get all entries grouped by tagId and date
    const allEntries = await db
      .select({
        id: tagHistory.id,
        tagId: tagHistory.tagId,
        viewCount: tagHistory.viewCount,
        change: tagHistory.change,
        createdAt: tagHistory.createdAt,
        date: sql<string>`date(${tagHistory.createdAt})`.as('date'),
      })
      .from(tagHistory)
      .orderBy(tagHistory.tagId, tagHistory.createdAt);

    // Group by tagId and date, keeping track of duplicates
    const toDelete: number[] = [];
    const processed = new Map<string, number>(); // key: "tagId|date", value: id to keep

    for (const entry of allEntries) {
      const key = `${entry.tagId}|${entry.date}`;

      if (processed.has(key)) {
        // This is a duplicate, mark for deletion
        toDelete.push(entry.id);
      } else {
        // This is the first (earliest) entry for this tag/date combo
        processed.set(key, entry.id);
      }
    }

    if (toDelete.length === 0) {
      logger.info('No duplicate entries found.');
      return;
    }

    logger.info(`Found ${toDelete.length} duplicate entries to delete.`);

    // Delete duplicates in batches
    const batchSize = 100;
    for (let i = 0; i < toDelete.length; i += batchSize) {
      const batch = toDelete.slice(i, i + batchSize);
      await db.delete(tagHistory).where(
        sql`${tagHistory.id} IN (${sql.join(
          batch.map((id) => sql`${id}`),
          sql`, `
        )})`
      );

      logger.info(
        `Deleted batch ${Math.floor(i / batchSize) + 1}/${Math.ceil(toDelete.length / batchSize)}`
      );
    }

    logger.info(`Cleanup completed. Deleted ${toDelete.length} duplicate entries.`);
  } catch (error) {
    logger.error('Error during cleanup:', error);
    throw error;
  }
}

// Run the cleanup if this script is executed directly
if (import.meta.main) {
  cleanupDuplicateHistory()
    .then(() => {
      logger.info('Cleanup script finished successfully');
      process.exit(0);
    })
    .catch((error) => {
      logger.error('Cleanup script failed:', error);
      process.exit(1);
    });
}

export { cleanupDuplicateHistory };

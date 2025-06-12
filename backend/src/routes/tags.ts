import { and, asc, desc, eq, gt, gte, like, lte } from 'drizzle-orm';
import { Hono } from 'hono';
import { db } from '../db';
import { tagHistory, tagRequests, tags } from '../db/schema';
import { fanslyClient } from '../fansly/client';
import { serializeDates } from '../utils/serialize';

const app = new Hono();

// Get all tags with pagination, sorting, and filtering
app.get('/', async (c) => {
  const page = parseInt(c.req.query('page') || '1');
  const limit = parseInt(c.req.query('limit') || '20');
  const offset = (page - 1) * limit;
  const search = c.req.query('search') || '';
  const sortBy = c.req.query('sortBy') || 'viewCount';
  const sortOrder = c.req.query('sortOrder') || 'desc';
  const includeHistory = c.req.query('includeHistory') === 'true';
  const historyStartDate = c.req.query('historyStartDate');
  const historyEndDate = c.req.query('historyEndDate');

  try {
    // Build where conditions
    const whereConditions = search ? like(tags.tag, `%${search}%`) : undefined;

    // For change sorting, we need to include history regardless
    const needsHistory = includeHistory || sortBy === 'change';

    // Build order by
    let orderByColumn;
    if (sortBy === 'tag') {
      orderByColumn = sortOrder === 'desc' ? desc(tags.tag) : asc(tags.tag);
    } else if (sortBy === 'viewCount') {
      orderByColumn = sortOrder === 'desc' ? desc(tags.viewCount) : asc(tags.viewCount);
    } else if (sortBy === 'updatedAt') {
      orderByColumn = sortOrder === 'desc' ? desc(tags.updatedAt) : asc(tags.updatedAt);
    } else if (sortBy === 'rank') {
      orderByColumn = sortOrder === 'desc' ? desc(tags.rank) : asc(tags.rank);
    }
    // Note: 'change' sorting will be handled after fetching history

    // Execute query
    const query = db.select().from(tags);
    if (whereConditions) {
      query.where(whereConditions);
    }

    // If sorting by change, we need to fetch all results first
    let results;
    if (sortBy === 'change') {
      // Fetch all matching results for change calculation
      results = await query;
    } else {
      // Normal pagination
      if (orderByColumn) {
        query.orderBy(orderByColumn);
      }
      results = await query.limit(limit).offset(offset);
    }

    // Get total count for pagination
    const countQuery = db.select().from(tags);
    if (whereConditions) {
      countQuery.where(whereConditions);
    }
    const countResult = await countQuery;
    const totalCount = countResult.length;

    // Fetch history for each tag if requested or if sorting by change
    let tagsWithHistory = results;
    if (needsHistory) {
      tagsWithHistory = await Promise.all(
        results.map(async (tag) => {
          const conditions = [eq(tagHistory.tagId, tag.id)];

          if (historyStartDate && historyEndDate) {
            conditions.push(
              gte(tagHistory.createdAt, new Date(historyStartDate)),
              lte(tagHistory.createdAt, new Date(historyEndDate))
            );
          }

          const history = await db
            .select()
            .from(tagHistory)
            .where(and(...conditions))
            .orderBy(desc(tagHistory.createdAt));

          // Calculate changes between history points
          const historyWithChanges = history.map((point, index) => {
            if (index === history.length - 1) {
              // Last point has no previous point to compare
              return { ...point, change: 0, changePercent: 0 };
            }

            const previousPoint = history[index + 1];
            if (!previousPoint) {
              return { ...point, change: 0, changePercent: 0 };
            }

            const change = point.viewCount - previousPoint.viewCount;
            const changePercent =
              previousPoint.viewCount > 0 ? (change / previousPoint.viewCount) * 100 : 0;

            return { ...point, change, changePercent };
          });

          // Calculate total change for sorting
          let totalChange = 0;
          if (history.length > 0) {
            const newest = history[0]?.viewCount || 0;
            const oldest = history[history.length - 1]?.viewCount || 0;
            totalChange = newest - oldest;
          }

          return { ...tag, history: includeHistory ? historyWithChanges : undefined, totalChange };
        })
      );
    }

    // Sort by change if requested
    if (sortBy === 'change') {
      tagsWithHistory.sort((a, b) => {
        const aChange = 'totalChange' in a ? (a.totalChange as number) : 0;
        const bChange = 'totalChange' in b ? (b.totalChange as number) : 0;
        return sortOrder === 'desc' ? bChange - aChange : aChange - bChange;
      });

      // Apply pagination after sorting
      tagsWithHistory = tagsWithHistory.slice(offset, offset + limit);
    }

    return c.json(
      serializeDates({
        tags: tagsWithHistory,
        pagination: {
          page,
          limit,
          totalCount,
          totalPages: Math.ceil(totalCount / limit),
        },
      })
    );
  } catch (error) {
    console.error('Error fetching tags:', error);
    return c.json({ error: 'Failed to fetch tags' }, 500);
  }
});

// Get tag history with date range
app.get('/:tagId/history', async (c) => {
  const tagId = c.req.param('tagId');
  const startDate = c.req.query('startDate');
  const endDate = c.req.query('endDate');

  try {
    const conditions = [eq(tagHistory.tagId, tagId)];

    if (startDate && endDate) {
      conditions.push(
        gte(tagHistory.createdAt, new Date(startDate)),
        lte(tagHistory.createdAt, new Date(endDate))
      );
    }

    const history = await db
      .select()
      .from(tagHistory)
      .where(and(...conditions))
      .orderBy(desc(tagHistory.createdAt));

    return c.json(serializeDates({ history }));
  } catch (error) {
    console.error('Error fetching tag history:', error);
    return c.json({ error: 'Failed to fetch tag history' }, 500);
  }
});

// Request a new tag to track
app.post('/request', async (c) => {
  const { tag } = await c.req.json();

  if (!tag) {
    return c.json({ error: 'Tag is required' }, 400);
  }

  try {
    // Check if tag already exists
    const existingTag = await db.select().from(tags).where(eq(tags.tag, tag)).limit(1);

    if (existingTag.length > 0) {
      return c.json(
        serializeDates({ message: 'Tag is already being tracked', tag: existingTag[0] })
      );
    }

    // Create tag request
    const requestResult = await db.insert(tagRequests).values({ tag }).returning();
    const request = requestResult[0];
    if (!request) {
      return c.json({ error: 'Failed to create tag request' }, 500);
    }

    // Immediately try to fetch tag data
    const tagData = await fanslyClient.getTagResponse(tag);

    if (tagData && tagData.success && tagData.response?.mediaOfferSuggestionTag) {
      const fanslyTag = tagData.response.mediaOfferSuggestionTag;

      // Insert tag into database
      const newTagResult = await db
        .insert(tags)
        .values({
          id: fanslyTag.id,
          tag: fanslyTag.tag,
          viewCount: fanslyTag.viewCount,
          fanslyCreatedAt: new Date(fanslyTag.createdAt),
          lastCheckedAt: new Date(),
        })
        .returning();
      const newTag = newTagResult[0];

      if (!newTag) {
        return c.json({ error: 'Failed to create tag' }, 500);
      }

      // Insert initial history record
      await db.insert(tagHistory).values({
        tagId: newTag.id,
        viewCount: newTag.viewCount,
        change: 0, // Initial entry has no change
      });

      // Update request status
      await db
        .update(tagRequests)
        .set({ status: 'completed', updatedAt: new Date() })
        .where(eq(tagRequests.id, request.id));

      // Calculate rank for the new tag
      const higherRankedCount = await db
        .select()
        .from(tags)
        .where(gt(tags.viewCount, newTag.viewCount))
        .then((rows) => rows.length);

      const rank = higherRankedCount + 1;
      await db.update(tags).set({ rank }).where(eq(tags.id, newTag.id));

      return c.json(
        serializeDates({
          message: 'Tag added successfully',
          tag: { ...newTag, rank },
        })
      );
    } else {
      // Update request status to failed
      await db
        .update(tagRequests)
        .set({ status: 'failed', updatedAt: new Date() })
        .where(eq(tagRequests.id, request.id));

      return c.json({ error: 'Tag not found on Fansly' }, 404);
    }
  } catch (error) {
    console.error('Error requesting tag:', error);
    return c.json({ error: 'Failed to request tag' }, 500);
  }
});

export default app;

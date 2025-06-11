import { and, asc, desc, eq, gte, like, lte } from 'drizzle-orm';
import { Hono } from 'hono';
import { db } from '../db';
import { tagHistory, tagRequests, tags } from '../db/schema';
import { FanslyAPI } from '../lib/fansly-api';

const app = new Hono();
const fanslyAPI = new FanslyAPI();

// Get all tags with pagination, sorting, and filtering
app.get('/', async (c) => {
  const page = parseInt(c.req.query('page') || '1');
  const limit = parseInt(c.req.query('limit') || '20');
  const offset = (page - 1) * limit;
  const search = c.req.query('search') || '';
  const sortBy = c.req.query('sortBy') || 'viewCount';
  const sortOrder = c.req.query('sortOrder') || 'desc';

  try {
    // Build where conditions
    const whereConditions = search ? like(tags.tag, `%${search}%`) : undefined;

    // Build order by
    let orderByColumn;
    if (sortBy === 'tag') {
      orderByColumn = sortOrder === 'desc' ? desc(tags.tag) : asc(tags.tag);
    } else if (sortBy === 'viewCount') {
      orderByColumn = sortOrder === 'desc' ? desc(tags.viewCount) : asc(tags.viewCount);
    } else if (sortBy === 'updatedAt') {
      orderByColumn = sortOrder === 'desc' ? desc(tags.updatedAt) : asc(tags.updatedAt);
    }

    // Execute query
    const query = db.select().from(tags);
    if (whereConditions) {
      query.where(whereConditions);
    }
    if (orderByColumn) {
      query.orderBy(orderByColumn);
    }
    const results = await query.limit(limit).offset(offset);

    // Get total count for pagination
    const countQuery = db.select().from(tags);
    if (whereConditions) {
      countQuery.where(whereConditions);
    }
    const countResult = await countQuery;
    const totalCount = countResult.length;

    return c.json({
      tags: results,
      pagination: {
        page,
        limit,
        totalCount,
        totalPages: Math.ceil(totalCount / limit),
      },
    });
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

    return c.json({ history });
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
      return c.json({ message: 'Tag is already being tracked', tag: existingTag[0] });
    }

    // Create tag request
    const requestResult = await db.insert(tagRequests).values({ tag }).returning();
    const request = requestResult[0];
    if (!request) {
      return c.json({ error: 'Failed to create tag request' }, 500);
    }

    // Immediately try to fetch tag data
    const tagData = await fanslyAPI.getTagViewCount(tag);

    if (tagData && tagData.success) {
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
        .where(gte(tags.viewCount, newTag.viewCount))
        .then((rows) => rows.length);

      await db.update(tags).set({ rank: higherRankedCount }).where(eq(tags.id, newTag.id));

      return c.json({
        message: 'Tag added successfully',
        tag: { ...newTag, rank: higherRankedCount },
      });
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

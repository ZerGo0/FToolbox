import { sql } from 'drizzle-orm';
import { index, integer, sqliteTable, text } from 'drizzle-orm/sqlite-core';

export const tags = sqliteTable(
  'tags',
  {
    id: text('id').primaryKey(),
    tag: text('tag').notNull().unique(),
    viewCount: integer('view_count').notNull(),
    rank: integer('rank'), // Global rank based on view count
    fanslyCreatedAt: integer('fansly_created_at', { mode: 'timestamp' }).notNull(),
    lastCheckedAt: integer('last_checked_at', { mode: 'timestamp' }),
    lastUsedForDiscovery: integer('last_used_for_discovery', { mode: 'timestamp' }),
    createdAt: integer('created_at', { mode: 'timestamp' })
      .default(sql`(unixepoch())`)
      .notNull(),
    updatedAt: integer('updated_at', { mode: 'timestamp' })
      .default(sql`(unixepoch())`)
      .notNull(),
  },
  (table) => {
    return {
      tagIdx: index('tag_idx').on(table.tag),
      viewCountIdx: index('view_count_idx').on(table.viewCount),
      rankIdx: index('rank_idx').on(table.rank),
    };
  }
);

export const tagHistory = sqliteTable(
  'tag_history',
  {
    id: integer('id').primaryKey({ autoIncrement: true }),
    tagId: text('tag_id')
      .notNull()
      .references(() => tags.id),
    viewCount: integer('view_count').notNull(),
    change: integer('change').notNull(), // Change from previous value
    createdAt: integer('created_at', { mode: 'timestamp' })
      .default(sql`(unixepoch())`)
      .notNull(),
    updatedAt: integer('updated_at', { mode: 'timestamp' })
      .default(sql`(unixepoch())`)
      .notNull(),
  },
  (table) => {
    return {
      tagIdIdx: index('tag_history_tag_id_idx').on(table.tagId),
      createdAtIdx: index('tag_history_created_at_idx').on(table.createdAt),
    };
  }
);

export const tagRequests = sqliteTable('tag_requests', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  tag: text('tag').notNull(),
  status: text('status').notNull().default('pending'), // pending, processing, completed, failed
  createdAt: integer('created_at', { mode: 'timestamp' })
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull(),
  updatedAt: integer('updated_at', { mode: 'timestamp' })
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull(),
});

export const workers = sqliteTable('workers', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  name: text('name').notNull().unique(),
  lastRunAt: integer('last_run_at', { mode: 'timestamp' }),
  nextRunAt: integer('next_run_at', { mode: 'timestamp' }),
  status: text('status').notNull().default('idle'), // idle, running, failed
  lastError: text('last_error'),
  runCount: integer('run_count').notNull().default(0),
  successCount: integer('success_count').notNull().default(0),
  failureCount: integer('failure_count').notNull().default(0),
  isEnabled: integer('is_enabled', { mode: 'boolean' }).default(true).notNull(),
  createdAt: integer('created_at', { mode: 'timestamp' })
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull(),
  updatedAt: integer('updated_at', { mode: 'timestamp' })
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull(),
});

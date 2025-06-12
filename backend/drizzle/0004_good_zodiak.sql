PRAGMA foreign_keys=OFF;--> statement-breakpoint
CREATE TABLE `__new_tag_history` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`tag_id` text NOT NULL,
	`view_count` integer NOT NULL,
	`change` integer NOT NULL,
	`created_at` integer DEFAULT (unixepoch()) NOT NULL,
	`updated_at` integer DEFAULT (unixepoch()) NOT NULL,
	FOREIGN KEY (`tag_id`) REFERENCES `tags`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
INSERT INTO `__new_tag_history`("id", "tag_id", "view_count", "change", "created_at", "updated_at") 
SELECT 
  "id", 
  "tag_id", 
  "view_count", 
  "change", 
  CASE 
    WHEN typeof("created_at") = 'text' THEN unixepoch("created_at")
    ELSE "created_at"
  END,
  CASE 
    WHEN typeof("updated_at") = 'text' THEN unixepoch("updated_at")
    ELSE "updated_at"
  END
FROM `tag_history`;--> statement-breakpoint
DROP TABLE `tag_history`;--> statement-breakpoint
ALTER TABLE `__new_tag_history` RENAME TO `tag_history`;--> statement-breakpoint
PRAGMA foreign_keys=ON;--> statement-breakpoint
CREATE INDEX `tag_history_tag_id_idx` ON `tag_history` (`tag_id`);--> statement-breakpoint
CREATE INDEX `tag_history_created_at_idx` ON `tag_history` (`created_at`);--> statement-breakpoint
CREATE TABLE `__new_tags` (
	`id` text PRIMARY KEY NOT NULL,
	`tag` text NOT NULL,
	`view_count` integer NOT NULL,
	`rank` integer,
	`fansly_created_at` integer NOT NULL,
	`last_checked_at` integer,
	`created_at` integer DEFAULT (unixepoch()) NOT NULL,
	`updated_at` integer DEFAULT (unixepoch()) NOT NULL
);
--> statement-breakpoint
INSERT INTO `__new_tags`("id", "tag", "view_count", "rank", "fansly_created_at", "last_checked_at", "created_at", "updated_at") 
SELECT 
  "id", 
  "tag", 
  "view_count", 
  "rank", 
  "fansly_created_at", 
  "last_checked_at", 
  CASE 
    WHEN typeof("created_at") = 'text' THEN unixepoch("created_at")
    ELSE "created_at"
  END,
  CASE 
    WHEN typeof("updated_at") = 'text' THEN unixepoch("updated_at")
    ELSE "updated_at"
  END
FROM `tags`;--> statement-breakpoint
DROP TABLE `tags`;--> statement-breakpoint
ALTER TABLE `__new_tags` RENAME TO `tags`;--> statement-breakpoint
CREATE UNIQUE INDEX `tags_tag_unique` ON `tags` (`tag`);--> statement-breakpoint
CREATE INDEX `tag_idx` ON `tags` (`tag`);--> statement-breakpoint
CREATE INDEX `view_count_idx` ON `tags` (`view_count`);--> statement-breakpoint
CREATE INDEX `rank_idx` ON `tags` (`rank`);
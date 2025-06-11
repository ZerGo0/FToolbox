ALTER TABLE `tag_history` DROP COLUMN `recorded_at`;--> statement-breakpoint
ALTER TABLE `tags` DROP COLUMN `description`;--> statement-breakpoint
ALTER TABLE `tags` DROP COLUMN `flags`;--> statement-breakpoint
ALTER TABLE `tags` DROP COLUMN `is_tracked`;
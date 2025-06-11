ALTER TABLE `tags` ADD `rank` integer;--> statement-breakpoint
CREATE INDEX `view_count_idx` ON `tags` (`view_count`);--> statement-breakpoint
CREATE INDEX `rank_idx` ON `tags` (`rank`);
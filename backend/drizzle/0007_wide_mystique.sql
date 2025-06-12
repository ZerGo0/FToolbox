CREATE INDEX `tag_history_tag_id_idx` ON `tag_history` (`tag_id`);--> statement-breakpoint
CREATE INDEX `tag_history_created_at_idx` ON `tag_history` (`created_at`);--> statement-breakpoint
CREATE INDEX `tag_idx` ON `tags` (`tag`);--> statement-breakpoint
CREATE INDEX `view_count_idx` ON `tags` (`view_count`);--> statement-breakpoint
CREATE INDEX `rank_idx` ON `tags` (`rank`);
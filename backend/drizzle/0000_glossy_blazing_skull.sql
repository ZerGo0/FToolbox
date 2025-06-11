CREATE TABLE `tag_history` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`tag_id` text NOT NULL,
	`view_count` integer NOT NULL,
	`created_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
	`updated_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
	FOREIGN KEY (`tag_id`) REFERENCES `tags`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE INDEX `tag_history_tag_id_idx` ON `tag_history` (`tag_id`);--> statement-breakpoint
CREATE INDEX `tag_history_created_at_idx` ON `tag_history` (`created_at`);--> statement-breakpoint
CREATE TABLE `tag_requests` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`tag` text NOT NULL,
	`status` text DEFAULT 'pending' NOT NULL,
	`created_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
	`updated_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL
);
--> statement-breakpoint
CREATE TABLE `tags` (
	`id` text PRIMARY KEY NOT NULL,
	`tag` text NOT NULL,
	`description` text,
	`view_count` integer NOT NULL,
	`flags` integer DEFAULT 0,
	`fansly_created_at` integer NOT NULL,
	`created_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
	`updated_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `tags_tag_unique` ON `tags` (`tag`);--> statement-breakpoint
CREATE INDEX `tag_idx` ON `tags` (`tag`);
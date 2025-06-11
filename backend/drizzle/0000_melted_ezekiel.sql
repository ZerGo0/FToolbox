CREATE TABLE `users` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`email` text NOT NULL,
	`name` text,
	`created_at` integer DEFAULT '"2025-06-11T18:41:36.962Z"' NOT NULL,
	`updated_at` integer DEFAULT '"2025-06-11T18:41:36.962Z"' NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `users_email_unique` ON `users` (`email`);
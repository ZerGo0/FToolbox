CREATE TABLE `workers` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`name` text NOT NULL,
	`last_run_at` integer,
	`next_run_at` integer,
	`status` text DEFAULT 'idle' NOT NULL,
	`last_error` text,
	`run_count` integer DEFAULT 0 NOT NULL,
	`success_count` integer DEFAULT 0 NOT NULL,
	`failure_count` integer DEFAULT 0 NOT NULL,
	`is_enabled` integer DEFAULT true NOT NULL,
	`created_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
	`updated_at` integer DEFAULT CURRENT_TIMESTAMP NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `workers_name_unique` ON `workers` (`name`);
# Database Architecture

## Database System
- **Type**: MariaDB
- **ORM**: GORM (Go Object-Relational Mapping)
- **Connection**: Configured in `/backend-go/database/connection.go`
- **Auto-Migration**: GORM AutoMigrate on startup

## Core Tables

### Tags
- Stores Fansly tags with metadata
- Fields: name, display_name, view_count, rank, heat, last_updated
- Relationships: Has many tag_histories

### Tag Histories
- Historical view count tracking for tags
- Fields: tag_id, view_count, timestamp
- Used for trend analysis and charts

### Creators
- Stores Fansly creator information
- Fields: username, display_name, subscriber_count, rank, heat
- Relationships: Has many creator_histories

### Creator Histories
- Historical subscriber tracking for creators
- Fields: creator_id, subscriber_count, timestamp

### Workers
- Worker status and run statistics
- Fields: name, enabled, last_run, next_run, run_count, error_count
- Tracks background job execution

### Tag/Creator Requests
- Queue for new tag/creator addition requests
- Fields: name, status, requested_at, processed_at
- Status: pending, approved, rejected

## Migration Strategy
1. **Auto-Migration**: GORM automatically creates/updates tables based on model structs
2. **Manual Migrations**: Complex operations in `/backend-go/database/sql_migrations.go`
3. **Model Changes**: Update struct in `/backend-go/models/` and GORM handles schema updates

## Indexing
- Primary keys on all ID fields
- Unique constraints on tag/creator names
- Indexes on frequently queried fields (rank, heat, last_updated)
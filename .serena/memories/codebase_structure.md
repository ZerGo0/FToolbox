# Codebase Structure

## Root Directory
```
/
├── frontend/           # SvelteKit frontend application
├── backend-go/         # Go backend API server
├── Taskfile.yml        # Task runner configuration
├── CLAUDE.md          # Instructions for Claude Code
├── README.md          # Project documentation
└── .gitignore         # Git ignore rules
```

## Frontend Structure
```
frontend/
├── src/
│   ├── routes/        # File-based routing (pages)
│   ├── lib/
│   │   ├── components/
│   │   │   ├── ui/    # shadcn-svelte components
│   │   │   └── *.svelte # Custom components
│   │   ├── hooks/     # Custom Svelte hooks
│   │   ├── utils.ts   # Utility functions
│   │   └── fanslyTypes.ts # TypeScript type definitions
│   ├── app.html       # HTML template
│   ├── app.css        # Global styles
│   └── app.d.ts       # App type definitions
├── static/            # Static assets
├── package.json       # Dependencies and scripts
├── svelte.config.js   # SvelteKit configuration
├── vite.config.ts     # Vite configuration
├── tailwind.config.js # Tailwind CSS configuration
└── wrangler.toml      # Cloudflare deployment config
```

## Backend Structure
```
backend-go/
├── main.go            # Application entry point
├── config/            # Environment configuration
├── database/
│   ├── connection.go  # Database connection setup
│   └── sql_migrations.go # Manual migrations
├── models/            # GORM database models
│   ├── tag.go
│   ├── creator.go
│   ├── worker.go
│   └── *_history.go   # Historical data models
├── handlers/          # HTTP request handlers
│   ├── tag.go
│   ├── creator.go
│   └── worker.go
├── routes/
│   └── routes.go      # API route definitions
├── workers/           # Background job system
│   ├── manager.go     # Worker orchestration
│   ├── tag_updater.go
│   ├── creator_updater.go
│   ├── tag_discovery.go
│   └── rank_calculator.go
├── fansly/
│   └── client.go      # Fansly API client
├── ratelimit/         # Rate limiting utilities
├── utils/             # Helper functions
├── .air.toml          # Air hot reload config
└── docker-compose.yml # Docker configuration
```

## Key API Endpoints
- `GET /api/tags` - List tags with pagination/filtering  
- `GET /api/creators` - List creators
- `POST /api/tags/request` - Request new tag tracking
- `POST /api/creators/request` - Request new creator tracking
- `GET /api/workers/status` - Worker system status
- `GET /api/health` - Health check
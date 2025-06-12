# FToolbox Backend

Bun-based backend API for FToolbox with SQLite database and background workers.

## Development

```bash
# Install dependencies
bun install

# Run development server with auto-reload
bun dev

# Run type checking
bun check

# Run linting
bun lint
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```env
DATABASE_URL=file:./sqlite.db
WORKER_ENABLED=true
CORS_ORIGIN=http://localhost:5173
PORT=3000
```

## Docker Deployment

### Build locally

```bash
# Build Docker image
docker build -t ftoolbox-backend:latest .

# Run with docker-compose
docker compose up -d
```

### Production Deployment

The backend is automatically deployed via GitHub Actions when pushing to main branch.

Required GitHub secrets:

- `SERVER_HOST` - Your server hostname
- `SERVER_PORT` - SSH port (usually 22)
- `SERVER_USERNAME` - SSH username
- `SSH_PRIVATE_KEY` - SSH private key for authentication
- `SSH_KEY_PASSPHRASE` - SSH key passphrase (if applicable)
- `SERVER_PATH` - Deployment path on server

### Manual Deployment

```bash
# Build and save Docker image
docker build -t ftoolbox-backend:latest .
docker save ftoolbox-backend:latest > image.tar

# Transfer to server
scp image.tar docker-compose.prod.yml load_container.sh user@server:/path/to/deployment/

# On server
sudo ./load_container.sh
```

## Database

- SQLite database with Drizzle ORM
- Migrations run automatically on startup
- Database file persisted in Docker volume

## Workers

Background workers run automatically when `WORKER_ENABLED=true`:

- **tag-updater**: Updates tag view counts (24h interval)
- **tag-discovery**: Discovers new tags from Fansly
- **rank-calculator**: Calculates global tag ranks

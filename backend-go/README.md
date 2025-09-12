# FToolbox Backend (Go)

Go implementation of the FToolbox backend API.

## Prerequisites

- Go 1.21+
- MariaDB
- Air (for hot reloading): `go install github.com/cosmtrek/air@latest`

## Setup

1. Copy `.env.example` to `.env` and configure your database settings
2. Install dependencies: `go mod download`
3. Run with hot reloading: `air`
4. Or run directly: `go run main.go`

## API Endpoints

- `GET /api/tags` - List tags with pagination/filtering
- `GET /api/tags/:name` - Get single tag details
- `POST /api/tags/request` - Request new tag tracking
- `GET /api/tags/:name/history` - Get tag history
- `GET /api/tags/related` - Get related tags
- `GET /api/creators` - List creators with pagination/filtering
- `GET /api/creators/statistics` - Creator system statistics
- `POST /api/creators/request` - Request a creator by username
- `GET /api/workers/status` - Worker system status
- `GET /api/health` - Health check

## Rate Limiting

Two layers of rate limiting exist:

- Fansly client global limiter (outbound):
  - `FANSLY_GLOBAL_RATE_LIMIT` (default `50` requests)
  - `FANSLY_GLOBAL_RATE_LIMIT_WINDOW` (default `10` seconds)
- Fiber HTTP limiter (per-client inbound):
  - `HTTP_RATE_LIMIT_MAX` (default `120` requests)
  - `HTTP_RATE_LIMIT_WINDOW` (default `60` seconds)

The Fiber limiter keys by `c.IP()` and trusts proxy headers. Ensure your reverse proxy (e.g., nginx) forwards the real client IP; see “Cloudflare/nginx real IP” below.

### Cloudflare/nginx real IP

If you’re running behind Cloudflare → nginx → Fiber, configure nginx to restore the real client IP and forward it upstream:

```
# /etc/nginx/conf.d/cloudflare-real-ip.conf
real_ip_header CF-Connecting-IP;
real_ip_recursive on;

# Trust Cloudflare proxies (keep this list updated from Cloudflare docs)
set_real_ip_from 103.21.244.0/22;
set_real_ip_from 103.22.200.0/22;
set_real_ip_from 103.31.4.0/22;
set_real_ip_from 104.16.0.0/13;
set_real_ip_from 104.24.0.0/14;
set_real_ip_from 108.162.192.0/18;
set_real_ip_from 131.0.72.0/22;
set_real_ip_from 141.101.64.0/18;
set_real_ip_from 162.158.0.0/15;
set_real_ip_from 172.64.0.0/13;
set_real_ip_from 173.245.48.0/20;
set_real_ip_from 188.114.96.0/20;
set_real_ip_from 190.93.240.0/20;
set_real_ip_from 197.234.240.0/22;
set_real_ip_from 198.41.128.0/17;
set_real_ip_from 2400:cb00::/32;
set_real_ip_from 2606:4700::/32;
set_real_ip_from 2803:f800::/32;
set_real_ip_from 2405:b500::/32;
set_real_ip_from 2405:8100::/32;
set_real_ip_from 2c0f:f248::/32;
set_real_ip_from 2a06:98c0::/29;
```

Then in your site config/server block:

```
location / {
    proxy_pass http://127.0.0.1:3000;

    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Real-IP $remote_addr; # Now the real client IP
}
```

With this, Fiber’s `c.IP()` returns the real client IP and the rate limiter applies per-user, not per-Cloudflare POP.

## Technologies

- **Fiber** - Web framework
- **GORM** - ORM with auto-migration
- **Zap** - Structured logging
- **Air** - Hot reloading
- **MariaDB** - Database

## Related Tags Endpoint

`GET /api/tags/related`

Query params:

- `tags` (required): comma-separated tag names to base the recommendations on.
- `limit` (optional): number of results to return (default 10, max 20).
- `windowDays` (optional): lookback window in days. Default 14, clamped to [7, 30].
- `minViewCount` (optional): minimum `view_count` for candidate tags. Default 5000.
- `minCoverage` (optional): minimum number of input tags a candidate must co-occur with. Default is ceil(40% of inputs).

Responses include per-tag fields:

- `id`, `tag`
- `score`: numeric score (equals `finalScore`)
- `normScore`, `coverage`, `finalScore`

Top-level metadata:

- `source`: `computed`
- `mode`, `windowDays`, `minViewCount`, `minCoverage`, `usedTagIds`

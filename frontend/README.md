# FToolbox Frontend

SvelteKit frontend for FToolbox, deployed on Cloudflare Pages.

## Development

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Run type checking
pnpm check

# Run linting
pnpm lint
```

## Environment Variables

Create a `.env` file in the frontend directory:

```env
PUBLIC_API_URL=http://localhost:3000
```

For production, set this in your Cloudflare Pages dashboard.

## Building

```bash
# Build for production
pnpm build

# Preview production build
pnpm preview
```

## Deployment to Cloudflare Pages

### First-time setup

1. Login to Cloudflare:

```bash
pnpm wrangler login
```

2. Create a new Pages project:

```bash
pnpm wrangler pages project create ftoolbox-frontend
```

### Deploy

```bash
# Deploy to production
pnpm pages:deploy

# Deploy to preview branch
pnpm pages:deploy:preview
```

### Environment Variables in Cloudflare

Set these in your Cloudflare Pages project settings:

- `PUBLIC_API_URL` - Your backend API URL (e.g., `https://api.yourdomain.com`)

### Local testing with Wrangler

```bash
# Test the production build locally with Cloudflare Pages environment
pnpm build
pnpm pages:dev
```

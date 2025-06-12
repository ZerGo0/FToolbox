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
# Deploy to production (uses env vars from Cloudflare dashboard)
pnpm pages:deploy

# Deploy to preview branch
pnpm pages:deploy:preview

# Deploy with custom API URL from command line
PUBLIC_API_URL=https://your-api.com pnpm pages:deploy:with-api
```

### Environment Variables in Cloudflare Pages

**IMPORTANT**: Environment variables for SvelteKit must be set in the Cloudflare Pages dashboard during the build process.

1. Go to your Cloudflare Pages project settings
2. Navigate to Settings → Environment variables
3. Add the following variables for **Production** (and optionally Preview):
   - `PUBLIC_API_URL` - Your backend API URL (e.g., `https://api.yourdomain.com`)

These variables are needed at BUILD TIME, not runtime. The wrangler.toml `[vars]` section does NOT work for SvelteKit environment variables.

### Local testing with Wrangler

```bash
# Test the production build locally with Cloudflare Pages environment
pnpm build
pnpm pages:dev
```

# FToolbox Project Overview

## Purpose
FToolbox is a full-stack web application that provides a collection of tools for Fansly creators and users. It tracks and analyzes Fansly tags and creators, providing statistics, rankings, and historical data.

## Tech Stack

### Frontend
- **Framework**: SvelteKit with Svelte 5
- **Styling**: Tailwind CSS v4
- **UI Components**: shadcn-svelte (pre-built components in `frontend/src/lib/components/ui/`)
- **Deployment**: Cloudflare Pages
- **Package Manager**: pnpm
- **Additional**: Chart.js for visualizations, mode-watcher for theme management

### Backend
- **Language**: Go 1.24+
- **Framework**: Fiber (web framework)
- **ORM**: GORM
- **Database**: MariaDB
- **Logging**: Zap
- **Hot Reload**: Air (for development)
- **Port**: 3000

### Task Runner
- **Tool**: Taskfile
- Commands available for starting frontend and backend dev servers

## Main Features
1. **Tag Management**: Track Fansly tags with view counts, rankings, and historical data
2. **Creator Management**: Track creators with subscriber counts and statistics
3. **Worker System**: Background jobs for updating data, discovering new tags/creators, and calculating rankings
4. **API**: RESTful API for frontend consumption
5. **Request Queue**: Users can request new tags/creators to be tracked

## External Integration
- **Fansly API**: Client with rate limiting and retry logic for fetching tag and creator data
#!/bin/sh
set -e

echo "========================================"
echo "FToolbox Development Environment Cleanup"
echo "========================================"

# Function to check if process is running
is_running() {
    pgrep -f "$1" > /dev/null 2>&1
}

# Stop development servers
echo "\n[1/5] Stopping development servers..."

# Function to kill processes on a port
kill_port() {
    local port=$1
    echo "Killing processes on port $port..."
    lsof -ti:$port | xargs kill -9 2>/dev/null || true
}

# Kill specific ports
kill_port 5173  # Frontend dev server
kill_port 3000  # Backend API server

# Kill any processes by name
echo "Stopping any Vite processes..."
pkill -f "vite" 2>/dev/null || true

echo "Stopping any Air processes..."
pkill -f "air" 2>/dev/null || true

echo "Stopping any Go processes..."
pkill -f "go run" 2>/dev/null || true
pkill -f "backend-go" 2>/dev/null || true

echo "Stopping any Node processes..."
pkill -f "node.*frontend" 2>/dev/null || true

echo "✓ Development servers stopped"

# Stop database if running (only if we have permission)
echo "\n[2/5] Checking database..."
if [ "$(id -u)" = "0" ]; then
    if is_running "mysqld"; then
        echo "Stopping MariaDB..."
        mysqladmin -u root shutdown 2>/dev/null || true
        echo "✓ MariaDB stopped"
    else
        echo "✓ MariaDB not running"
    fi
else
    echo "⚠️  Skipping database shutdown (requires root)"
fi

# Clean up temporary files
echo "\n[3/5] Cleaning up temporary files..."

# Remove database setup note if exists
if [ -f "database-setup-note.txt" ]; then
    rm -f database-setup-note.txt
    echo "✓ Removed database-setup-note.txt"
fi

# Clean frontend cache
if [ -d "frontend/node_modules/.cache" ]; then
    echo "Cleaning frontend cache..."
    rm -rf frontend/node_modules/.cache
    echo "✓ Frontend cache cleaned"
fi

# Clean backend tmp directory if exists
if [ -d "backend-go/tmp" ]; then
    echo "Cleaning backend tmp directory..."
    rm -rf backend-go/tmp
    echo "✓ Backend tmp cleaned"
fi

# Clean ESLint cache
if [ -f "frontend/.eslintcache" ]; then
    echo "Cleaning ESLint cache..."
    rm -f frontend/.eslintcache
    echo "✓ ESLint cache cleaned"
fi

# Clean Go build cache (optional, with flag)
if [ "$1" = "--deep" ]; then
    echo "\nPerforming deep cleanup..."
    
    # Clean Go module cache
    echo "Cleaning Go module cache..."
    go clean -modcache 2>/dev/null || true
    
    # Clean pnpm store
    echo "Cleaning pnpm store..."
    cd frontend && pnpm store prune 2>/dev/null || true && cd ..
    
    # Remove node_modules completely
    if [ -d "frontend/node_modules" ]; then
        echo "Removing frontend/node_modules..."
        rm -rf frontend/node_modules
        echo "✓ node_modules removed"
    fi
    
    # Clean build artifacts
    if [ -d "frontend/.svelte-kit" ]; then
        echo "Removing .svelte-kit build directory..."
        rm -rf frontend/.svelte-kit
        echo "✓ .svelte-kit removed"
    fi
    
    if [ -d "frontend/build" ]; then
        echo "Removing frontend build directory..."
        rm -rf frontend/build
        echo "✓ Frontend build removed"
    fi
fi

# Clean workspace-specific files
echo "\n[4/5] Cleaning workspace files..."

# Remove any generated .env files if they match defaults
if [ -f "backend-go/.env" ]; then
    if grep -q "DB_PASSWORD=mysql" backend-go/.env 2>/dev/null; then
        echo "Removing default backend-go/.env..."
        rm -f backend-go/.env
        echo "✓ Default .env removed"
    else
        echo "⚠️  Keeping customized backend-go/.env"
    fi
fi

if [ -f "frontend/.env" ]; then
    if grep -q "PUBLIC_API_URL=http://localhost:3000" frontend/.env 2>/dev/null; then
        echo "Removing default frontend/.env..."
        rm -f frontend/.env
        echo "✓ Default .env removed"
    else
        echo "⚠️  Keeping customized frontend/.env"
    fi
fi

# Summary
echo "\n[5/5] Summary..."

# Check if any dev processes are still running
PROCESSES_RUNNING=false
for process in "vite" "air" "task watch"; do
    if is_running "$process"; then
        echo "⚠️  Process still running: $process"
        PROCESSES_RUNNING=true
    fi
done

if [ "$PROCESSES_RUNNING" = "false" ]; then
    echo "✓ All development processes stopped"
fi

echo "\n========================================"
echo "Cleanup complete!"
echo "========================================"

echo "\nEnvironment has been cleaned up."
echo "Dev tools remain installed for future use."

if [ "$1" = "--deep" ]; then
    echo "\nDeep cleanup performed. Dependencies removed."
    echo "Run setup script to reinstall dependencies."
fi

echo "\nTo restart development:"
echo "  ./claude-code-github-bot-setup.sh"
echo "  task watch-frontend  # Frontend on port 5173"
echo "  task watch-backend   # Backend on port 3000"
echo "\nFor deep cleanup (remove all caches/dependencies):"
echo "  ./claude-code-github-bot-cleanup.sh --deep"
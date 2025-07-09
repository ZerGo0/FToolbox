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
echo "\n[1/4] Stopping development servers..."

# Kill any processes by name
echo "Stopping any Vite processes..."
pkill -f "vite" 2>/dev/null || true

echo "Stopping any Air processes..."
pkill -f "air" 2>/dev/null || true

echo "Stopping any Go processes on port 3000..."
pkill -f "port.*3000" 2>/dev/null || true
pkill -f ":3000" 2>/dev/null || true

echo "✓ Development servers stopped"

# Stop database if running (only if we have permission)
echo "\n[2/4] Checking database..."
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
echo "\n[3/4] Cleaning up temporary files..."

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

# Summary
echo "\n[4/4] Summary..."

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
echo "\nTo restart development:"
echo "  ./claude-code-github-bot-setup.sh"
echo "  task watch-frontend"
echo "  task watch-backend"
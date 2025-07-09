#!/bin/sh
set -e

echo "======================================"
echo "FToolbox Development Environment Setup"
echo "======================================"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check tools
echo "\n[1/6] Checking required tools..."
TOOLS_OK=true
for tool in go node npm pnpm task; do
    if command_exists $tool; then
        case $tool in
            go) echo "✓ Go $(go version | cut -d' ' -f3)" ;;
            node) echo "✓ Node.js $(node --version)" ;;
            npm) echo "✓ npm v$(npm --version 2>/dev/null | head -1)" ;;
            pnpm) echo "✓ pnpm v$(pnpm --version)" ;;
            task) echo "✓ Task v$(task --version | grep -o '[0-9.]*' | head -1)" ;;
        esac
    else
        echo "✗ $tool not found"
        TOOLS_OK=false
    fi
done

if [ "$TOOLS_OK" = "false" ]; then
    echo "\nSome required tools are missing. Please install them first."
    exit 1
fi

# Install Air (Go live reload)
echo "\n[2/6] Installing Air..."
if ! command_exists air; then
    echo "Installing Air for Go hot reload..."
    go install github.com/air-verse/air@latest
    echo "✓ Air installed"
else
    echo "✓ Air already installed"
fi

# Create .env files
echo "\n[3/6] Creating .env files..."

# Backend .env
if [ ! -f "backend-go/.env" ]; then
    if [ -f "backend-go/.env.example" ]; then
        cp backend-go/.env.example backend-go/.env
        echo "✓ Created backend-go/.env from .env.example"
    else
        cat > backend-go/.env <<'EOF'
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=mysql
DB_PASSWORD=mysql
DB_DATABASE=ftoolbox

# Server Configuration
PORT=3000
LOG_LEVEL=info

# Worker Configuration
WORKER_ENABLED=true
WORKER_UPDATE_INTERVAL=10000
WORKER_DISCOVERY_INTERVAL=60000
RANK_CALCULATION_INTERVAL=60000

# Fansly API Configuration
FANSLY_AUTH_TOKEN=
FANSLY_GLOBAL_RATE_LIMIT=5
FANSLY_GLOBAL_RATE_LIMIT_WINDOW=10
EOF
        echo "✓ Created backend-go/.env with defaults"
    fi
else
    echo "✓ backend-go/.env already exists"
fi

# Frontend .env
if [ ! -f "frontend/.env" ]; then
    if [ -f "frontend/.env.example" ]; then
        cp frontend/.env.example frontend/.env
        echo "✓ Created frontend/.env from .env.example"
    else
        cat > frontend/.env <<'EOF'
PUBLIC_API_URL=http://localhost:3000
EOF
        echo "✓ Created frontend/.env with defaults"
    fi
else
    echo "✓ frontend/.env already exists"
fi

# Install dependencies
echo "\n[4/6] Installing project dependencies..."

# Backend dependencies
echo "\nInstalling Go dependencies..."
cd backend-go
go mod download
cd ..
echo "✓ Go dependencies installed"

# Frontend dependencies
echo "\nInstalling frontend dependencies..."
cd frontend
pnpm install
cd ..
echo "✓ Frontend dependencies installed"

# Verify setup
echo "\n[5/6] Verifying setup..."

# Test backend commands
echo "\nTesting backend commands..."
cd backend-go
go fmt ./...
go vet ./...
cd ..
echo "✓ Backend commands working"

# Test frontend commands  
echo "\nTesting frontend commands..."
cd frontend
echo "Running pnpm check (may show existing type errors)..."
pnpm check || true
echo "\nRunning pnpm lint (may show existing lint issues)..."
pnpm lint || true
cd ..
echo "✓ Frontend commands available"

# Test task runner
echo "\nChecking task commands..."
if task --list >/dev/null 2>&1; then
    echo "✓ Task runner working"
    echo "\nAvailable development tasks:"
    task --list | grep -E "watch-frontend|watch-backend" || true
else
    echo "✗ Task runner not working properly"
fi

# Database setup note
echo "\n[6/6] Database Setup..."
if [ -f "/app/workspaces/bc70c60e-ef9a-4e5f-9038-cc8173f11f09/database-setup-note.txt" ]; then
    rm -f /app/workspaces/bc70c60e-ef9a-4e5f-9038-cc8173f11f09/database-setup-note.txt
fi

cat > database-setup-note.txt <<'EOF'
MariaDB/MySQL Setup Required:
=============================
The application requires a database connection.

Expected configuration (from .env files):
- Database: ftoolbox
- Username: mysql  
- Password: mysql
- Host: localhost
- Port: 3306

If you have Docker, you can run:
  docker run -d --name ftoolbox-db \
    -e MYSQL_ROOT_PASSWORD=root \
    -e MYSQL_DATABASE=ftoolbox \
    -e MYSQL_USER=mysql \
    -e MYSQL_PASSWORD=mysql \
    -p 3306:3306 \
    mariadb:latest

Or use any existing MariaDB/MySQL instance.
EOF

echo "⚠️  Database setup required - see database-setup-note.txt"

echo "\n======================================"
echo "Setup complete!"
echo "======================================"
echo "\nEnvironment ready. You can now run:"
echo "  task watch-frontend  # Start frontend dev server (port 5173)"
echo "  task watch-backend   # Start backend with live reload (port 3000)"
echo "\nNote: Ensure database is running before starting the backend."
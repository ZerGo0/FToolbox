#!/bin/sh
set -e

echo "======================================"
echo "FToolbox Development Environment Setup"
echo "======================================"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install system dependencies
echo "\n[1/8] Installing system dependencies..."
apk update
apk add --no-cache \
    curl \
    wget \
    git \
    bash \
    build-base \
    mariadb \
    mariadb-client \
    mariadb-dev

# Install Go 1.24.4
echo "\n[2/8] Installing Go 1.24.4..."
if ! command_exists go || ! go version | grep -q "go1.24.4"; then
    GO_VERSION="1.24.4"
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"
    export PATH="/usr/local/go/bin:$PATH"
    echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.profile
fi
echo "Go version: $(go version)"

# Install Node.js (LTS)
echo "\n[3/8] Installing Node.js..."
if ! command_exists node; then
    NODE_VERSION="20"
    curl -fsSL https://raw.githubusercontent.com/tj/n/master/bin/n | bash -s lts
    export PATH="/usr/local/bin:$PATH"
fi
echo "Node.js version: $(node --version)"

# Install pnpm
echo "\n[4/8] Installing pnpm..."
if ! command_exists pnpm; then
    npm install -g pnpm
fi
echo "pnpm version: $(pnpm --version)"

# Install Task runner
echo "\n[5/8] Installing Task runner..."
if ! command_exists task; then
    curl -sL https://taskfile.dev/install.sh | sh -s -- -d -b /usr/local/bin
fi
echo "Task version: $(task --version)"

# Install Air (Go live reload)
echo "\n[6/8] Installing Air..."
export GOPATH="/root/go"
export PATH="$GOPATH/bin:$PATH"
echo 'export GOPATH="/root/go"' >> ~/.profile
echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.profile
go install github.com/cosmtrek/air@latest
echo "Air installed at: $(which air)"

# Setup MariaDB
echo "\n[7/8] Setting up MariaDB..."
if [ ! -d "/run/mysqld" ]; then
    mkdir -p /run/mysqld
    chown mysql:mysql /run/mysqld
fi

# Initialize MariaDB if needed
if [ ! -d "/var/lib/mysql/mysql" ]; then
    mysql_install_db --user=mysql --ldata=/var/lib/mysql
fi

# Start MariaDB
echo "Starting MariaDB..."
mysqld_safe --user=mysql --datadir=/var/lib/mysql &
sleep 5

# Create database and user
mysql -u root <<EOF || true
CREATE DATABASE IF NOT EXISTS ftoolbox;
CREATE USER IF NOT EXISTS 'mysql'@'localhost' IDENTIFIED BY 'mysql';
GRANT ALL PRIVILEGES ON ftoolbox.* TO 'mysql'@'localhost';
FLUSH PRIVILEGES;
EOF

# Create .env files
echo "\n[8/8] Creating .env files..."

# Backend .env
if [ ! -f "backend-go/.env" ]; then
    cat > backend-go/.env <<EOF
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
    echo "Created backend-go/.env"
fi

# Frontend .env
if [ ! -f "frontend/.env" ]; then
    cat > frontend/.env <<EOF
PUBLIC_API_URL=http://localhost:3000
EOF
    echo "Created frontend/.env"
fi

# Install dependencies
echo "\n======================================"
echo "Installing project dependencies..."
echo "======================================"

# Backend dependencies
echo "\n[Backend] Installing Go dependencies..."
cd backend-go
go mod download
cd ..

# Frontend dependencies
echo "\n[Frontend] Installing pnpm dependencies..."
cd frontend
pnpm install
cd ..

# Verify setup
echo "\n======================================"
echo "Verifying setup..."
echo "======================================"

# Test backend commands
echo "\n[Backend] Testing Go commands..."
cd backend-go
go fmt ./...
go vet ./...
cd ..
echo "✓ Backend commands working"

# Test frontend commands  
echo "\n[Frontend] Testing pnpm commands..."
cd frontend
pnpm check || echo "Note: pnpm check may fail if there are type errors in the codebase"
pnpm lint || echo "Note: pnpm lint may fail if there are linting issues in the codebase"
cd ..
echo "✓ Frontend commands available"

# Test task runner
echo "\n[Task Runner] Checking task availability..."
if task --list >/dev/null 2>&1; then
    echo "✓ Task runner working"
    echo "Available tasks:"
    task --list | grep -E "watch-frontend|watch-backend" || true
else
    echo "✗ Task runner not working properly"
fi

echo "\n======================================"
echo "Setup complete!"
echo "======================================"
echo "\nEnvironment ready with:"
echo "- Go $(go version | cut -d' ' -f3)"
echo "- Node.js $(node --version)"
echo "- pnpm $(pnpm --version)"
echo "- Task $(task --version | head -1)"
echo "- Air (Go live reload)"
echo "- MariaDB running on localhost:3306"
echo "\nYou can now run:"
echo "  task watch-frontend  # Start frontend dev server"
echo "  task watch-backend   # Start backend with live reload"
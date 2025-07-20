#!/bin/sh
set -e

echo "======================================"
echo "FToolbox Development Environment Setup"
echo "======================================"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install missing tools first
echo "\n[1/7] Installing missing development tools..."

# Install pnpm if not present
if ! command_exists pnpm; then
    echo "Installing pnpm..."
    npm install -g pnpm
    echo "✓ pnpm installed"
fi

# Install task if not present
if ! command_exists task; then
    echo "Installing Task (Taskfile runner)..."
    curl -sL https://taskfile.dev/install.sh | sh
    # Move to a directory in PATH (try with sudo first, fallback to user directory)
    if sudo mv ./bin/task /usr/local/bin/task 2>/dev/null; then
        echo "✓ Task installed globally"
    else
        # Fallback to user directory
        mkdir -p "$HOME/.local/bin"
        mv ./bin/task "$HOME/.local/bin/task"
        export PATH="$HOME/.local/bin:$PATH"
        echo "✓ Task installed in user directory"
    fi
fi

# Check all required tools
echo "\n[2/7] Verifying all required tools..."
TOOLS_OK=true
for tool in go node npm pnpm task bun; do
    if command_exists $tool; then
        case $tool in
            go) echo "✓ Go $(go version | cut -d' ' -f3)" ;;
            node) echo "✓ Node.js $(node --version)" ;;
            npm) echo "✓ npm v$(npm --version 2>/dev/null | head -1)" ;;
            pnpm) echo "✓ pnpm v$(pnpm --version)" ;;
            task) echo "✓ Task v$(task --version | grep -o '[0-9.]*' | head -1)" ;;
            bun) echo "✓ Bun v$(bun --version)" ;;
        esac
    else
        echo "✗ $tool not found"
        TOOLS_OK=false
    fi
done

if [ "$TOOLS_OK" = "false" ]; then
    echo "\nSome required tools are still missing after installation attempts."
    exit 1
fi

# Install Air (Go live reload)
echo "\n[3/8] Installing Air..."
if ! command_exists air; then
    echo "Installing Air for Go hot reload..."
    go install github.com/air-verse/air@latest
    echo "✓ Air installed"
else
    echo "✓ Air already installed"
fi

# Create .env files
echo "\n[4/8] Creating .env files..."

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
echo "\n[5/8] Installing project dependencies..."

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

# Run mandatory checks
echo "\n[6/8] Running mandatory checks from CLAUDE.md..."

# Backend mandatory checks (from CLAUDE.md)
echo "\nRunning backend mandatory checks..."
cd backend-go
echo "Running go fmt ./..."
go fmt ./...
echo "Running go vet ./..."
go vet ./...
cd ..
echo "✓ Backend checks passed"

# Frontend mandatory checks (from CLAUDE.md)
echo "\nRunning frontend mandatory checks..."
cd frontend
echo "Running pnpm check && pnpm lint..."
if pnpm check && pnpm lint; then
    echo "✓ Frontend checks passed"
else
    echo "⚠️  Frontend checks failed - please fix errors before committing"
    exit 1
fi
cd ..

# Bot mandatory checks (if in bot project)
echo "\n[7/8] Checking for bot project..."
if [ -f "../../package.json" ] && command_exists bun; then
    echo "Running bot mandatory checks (bun tsc && bun lint)..."
    cd ../..
    if [ -f "tsconfig.json" ]; then
        echo "Running bun tsc..."
        bun tsc || echo "⚠️  Bot TypeScript check failed"
    fi
    if [ -f ".eslintrc.json" ] || [ -f "eslint.config.js" ]; then
        echo "Running bun lint..."
        bun lint || echo "⚠️  Bot lint check failed"
    fi
    cd workspaces/*/ 2>/dev/null || cd -
else
    echo "✓ Not in bot project workspace - skipping bot checks"
fi

# Verify setup
echo "\n[8/8] Final verification..."

# Test task runner
echo "\nChecking task commands..."
if task --list >/dev/null 2>&1; then
    echo "✓ Task runner working"
    echo "\nAvailable development tasks:"
    task --list | grep -E "watch-frontend|watch-backend" || true
else
    echo "✗ Task runner not working properly"
fi

echo "\n======================================"
echo "Setup complete!"
echo "======================================"
echo "\n⚠️  IMPORTANT REMINDERS FROM CLAUDE.md:"
echo "  Frontend: ALWAYS run 'pnpm check && pnpm lint' after changes"
echo "  Backend:  ALWAYS run 'go fmt ./...' and 'go vet ./...' after changes"
echo "  Bot:      ALWAYS run 'bun tsc && bun lint' after changes (if applicable)"
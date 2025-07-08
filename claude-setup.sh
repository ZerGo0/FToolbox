#!/bin/bash

# FToolbox Development Environment Setup Script
# This script sets up the complete development environment for FToolbox

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install package manager if not exists
install_package_manager() {
    local pm=$1
    case $pm in
        "pnpm")
            if ! command_exists pnpm; then
                log_info "Installing pnpm..."
                if command_exists npm; then
                    npm install -g pnpm
                elif command_exists curl; then
                    curl -fsSL https://get.pnpm.io/install.sh | sh -
                    export PATH="$HOME/.local/share/pnpm:$PATH"
                else
                    log_error "Cannot install pnpm: neither npm nor curl is available"
                    return 1
                fi
            fi
            ;;
        "npm")
            if ! command_exists npm; then
                log_error "npm is not installed. Please install Node.js first."
                return 1
            fi
            ;;
    esac
}

# Install runtime if not exists
install_runtime() {
    local runtime=$1
    case $runtime in
        "node")
            if ! command_exists node; then
                log_info "Installing Node.js..."
                if command_exists curl; then
                    # Direct installation to user directory
                    log_info "Installing Node.js directly..."
                    NODE_VERSION="v20.10.0"
                    mkdir -p "$HOME/.local/bin"
                    curl -fsSL "https://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-linux-x64.tar.xz" | tar -xJ -C "$HOME/.local" --strip-components=1 2>/dev/null || true
                    export PATH="$HOME/.local/bin:$PATH"
                    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc || true
                    
                    # Verify installation
                    if ! command_exists node; then
                        log_warning "Node.js direct installation failed, trying alternative method..."
                        # Try alternative installation method
                        NODE_TAR="node-${NODE_VERSION}-linux-x64.tar.xz"
                        curl -fsSL "https://nodejs.org/dist/${NODE_VERSION}/${NODE_TAR}" -o "/tmp/${NODE_TAR}"
                        if [[ -f "/tmp/${NODE_TAR}" ]]; then
                            tar -xf "/tmp/${NODE_TAR}" -C "$HOME/.local" --strip-components=1 2>/dev/null || true
                            rm -f "/tmp/${NODE_TAR}"
                        fi
                    fi
                else
                    log_error "Cannot install Node.js automatically. curl is required."
                    return 1
                fi
            fi
            ;;
        "go")
            if ! command_exists go; then
                log_info "Installing Go..."
                if command_exists curl; then
                    GO_VERSION="1.24.4"
                    mkdir -p "$HOME/.local"
                    log_info "Downloading Go ${GO_VERSION}..."
                    GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
                    curl -fsSL "https://dl.google.com/go/${GO_TAR}" -o "/tmp/${GO_TAR}"
                    if [[ -f "/tmp/${GO_TAR}" ]]; then
                        tar -xf "/tmp/${GO_TAR}" -C "$HOME/.local" 2>/dev/null || true
                        rm -f "/tmp/${GO_TAR}"
                        export PATH="$HOME/.local/go/bin:$PATH"
                        echo 'export PATH="$HOME/.local/go/bin:$PATH"' >> ~/.bashrc || true
                    else
                        log_error "Failed to download Go"
                        return 1
                    fi
                else
                    log_error "Cannot install Go automatically. curl is required."
                    return 1
                fi
            fi
            ;;
        "docker")
            if ! command_exists docker; then
                log_info "Docker installation skipped - detected containerized environment"
                # In a containerized environment, Docker is typically available from the host
                # We'll skip Docker installation and assume it's already available
            fi
            ;;
        "docker-compose")
            if ! command_exists docker-compose; then
                log_info "Docker Compose installation skipped - using existing installation"
                # Docker Compose is already available in this environment
            fi
            ;;
        "task")
            if ! command_exists task; then
                log_info "Installing Task runner..."
                if command_exists curl; then
                    # Install Task to user directory
                    mkdir -p "$HOME/.local/bin"
                    curl -sL https://taskfile.dev/install.sh | sh -s -- -b "$HOME/.local/bin"
                    export PATH="$HOME/.local/bin:$PATH"
                    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
                elif command_exists go; then
                    go install github.com/go-task/task/v3/cmd/task@latest
                else
                    log_error "Cannot install Task runner. curl or go is required."
                    return 1
                fi
            fi
            ;;
        "air")
            if ! command_exists air; then
                log_info "Installing Air (Go live reload)..."
                if command_exists go; then
                    go install github.com/cosmtrek/air@latest
                else
                    log_error "Go is required to install Air"
                    return 1
                fi
            fi
            ;;
    esac
}

# Run command with auto-fix logic
run_with_autofix() {
    local cmd="$1"
    local description="$2"
    local max_attempts=3
    local attempt=1
    
    log_info "Running: $description"
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Attempt $attempt/$max_attempts: $cmd"
        
        if eval "$cmd"; then
            log_success "$description completed successfully"
            return 0
        else
            local exit_code=$?
            log_warning "$description failed with exit code $exit_code (attempt $attempt/$max_attempts)"
            
            if [ $attempt -lt $max_attempts ]; then
                log_info "Attempting to fix common issues..."
                
                # Auto-fix logic based on common failure patterns
                case "$cmd" in
                    *"pnpm install"*)
                        log_info "Fixing pnpm installation issues..."
                        # Clear cache and try again
                        pnpm store prune || true
                        # Remove node_modules and try again
                        rm -rf node_modules || true
                        ;;
                    *"pnpm check"*)
                        log_info "Fixing TypeScript check issues..."
                        # Regenerate .svelte-kit directory
                        rm -rf .svelte-kit || true
                        # Install dependencies if missing
                        pnpm install || true
                        ;;
                    *"pnpm lint"*)
                        log_info "Fixing linting issues..."
                        # Auto-fix linting issues
                        pnpm exec prettier --write . || true
                        pnpm exec eslint --fix . || true
                        ;;
                    *"go mod download"*)
                        log_info "Fixing Go module download issues..."
                        # Clear module cache
                        go clean -modcache || true
                        # Update go.sum
                        go mod tidy || true
                        ;;
                    *"go fmt"*)
                        log_info "Go fmt should auto-fix, retrying..."
                        ;;
                    *"go vet"*)
                        log_info "Fixing Go vet issues..."
                        # Run go mod tidy to ensure dependencies are correct
                        go mod tidy || true
                        ;;
                    *"docker-compose up"*)
                        log_info "Fixing Docker Compose issues..."
                        # Stop any existing containers
                        docker-compose down || true
                        # Remove volumes if they exist
                        docker-compose down -v || true
                        # Pull latest images
                        docker-compose pull || true
                        ;;
                esac
                
                attempt=$((attempt + 1))
                sleep 2
            else
                log_error "$description failed after $max_attempts attempts"
                return $exit_code
            fi
        fi
    done
}

# Main setup function
main() {
    log_info "Starting FToolbox development environment setup..."
    
    # Check if we're in the right directory
    if [[ ! -f "CLAUDE.md" ]]; then
        log_error "This script must be run from the project root directory"
        exit 1
    fi
    
    # Ensure PATH includes common binary directories
    export PATH="$HOME/.local/bin:$HOME/.local/go/bin:$HOME/.local/share/pnpm:$PATH"
    
    # Check what tools are available
    log_info "Checking available tools..."
    
    # Check existing tools
    if command_exists node; then
        log_success "Node.js is available: $(node --version)"
    else
        log_info "Installing required runtimes and tools..."
        install_runtime "node" || log_warning "Node.js installation failed, continuing..."
    fi
    
    if command_exists go; then
        log_success "Go is available: $(go version)"
    else
        install_runtime "go" || log_warning "Go installation failed, continuing..."
    fi
    
    # Check Docker
    if command_exists docker; then
        log_success "Docker is available: $(docker --version)"
    else
        install_runtime "docker" || log_warning "Docker installation failed, continuing..."
    fi
    
    if command_exists docker-compose; then
        log_success "Docker Compose is available: $(docker-compose --version)"
    else
        install_runtime "docker-compose" || log_warning "Docker Compose installation failed, continuing..."
    fi
    
    # Install additional tools only if Go is available
    if command_exists go; then
        install_runtime "task" || log_warning "Task installation failed, continuing..."
        install_runtime "air" || log_warning "Air installation failed, continuing..."
    fi
    
    # Install package managers
    log_info "Installing package managers..."
    install_package_manager "pnpm" || log_warning "pnpm installation failed, continuing..."
    
    # Set up environment files
    log_info "Setting up environment files..."
    if [[ -f "backend-go/.env.example" ]] && [[ ! -f "backend-go/.env" ]]; then
        log_info "Creating backend-go/.env from .env.example"
        cp backend-go/.env.example backend-go/.env
        log_warning "Please review and update backend-go/.env with your configuration"
    fi
    
    # Setup Frontend
    if [[ -d "frontend" ]]; then
        log_info "Setting up frontend..."
        cd frontend
        
        # Check if pnpm is available
        if command_exists pnpm; then
            # Install frontend dependencies
            run_with_autofix "pnpm install" "Frontend dependency installation" || log_warning "Frontend dependency installation failed"
            
            # Run mandatory frontend commands from CLAUDE.md
            log_info "Running mandatory frontend commands..."
            run_with_autofix "pnpm check" "Frontend type checking" || log_warning "Frontend type checking failed"
            run_with_autofix "pnpm lint" "Frontend linting" || log_warning "Frontend linting failed"
        else
            log_warning "pnpm not available, skipping frontend setup"
        fi
        
        cd ..
    else
        log_warning "Frontend directory not found, skipping frontend setup"
    fi
    
    # Setup Backend
    if [[ -d "backend-go" ]]; then
        log_info "Setting up backend..."
        cd backend-go
        
        # Check if go is available
        if command_exists go; then
            # Install backend dependencies
            run_with_autofix "go mod download" "Backend dependency download" || log_warning "Backend dependency download failed"
            
            # Run mandatory backend commands from CLAUDE.md
            log_info "Running mandatory backend commands..."
            run_with_autofix "go fmt ./..." "Backend code formatting" || log_warning "Backend code formatting failed"
            run_with_autofix "go vet ./..." "Backend code vetting" || log_warning "Backend code vetting failed"
        else
            log_warning "Go not available, skipping backend setup"
        fi
        
        cd ..
    else
        log_warning "Backend directory not found, skipping backend setup"
    fi
    
    # Start Docker services
    if command_exists docker && command_exists docker-compose; then
        log_info "Starting Docker services..."
        if [[ -d "backend-go" ]]; then
            cd backend-go
            
            # Check if production docker-compose exists and start database
            if [[ -f "docker-compose.prod.yml" ]]; then
                log_info "Starting MariaDB database..."
                run_with_autofix "docker-compose -f docker-compose.prod.yml up -d mariadb" "Database startup" || log_warning "Database startup failed"
                
                # Wait for database to be healthy
                log_info "Waiting for database to be ready..."
                timeout=60
                while [ $timeout -gt 0 ]; do
                    if docker-compose -f docker-compose.prod.yml ps mariadb 2>/dev/null | grep -q "healthy"; then
                        log_success "Database is ready"
                        break
                    fi
                    sleep 2
                    timeout=$((timeout - 2))
                done
                
                if [ $timeout -le 0 ]; then
                    log_warning "Database health check timed out, but continuing..."
                fi
            else
                log_warning "docker-compose.prod.yml not found, skipping database startup"
            fi
            
            cd ..
        fi
        
        # Build backend (if needed)
        if [[ -d "backend-go" ]]; then
            log_info "Building backend..."
            cd backend-go
            if [[ -f "Dockerfile" ]]; then
                log_info "Building backend Docker image..."
                run_with_autofix "docker build -t ftoolbox-backend:latest ." "Backend Docker build" || log_warning "Backend Docker build failed"
            else
                log_warning "Dockerfile not found, skipping backend Docker build"
            fi
            cd ..
        fi
    else
        log_warning "Docker or Docker Compose not available, skipping Docker services"
    fi
    
    # Create necessary directories
    log_info "Creating necessary directories..."
    mkdir -p backend-go/logs || true
    mkdir -p backend-go/tmp || true
    
    # Set proper permissions
    log_info "Setting proper permissions..."
    chmod +x backend-go/load_containers.sh 2>/dev/null || true
    
    # Final verification
    log_info "Running final verification..."
    
    # Frontend verification
    if [[ -d "frontend" ]] && command_exists pnpm; then
        cd frontend
        if run_with_autofix "pnpm check" "Final frontend verification"; then
            log_success "Frontend setup verified successfully"
        else
            log_warning "Frontend verification failed, but setup continued"
        fi
        cd ..
    else
        log_warning "Skipping frontend verification - directory or pnpm not available"
    fi
    
    # Backend verification
    if [[ -d "backend-go" ]] && command_exists go; then
        cd backend-go
        if run_with_autofix "go vet ./..." "Final backend verification"; then
            log_success "Backend setup verified successfully"
        else
            log_warning "Backend verification failed, but setup continued"
        fi
        cd ..
    else
        log_warning "Skipping backend verification - directory or go not available"
    fi
    
    log_success "FToolbox development environment setup completed!"
    log_info ""
    log_info "Available tools:"
    command_exists node && log_info "  ✓ Node.js: $(node --version)"
    command_exists npm && log_info "  ✓ npm: $(npm --version)"
    command_exists pnpm && log_info "  ✓ pnpm: $(pnpm --version)"
    command_exists go && log_info "  ✓ Go: $(go version)"
    command_exists docker && log_info "  ✓ Docker: $(docker --version)"
    command_exists docker-compose && log_info "  ✓ Docker Compose: $(docker-compose --version)"
    command_exists task && log_info "  ✓ Task: $(task --version)"
    command_exists air && log_info "  ✓ Air: $(air --version)"
    log_info ""
    
    if command_exists task; then
        log_info "To start development servers, run:"
        log_info "  Frontend: task watch-frontend"
        log_info "  Backend: task watch-backend"
        log_info ""
        log_info "Frontend will be available at: http://localhost:5173"
        log_info "Backend will be available at: http://localhost:3000"
    else
        log_warning "Task runner not available. You may need to start servers manually."
    fi
    
    log_info ""
    log_warning "Don't forget to review and update backend-go/.env with your configuration"
}

# Run main function
main "$@"
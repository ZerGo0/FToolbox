#!/bin/bash

# FToolbox Development Environment Setup Script
# This script sets up the complete development environment for FToolbox

set -e  # Exit on error

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
                    curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
                    sudo apt-get install -y nodejs
                elif command_exists brew; then
                    brew install node
                else
                    log_error "Cannot install Node.js automatically. Please install manually."
                    return 1
                fi
            fi
            ;;
        "go")
            if ! command_exists go; then
                log_info "Installing Go..."
                if command_exists curl; then
                    GO_VERSION="1.24.4"
                    curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" | sudo tar -C /usr/local -xzf -
                    export PATH="/usr/local/go/bin:$PATH"
                    echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
                elif command_exists brew; then
                    brew install go
                else
                    log_error "Cannot install Go automatically. Please install manually."
                    return 1
                fi
            fi
            ;;
        "docker")
            if ! command_exists docker; then
                log_info "Installing Docker..."
                if command_exists curl; then
                    curl -fsSL https://get.docker.com -o get-docker.sh
                    sh get-docker.sh
                    sudo usermod -aG docker $USER
                    rm get-docker.sh
                    log_warning "Docker installed. You may need to log out and back in for group membership to take effect."
                else
                    log_error "Cannot install Docker automatically. Please install manually."
                    return 1
                fi
            fi
            ;;
        "docker-compose")
            if ! command_exists docker-compose; then
                log_info "Installing Docker Compose..."
                if command_exists curl; then
                    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
                    sudo chmod +x /usr/local/bin/docker-compose
                elif command_exists pip; then
                    pip install docker-compose
                else
                    log_error "Cannot install Docker Compose automatically. Please install manually."
                    return 1
                fi
            fi
            ;;
        "task")
            if ! command_exists task; then
                log_info "Installing Task runner..."
                if command_exists curl; then
                    sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d
                elif command_exists brew; then
                    brew install go-task/tap/go-task
                else
                    log_error "Cannot install Task runner automatically. Please install manually."
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
    
    # Install required runtimes and tools
    log_info "Installing required runtimes and tools..."
    install_runtime "node"
    install_runtime "go"
    install_runtime "docker"
    install_runtime "docker-compose"
    install_runtime "task"
    install_runtime "air"
    
    # Install package managers
    log_info "Installing package managers..."
    install_package_manager "pnpm"
    
    # Set up environment files
    log_info "Setting up environment files..."
    if [[ -f "backend-go/.env.example" ]] && [[ ! -f "backend-go/.env" ]]; then
        log_info "Creating backend-go/.env from .env.example"
        cp backend-go/.env.example backend-go/.env
        log_warning "Please review and update backend-go/.env with your configuration"
    fi
    
    # Setup Frontend
    log_info "Setting up frontend..."
    cd frontend
    
    # Install frontend dependencies
    run_with_autofix "pnpm install" "Frontend dependency installation"
    
    # Run mandatory frontend commands from CLAUDE.md
    log_info "Running mandatory frontend commands..."
    run_with_autofix "pnpm check" "Frontend type checking"
    run_with_autofix "pnpm lint" "Frontend linting"
    
    cd ..
    
    # Setup Backend
    log_info "Setting up backend..."
    cd backend-go
    
    # Install backend dependencies
    run_with_autofix "go mod download" "Backend dependency download"
    
    # Run mandatory backend commands from CLAUDE.md
    log_info "Running mandatory backend commands..."
    run_with_autofix "go fmt ./..." "Backend code formatting"
    run_with_autofix "go vet ./..." "Backend code vetting"
    
    cd ..
    
    # Start Docker services
    log_info "Starting Docker services..."
    cd backend-go
    
    # Check if production docker-compose exists and start database
    if [[ -f "docker-compose.prod.yml" ]]; then
        log_info "Starting MariaDB database..."
        run_with_autofix "docker-compose -f docker-compose.prod.yml up -d mariadb" "Database startup"
        
        # Wait for database to be healthy
        log_info "Waiting for database to be ready..."
        timeout=60
        while [ $timeout -gt 0 ]; do
            if docker-compose -f docker-compose.prod.yml ps mariadb | grep -q "healthy"; then
                log_success "Database is ready"
                break
            fi
            sleep 2
            timeout=$((timeout - 2))
        done
        
        if [ $timeout -le 0 ]; then
            log_warning "Database health check timed out, but continuing..."
        fi
    fi
    
    cd ..
    
    # Build backend (if needed)
    log_info "Building backend..."
    cd backend-go
    if [[ -f "Dockerfile" ]]; then
        log_info "Building backend Docker image..."
        run_with_autofix "docker build -t ftoolbox-backend:latest ." "Backend Docker build"
    fi
    cd ..
    
    # Create necessary directories
    log_info "Creating necessary directories..."
    mkdir -p backend-go/logs || true
    mkdir -p backend-go/tmp || true
    
    # Set proper permissions
    log_info "Setting proper permissions..."
    chmod +x backend-go/load_containers.sh || true
    
    # Final verification
    log_info "Running final verification..."
    cd frontend
    if run_with_autofix "pnpm check" "Final frontend verification"; then
        log_success "Frontend setup verified successfully"
    else
        log_warning "Frontend verification failed, but setup continued"
    fi
    cd ..
    
    cd backend-go
    if run_with_autofix "go vet ./..." "Final backend verification"; then
        log_success "Backend setup verified successfully"
    else
        log_warning "Backend verification failed, but setup continued"
    fi
    cd ..
    
    log_success "FToolbox development environment setup completed!"
    log_info "To start development servers, run:"
    log_info "  Frontend: task watch-frontend"
    log_info "  Backend: task watch-backend"
    log_info ""
    log_info "Frontend will be available at: http://localhost:5173"
    log_info "Backend will be available at: http://localhost:3000"
    log_info ""
    log_warning "Don't forget to review and update backend-go/.env with your configuration"
}

# Run main function
main "$@"
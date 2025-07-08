#!/bin/bash

# FToolbox Development Environment Setup Script (Simple Version)
# This script sets up the environment using tools that are available or can be easily installed

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

# Main setup function
main() {
    log_info "Starting FToolbox development environment setup..."
    
    # Check if we're in the right directory
    if [[ ! -f "CLAUDE.md" ]]; then
        log_error "This script must be run from the project root directory"
        exit 1
    fi
    
    # Check what tools are available
    log_info "Checking available tools..."
    
    # Check existing tools
    if command_exists node; then
        log_success "Node.js is available: $(node --version)"
    else
        log_warning "Node.js is not available. You may need to install it manually."
    fi
    
    if command_exists npm; then
        log_success "npm is available: $(npm --version)"
    else
        log_warning "npm is not available."
    fi
    
    if command_exists go; then
        log_success "Go is available: $(go version)"
    else
        log_warning "Go is not available. You may need to install it manually."
    fi
    
    # Check Docker
    if command_exists docker; then
        log_success "Docker is available: $(docker --version)"
    else
        log_warning "Docker is not available."
    fi
    
    if command_exists docker-compose; then
        log_success "Docker Compose is available: $(docker-compose --version)"
    else
        log_warning "Docker Compose is not available."
    fi
    
    # Install pnpm using npm if available
    if command_exists npm && ! command_exists pnpm; then
        log_info "Installing pnpm..."
        npm install -g pnpm || log_warning "pnpm installation failed"
    fi
    
    if command_exists pnpm; then
        log_success "pnpm is available: $(pnpm --version)"
    else
        log_warning "pnpm is not available."
    fi
    
    # Set up environment files
    log_info "Setting up environment files..."
    if [[ -f "backend-go/.env.example" ]] && [[ ! -f "backend-go/.env" ]]; then
        log_info "Creating backend-go/.env from .env.example"
        cp backend-go/.env.example backend-go/.env
        log_warning "Please review and update backend-go/.env with your configuration"
    fi
    
    # Setup Frontend (if tools are available)
    if [[ -d "frontend" ]] && command_exists pnpm; then
        log_info "Setting up frontend..."
        cd frontend
        
        # Install frontend dependencies
        log_info "Installing frontend dependencies..."
        pnpm install || log_warning "Frontend dependency installation failed"
        
        # Run mandatory frontend commands from CLAUDE.md
        log_info "Running frontend checks..."
        pnpm check || log_warning "Frontend type checking failed"
        pnpm lint || log_warning "Frontend linting failed"
        
        cd ..
    else
        log_warning "Skipping frontend setup - directory or pnpm not available"
    fi
    
    # Setup Backend (if tools are available)
    if [[ -d "backend-go" ]] && command_exists go; then
        log_info "Setting up backend..."
        cd backend-go
        
        # Install backend dependencies
        log_info "Installing backend dependencies..."
        go mod download || log_warning "Backend dependency download failed"
        
        # Run mandatory backend commands from CLAUDE.md
        log_info "Running backend checks..."
        go fmt ./... || log_warning "Backend code formatting failed"
        go vet ./... || log_warning "Backend code vetting failed"
        
        cd ..
    else
        log_warning "Skipping backend setup - directory or go not available"
    fi
    
    # Start Docker services (if available)
    if command_exists docker && command_exists docker-compose; then
        log_info "Starting Docker services..."
        if [[ -d "backend-go" ]]; then
            cd backend-go
            
            # Check if production docker-compose exists and start database
            if [[ -f "docker-compose.prod.yml" ]]; then
                log_info "Starting MariaDB database..."
                docker-compose -f docker-compose.prod.yml up -d mariadb || log_warning "Database startup failed"
            else
                log_warning "docker-compose.prod.yml not found, skipping database startup"
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
    
    log_success "FToolbox development environment setup completed!"
    log_info ""
    log_info "Available tools:"
    command_exists node && log_info "  ✓ Node.js: $(node --version)"
    command_exists npm && log_info "  ✓ npm: $(npm --version)"
    command_exists pnpm && log_info "  ✓ pnpm: $(pnpm --version)"
    command_exists go && log_info "  ✓ Go: $(go version)"
    command_exists docker && log_info "  ✓ Docker: $(docker --version)"
    command_exists docker-compose && log_info "  ✓ Docker Compose: $(docker-compose --version)"
    log_info ""
    
    if command_exists pnpm && command_exists go; then
        log_info "Development environment is ready!"
        log_info "You can now start the development servers manually:"
        log_info "  Frontend: cd frontend && pnpm dev"
        log_info "  Backend: cd backend-go && go run main.go"
    else
        log_warning "Some tools are missing. Please install them manually to complete the setup."
    fi
    
    log_info ""
    log_warning "Don't forget to review and update backend-go/.env with your configuration"
}

# Run main function
main "$@"
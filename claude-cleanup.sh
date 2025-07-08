#!/bin/bash

# FToolbox Development Environment Cleanup Script
# This script tears down the development environment and cleans up resources

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

# Kill processes on specific ports
kill_port() {
    local port=$1
    local process_name=$2
    
    if command_exists lsof; then
        local pid=$(lsof -ti:$port 2>/dev/null | head -1 || true)
        if [[ -n "$pid" && "$pid" =~ ^[0-9]+$ ]]; then
            log_info "Killing $process_name process on port $port (PID: $pid)"
            kill -9 $pid 2>/dev/null || true
            log_success "$process_name process killed"
        else
            log_info "No process found on port $port"
        fi
    elif command_exists netstat; then
        local pid=$(netstat -tlnp 2>/dev/null | grep ":$port " | awk '{print $7}' | cut -d/ -f1 | head -1 || true)
        if [[ -n "$pid" && "$pid" != "-" && "$pid" =~ ^[0-9]+$ ]]; then
            log_info "Killing $process_name process on port $port (PID: $pid)"
            kill -9 $pid 2>/dev/null || true
            log_success "$process_name process killed"
        else
            log_info "No process found on port $port"
        fi
    elif command_exists ss; then
        # Alternative using ss command if netstat is not available
        local pid=$(ss -tlnp 2>/dev/null | grep ":$port " | sed 's/.*pid=\([0-9]*\).*/\1/' | head -1 || true)
        if [[ -n "$pid" && "$pid" != "-" && "$pid" =~ ^[0-9]+$ ]]; then
            log_info "Killing $process_name process on port $port (PID: $pid)"
            kill -9 $pid 2>/dev/null || true
            log_success "$process_name process killed"
        else
            log_info "No process found on port $port"
        fi
    else
        log_warning "Cannot check port $port - neither lsof, netstat, nor ss available"
        # Try a simple approach - kill any process that might be using the port
        local processes=$(ps aux | grep ":$port" | grep -v grep | awk '{print $2}' || true)
        if [[ -n "$processes" ]]; then
            for p in $processes; do
                if [[ "$p" =~ ^[0-9]+$ ]]; then
                    log_info "Killing process $p that might be using port $port"
                    kill -9 $p 2>/dev/null || true
                fi
            done
        fi
    fi
}

# Clean up Docker containers and volumes
cleanup_docker() {
    if command_exists docker; then
        log_info "Cleaning up Docker containers and volumes..."
        
        # Stop and remove containers from docker-compose files
        if command_exists docker-compose; then
            if [[ -f "backend-go/docker-compose.yml" ]]; then
                log_info "Stopping development docker-compose services..."
                cd backend-go
                docker-compose down --remove-orphans 2>/dev/null || true
                docker-compose down -v --remove-orphans 2>/dev/null || true
                cd ..
            fi
            
            if [[ -f "backend-go/docker-compose.prod.yml" ]]; then
                log_info "Stopping production docker-compose services..."
                cd backend-go
                docker-compose -f docker-compose.prod.yml down --remove-orphans 2>/dev/null || true
                docker-compose -f docker-compose.prod.yml down -v --remove-orphans 2>/dev/null || true
                cd ..
            fi
        else
            log_warning "docker-compose not found, skipping compose cleanup"
        fi
        
        # Remove FToolbox-specific containers
        log_info "Removing FToolbox containers..."
        # Use alternative approach for BusyBox compatibility
        docker ps -a 2>/dev/null | grep "ftoolbox" | awk '{print $1}' | xargs -r docker rm -f 2>/dev/null || true
        docker ps -a 2>/dev/null | grep "backend-go" | awk '{print $1}' | xargs -r docker rm -f 2>/dev/null || true
        docker ps -a 2>/dev/null | grep "mariadb" | awk '{print $1}' | xargs -r docker rm -f 2>/dev/null || true
        
        # Remove FToolbox images
        log_info "Removing FToolbox images..."
        docker images 2>/dev/null | grep "ftoolbox-backend" | awk '{print $3}' | xargs -r docker rmi -f 2>/dev/null || true
        
        # Clean up dangling volumes
        log_info "Cleaning up dangling volumes..."
        docker volume prune -f 2>/dev/null || true
        
        log_success "Docker cleanup completed"
    else
        log_warning "Docker not found, skipping Docker cleanup"
    fi
}

# Clean up Node.js/pnpm artifacts
cleanup_frontend() {
    log_info "Cleaning up frontend artifacts..."
    
    if [[ -d "frontend" ]]; then
        cd frontend
        
        # Remove build artifacts
        log_info "Removing build artifacts..."
        rm -rf .svelte-kit build dist .vercel .cloudflare 2>/dev/null || true
        
        # Remove dependency directories
        log_info "Removing dependency directories..."
        rm -rf node_modules .pnpm-store 2>/dev/null || true
        
        # Clean pnpm cache if pnpm exists
        if command_exists pnpm; then
            log_info "Cleaning pnpm cache..."
            pnpm store prune 2>/dev/null || true
        fi
        
        # Remove temporary files
        log_info "Removing temporary files..."
        rm -rf .temp tmp .cache 2>/dev/null || true
        
        cd ..
        log_success "Frontend cleanup completed"
    else
        log_warning "Frontend directory not found"
    fi
}

# Clean up Go artifacts
cleanup_backend() {
    log_info "Cleaning up backend artifacts..."
    
    if [[ -d "backend-go" ]]; then
        cd backend-go
        
        # Remove Go build artifacts
        log_info "Removing Go build artifacts..."
        rm -rf bin tmp logs .air_tmp 2>/dev/null || true
        
        # Clean Go module cache (optional - commented out as it affects global cache)
        # if command_exists go; then
        #     log_info "Cleaning Go module cache..."
        #     go clean -modcache 2>/dev/null || true
        # fi
        
        # Remove temporary files
        log_info "Removing temporary files..."
        rm -rf .temp tmp .cache 2>/dev/null || true
        
        cd ..
        log_success "Backend cleanup completed"
    else
        log_warning "Backend directory not found"
    fi
}

# Clean up database files (if using local database)
cleanup_database() {
    log_info "Cleaning up database files..."
    
    if [[ -d "backend-go/mariadb_data" ]]; then
        log_warning "Removing MariaDB data directory..."
        rm -rf backend-go/mariadb_data 2>/dev/null || true
        log_success "MariaDB data directory removed"
    fi
    
    # Remove any SQLite files if they exist
    if command_exists find; then
        find . -name "*.db" -o -name "*.sqlite" -o -name "*.sqlite3" 2>/dev/null | while read -r file; do
            if [[ -n "$file" ]]; then
                log_info "Removing database file: $file"
                rm -f "$file" 2>/dev/null || true
            fi
        done
    else
        log_warning "find command not available, skipping SQLite file cleanup"
    fi
}

# Clean up environment files (optional)
cleanup_env_files() {
    local remove_env_files=$1
    
    if [[ "$remove_env_files" == "true" ]]; then
        log_info "Removing environment files..."
        
        if [[ -f "backend-go/.env" ]]; then
            log_warning "Removing backend-go/.env"
            rm -f backend-go/.env 2>/dev/null || true
        fi
        
        if [[ -f "frontend/.env" ]]; then
            log_warning "Removing frontend/.env"
            rm -f frontend/.env 2>/dev/null || true
        fi
        
        if [[ -f "frontend/.env.local" ]]; then
            log_warning "Removing frontend/.env.local"
            rm -f frontend/.env.local 2>/dev/null || true
        fi
        
        log_success "Environment files removed"
    else
        log_info "Keeping environment files (use --remove-env to remove them)"
    fi
}

# Clean up logs and temporary files
cleanup_logs_and_temp() {
    log_info "Cleaning up logs and temporary files..."
    
    # Remove log files
    if command_exists find; then
        find . -name "*.log" -type f 2>/dev/null | while read -r file; do
            if [[ -n "$file" ]]; then
                log_info "Removing log file: $file"
                rm -f "$file" 2>/dev/null || true
            fi
        done
        
        # Clean up OS-specific temporary files
        find . -name ".DS_Store" -type f -delete 2>/dev/null || true
        find . -name "Thumbs.db" -type f -delete 2>/dev/null || true
    else
        log_warning "find command not available, skipping log file cleanup"
    fi
    
    # Remove temporary directories
    rm -rf .temp tmp .cache 2>/dev/null || true
    
    log_success "Logs and temporary files cleaned"
}

# Main cleanup function
main() {
    log_info "Starting FToolbox development environment cleanup..."
    
    # Parse command line arguments
    local remove_env_files=false
    local full_cleanup=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --remove-env)
                remove_env_files=true
                shift
                ;;
            --full)
                full_cleanup=true
                remove_env_files=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --remove-env    Remove environment files (.env)"
                echo "  --full          Full cleanup including environment files"
                echo "  -h, --help      Show this help message"
                echo ""
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Check if we're in the right directory
    if [[ ! -f "CLAUDE.md" ]]; then
        log_error "This script must be run from the project root directory"
        exit 1
    fi
    
    # Kill development servers
    log_info "Stopping development servers..."
    kill_port 5173 "SvelteKit dev server"
    kill_port 3000 "Go backend server"
    kill_port 3001 "Go backend server (Air)"
    
    # Kill any Task processes
    if command_exists pkill; then
        log_info "Stopping Task processes..."
        pkill -f "task watch-" 2>/dev/null || true
    else
        log_warning "pkill not available, unable to stop Task processes automatically"
    fi
    
    # Clean up Docker
    cleanup_docker
    
    # Clean up frontend
    cleanup_frontend
    
    # Clean up backend
    cleanup_backend
    
    # Clean up database
    if [[ "$full_cleanup" == "true" ]]; then
        cleanup_database
    else
        log_info "Skipping database cleanup (use --full for complete cleanup)"
    fi
    
    # Clean up environment files
    cleanup_env_files $remove_env_files
    
    # Clean up logs and temporary files
    cleanup_logs_and_temp
    
    # Final Docker cleanup
    if command_exists docker && [[ "$full_cleanup" == "true" ]]; then
        log_info "Performing final Docker cleanup..."
        docker system prune -f 2>/dev/null || true
        log_success "Docker system pruned"
    fi
    
    log_success "FToolbox development environment cleanup completed!"
    
    if [[ "$full_cleanup" != "true" ]]; then
        log_info "For complete cleanup including database and Docker system, run:"
        log_info "  $0 --full"
    fi
    
    log_info "To set up the environment again, run:"
    log_info "  ./claude-setup.sh"
}

# Run main function
main "$@"
#!/bin/bash

# Production deployment script for Collaborative Bucket List
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DOCKER_COMPOSE_FILE="docker-compose.prod.yml"
BACKUP_DIR="./backups"
LOG_FILE="./deploy.log"

# Functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}" | tee -a "$LOG_FILE"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" | tee -a "$LOG_FILE"
    exit 1
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed"
    fi
    
    if ! docker info &> /dev/null; then
        error "Docker is not running"
    fi
    
    # Check if Docker Compose is available
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        error "Docker Compose is not installed"
    fi
    
    # Check if required files exist
    if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
        error "Docker Compose file not found: $DOCKER_COMPOSE_FILE"
    fi
    
    if [ ! -f "backend/.env" ]; then
        error "Backend environment file not found: backend/.env"
    fi
    
    log "Prerequisites check passed"
}

# Create backup
create_backup() {
    log "Creating backup..."
    
    mkdir -p "$BACKUP_DIR"
    BACKUP_NAME="backup-$(date +'%Y%m%d-%H%M%S')"
    
    # Export current database if running
    if docker-compose -f "$DOCKER_COMPOSE_FILE" ps | grep -q "Up"; then
        log "Exporting database..."
        # Add database backup logic here if using a database container
        # docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T postgres pg_dump -U postgres collaborative_bucket_list > "$BACKUP_DIR/$BACKUP_NAME.sql"
    fi
    
    # Backup current environment files
    cp backend/.env "$BACKUP_DIR/$BACKUP_NAME-backend.env" 2>/dev/null || true
    cp frontend/.env.local "$BACKUP_DIR/$BACKUP_NAME-frontend.env" 2>/dev/null || true
    
    log "Backup created: $BACKUP_NAME"
}

# Build images
build_images() {
    log "Building Docker images..."
    
    # Build backend
    log "Building backend image..."
    docker build -t collaborative-bucket-list-backend:latest ./backend
    
    # Build frontend
    log "Building frontend image..."
    docker build -t collaborative-bucket-list-frontend:latest ./frontend
    
    log "Images built successfully"
}

# Deploy application
deploy() {
    log "Deploying application..."
    
    # Stop existing containers
    log "Stopping existing containers..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" down --remove-orphans
    
    # Start new containers
    log "Starting new containers..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d
    
    # Wait for services to be healthy
    log "Waiting for services to be healthy..."
    sleep 30
    
    # Check health
    check_health
    
    log "Deployment completed successfully"
}

# Check application health
check_health() {
    log "Checking application health..."
    
    # Check backend health
    for i in {1..30}; do
        if curl -f http://localhost:8080/health &> /dev/null; then
            log "Backend is healthy"
            break
        fi
        if [ $i -eq 30 ]; then
            error "Backend health check failed"
        fi
        sleep 2
    done
    
    # Check frontend health
    for i in {1..30}; do
        if curl -f http://localhost:3000/health &> /dev/null; then
            log "Frontend is healthy"
            break
        fi
        if [ $i -eq 30 ]; then
            error "Frontend health check failed"
        fi
        sleep 2
    done
    
    log "All services are healthy"
}

# Cleanup old images
cleanup() {
    log "Cleaning up old Docker images..."
    
    # Remove dangling images
    docker image prune -f
    
    # Remove old backups (keep last 5)
    if [ -d "$BACKUP_DIR" ]; then
        ls -t "$BACKUP_DIR"/backup-* 2>/dev/null | tail -n +6 | xargs -r rm -f
    fi
    
    log "Cleanup completed"
}

# Rollback function
rollback() {
    warn "Rolling back deployment..."
    
    # Stop current containers
    docker-compose -f "$DOCKER_COMPOSE_FILE" down
    
    # Restore from backup if available
    LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/backup-*.env 2>/dev/null | head -n 1)
    if [ -n "$LATEST_BACKUP" ]; then
        BACKUP_PREFIX=$(basename "$LATEST_BACKUP" | sed 's/-backend\.env$//')
        cp "$BACKUP_DIR/$BACKUP_PREFIX-backend.env" backend/.env 2>/dev/null || true
        cp "$BACKUP_DIR/$BACKUP_PREFIX-frontend.env" frontend/.env.local 2>/dev/null || true
        log "Environment files restored from backup"
    fi
    
    # Start with previous configuration
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d
    
    warn "Rollback completed"
}

# Main deployment process
main() {
    log "Starting deployment process..."
    
    # Set trap for cleanup on error
    trap rollback ERR
    
    check_prerequisites
    create_backup
    build_images
    deploy
    cleanup
    
    log "Deployment process completed successfully!"
    log "Application is running at:"
    log "  Frontend: http://localhost:3000"
    log "  Backend API: http://localhost:8080"
    log "  Backend Health: http://localhost:8080/health"
    log "  Backend Metrics: http://localhost:9090/metrics"
}

# Handle command line arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "rollback")
        rollback
        ;;
    "health")
        check_health
        ;;
    "cleanup")
        cleanup
        ;;
    "backup")
        create_backup
        ;;
    *)
        echo "Usage: $0 {deploy|rollback|health|cleanup|backup}"
        echo "  deploy   - Full deployment process (default)"
        echo "  rollback - Rollback to previous version"
        echo "  health   - Check application health"
        echo "  cleanup  - Clean up old images and backups"
        echo "  backup   - Create backup only"
        exit 1
        ;;
esac
#!/bin/bash

# Production Deployment Script for Collaborative Bucket List
# This script handles the complete production deployment process

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="collaborative-bucket-list"
BACKEND_DIR="backend"
FRONTEND_DIR="frontend"
DOCKER_COMPOSE_PROD="docker-compose.prod.yml"
DOCKER_COMPOSE_MONITORING="docker-compose.monitoring.yml"

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if required tools are installed
check_requirements() {
    log "Checking requirements..."
    
    local missing_tools=()
    
    for tool in docker docker-compose git make; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
    
    success "All requirements satisfied"
}

# Validate environment files
validate_environment() {
    log "Validating environment configuration..."
    
    # Check backend environment
    if [ ! -f "$BACKEND_DIR/.env" ]; then
        error "Backend environment file not found: $BACKEND_DIR/.env"
        error "Please copy $BACKEND_DIR/env.production to $BACKEND_DIR/.env and configure it"
        exit 1
    fi
    
    # Check frontend environment
    if [ ! -f "$FRONTEND_DIR/.env.local" ]; then
        error "Frontend environment file not found: $FRONTEND_DIR/.env.local"
        error "Please copy $FRONTEND_DIR/env.production to $FRONTEND_DIR/.env.local and configure it"
        exit 1
    fi
    
    # Validate required environment variables
    local backend_vars=("DATABASE_URL" "SUPABASE_URL" "SUPABASE_SERVICE_ROLE_KEY")
    local frontend_vars=("VITE_SUPABASE_URL" "VITE_SUPABASE_ANON_KEY" "VITE_API_URL")
    
    for var in "${backend_vars[@]}"; do
        if ! grep -q "^$var=" "$BACKEND_DIR/.env"; then
            error "Missing required backend environment variable: $var"
            exit 1
        fi
    done
    
    for var in "${frontend_vars[@]}"; do
        if ! grep -q "^$var=" "$FRONTEND_DIR/.env.local"; then
            error "Missing required frontend environment variable: $var"
            exit 1
        fi
    done
    
    success "Environment validation passed"
}

# Build applications
build_applications() {
    log "Building applications..."
    
    # Build backend
    log "Building backend..."
    cd "$BACKEND_DIR"
    make build-prod-with-version
    cd ..
    
    # Build frontend
    log "Building frontend..."
    cd "$FRONTEND_DIR"
    npm ci --only=production
    npm run build:prod
    cd ..
    
    success "Applications built successfully"
}

# Build Docker images
build_docker_images() {
    log "Building Docker images..."
    
    # Build backend image
    log "Building backend Docker image..."
    docker build -t "$PROJECT_NAME-backend:latest" "$BACKEND_DIR"
    
    # Build frontend image
    log "Building frontend Docker image..."
    docker build -t "$PROJECT_NAME-frontend:latest" "$FRONTEND_DIR"
    
    success "Docker images built successfully"
}

# Deploy application
deploy_application() {
    log "Deploying application..."
    
    # Stop existing containers
    log "Stopping existing containers..."
    docker-compose -f "$DOCKER_COMPOSE_PROD" down --remove-orphans
    
    # Start application
    log "Starting application..."
    docker-compose -f "$DOCKER_COMPOSE_PROD" up -d
    
    # Wait for services to be healthy
    log "Waiting for services to be healthy..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f "$DOCKER_COMPOSE_PROD" ps | grep -q "healthy"; then
            success "All services are healthy"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            error "Services failed to become healthy after $max_attempts attempts"
            docker-compose -f "$DOCKER_COMPOSE_PROD" logs
            exit 1
        fi
        
        log "Waiting for services to be healthy... (attempt $attempt/$max_attempts)"
        sleep 10
        ((attempt++))
    done
}

# Deploy monitoring (optional)
deploy_monitoring() {
    if [ "${DEPLOY_MONITORING:-false}" = "true" ]; then
        log "Deploying monitoring stack..."
        
        # Create monitoring directories if they don't exist
        mkdir -p monitoring/grafana/dashboards
        mkdir -p monitoring/grafana/datasources
        
        # Start monitoring services
        docker-compose -f "$DOCKER_COMPOSE_MONITORING" up -d
        
        success "Monitoring stack deployed"
        log "Grafana available at: http://localhost:3001 (admin/admin)"
        log "Prometheus available at: http://localhost:9090"
        log "Jaeger available at: http://localhost:16686"
    else
        log "Skipping monitoring deployment (set DEPLOY_MONITORING=true to enable)"
    fi
}

# Health check
health_check() {
    log "Performing health checks..."
    
    local endpoints=(
        "http://localhost:3000/health"
        "http://localhost:8080/health"
    )
    
    for endpoint in "${endpoints[@]}"; do
        log "Checking $endpoint..."
        if curl -f -s "$endpoint" > /dev/null; then
            success "$endpoint is healthy"
        else
            error "$endpoint is not responding"
            return 1
        fi
    done
    
    success "All health checks passed"
}

# Main deployment function
main() {
    log "Starting production deployment..."
    
    # Parse command line arguments
    local deploy_monitoring=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --monitoring)
                deploy_monitoring=true
                shift
                ;;
            --help)
                echo "Usage: $0 [--monitoring]"
                echo "  --monitoring  Deploy monitoring stack"
                echo "  --help        Show this help message"
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Set monitoring flag
    export DEPLOY_MONITORING="$deploy_monitoring"
    
    # Run deployment steps
    check_requirements
    validate_environment
    build_applications
    build_docker_images
    deploy_application
    deploy_monitoring
    health_check
    
    success "Production deployment completed successfully!"
    
    log "Application URLs:"
    log "  Frontend: http://localhost:3000"
    log "  Backend API: http://localhost:8080"
    log "  Backend Metrics: http://localhost:9090"
    
    if [ "$deploy_monitoring" = true ]; then
        log "Monitoring URLs:"
        log "  Grafana: http://localhost:3001 (admin/admin)"
        log "  Prometheus: http://localhost:9090"
        log "  Jaeger: http://localhost:16686"
    fi
}

# Run main function
main "$@" 
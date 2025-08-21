#!/bin/bash

# Deployment script for Crypto Trading Bots
# This script automates the deployment process for different environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENVIRONMENT=${1:-development}
PLATFORM=${2:-docker}

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    case $PLATFORM in
        docker)
            if ! command -v docker &> /dev/null; then
                error "Docker is not installed"
            fi
            if ! command -v docker-compose &> /dev/null; then
                error "Docker Compose is not installed"
            fi
            log "Docker and Docker Compose are available"
            ;;
        kubernetes)
            if ! command -v kubectl &> /dev/null; then
                error "kubectl is not installed"
            fi
            if ! kubectl cluster-info &> /dev/null; then
                error "Kubernetes cluster is not accessible"
            fi
            log "Kubernetes cluster is accessible"
            ;;
        *)
            error "Unsupported platform: $PLATFORM"
            ;;
    esac
}

# Validate environment variables
validate_env() {
    log "Validating environment variables..."
    
    # Check if .env file exists
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        warn ".env file not found, creating from template..."
        cp "$PROJECT_ROOT/env.example" "$PROJECT_ROOT/.env"
        warn "Please update .env file with your configuration"
    fi
    
    # Source environment variables
    source "$PROJECT_ROOT/.env"
    
    # Validate required variables
    if [ "$ENVIRONMENT" = "production" ]; then
        if [ -z "$EXCHANGE_API_KEY" ] || [ "$EXCHANGE_API_KEY" = "your-api-key-here" ]; then
            error "EXCHANGE_API_KEY is not set"
        fi
        if [ -z "$EXCHANGE_SECRET_KEY" ] || [ "$EXCHANGE_SECRET_KEY" = "your-secret-key-here" ]; then
            error "EXCHANGE_SECRET_KEY is not set"
        fi
        if [ -z "$POSTGRES_PASSWORD" ]; then
            error "POSTGRES_PASSWORD is not set"
        fi
    fi
    
    log "Environment variables validated"
}

# Build application
build_app() {
    log "Building application..."
    
    cd "$PROJECT_ROOT"
    
    # Build binaries
    make build
    
    # Build Docker image if using Docker
    if [ "$PLATFORM" = "docker" ]; then
        log "Building Docker image..."
        docker build -t crypto-trading-bot:latest .
    fi
    
    log "Application built successfully"
}

# Deploy to Docker
deploy_docker() {
    log "Deploying to Docker..."
    
    cd "$PROJECT_ROOT/deployments/docker"
    
    case $ENVIRONMENT in
        development)
            log "Starting development environment..."
            docker-compose -f docker-compose.dev.yml up -d
            ;;
        production)
            log "Starting production environment..."
            docker-compose -f docker-compose.prod.yml up -d
            ;;
        *)
            error "Unsupported environment: $ENVIRONMENT"
            ;;
    esac
    
    log "Docker deployment completed"
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log "Deploying to Kubernetes..."
    
    cd "$PROJECT_ROOT/deployments/kubernetes"
    
    # Create namespace
    log "Creating namespace..."
    kubectl apply -f namespace.yaml
    
    # Apply configurations
    log "Applying configurations..."
    kubectl apply -f configmap.yaml
    kubectl apply -f secret.yaml
    
    # Deploy services
    log "Deploying services..."
    kubectl apply -f dca-bot-deployment.yaml
    
    # Apply ingress
    log "Applying ingress..."
    kubectl apply -f ingress.yaml
    
    # Wait for deployment
    log "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/dca-bot -n crypto-trading
    
    log "Kubernetes deployment completed"
}

# Setup monitoring
setup_monitoring() {
    log "Setting up monitoring..."
    
    cd "$PROJECT_ROOT/deployments/monitoring"
    
    if [ "$PLATFORM" = "kubernetes" ]; then
        # Apply monitoring configurations
        kubectl apply -f prometheus-config.yaml
        kubectl apply -f alerting-rules.yml
        kubectl apply -f grafana-dashboards.yaml
        
        log "Monitoring setup completed"
    else
        warn "Monitoring setup is only available for Kubernetes deployments"
    fi
}

# Health check
health_check() {
    log "Performing health check..."
    
    case $PLATFORM in
        docker)
            # Check if containers are running
            if docker-compose -f "$PROJECT_ROOT/deployments/docker/docker-compose.$ENVIRONMENT.yml" ps | grep -q "Up"; then
                log "All containers are running"
            else
                error "Some containers are not running"
            fi
            
            # Check health endpoints
            sleep 10
            if curl -f http://localhost:8080/health &> /dev/null; then
                log "DCA Bot health check passed"
            else
                warn "DCA Bot health check failed"
            fi
            ;;
        kubernetes)
            # Check if pods are running
            if kubectl get pods -n crypto-trading | grep -q "Running"; then
                log "All pods are running"
            else
                error "Some pods are not running"
            fi
            
            # Check services
            if kubectl get svc -n crypto-trading | grep -q "ClusterIP"; then
                log "Services are configured"
            else
                error "Services are not configured"
            fi
            ;;
    esac
    
    log "Health check completed"
}

# Show deployment info
show_info() {
    log "Deployment completed successfully!"
    echo
    echo "Environment: $ENVIRONMENT"
    echo "Platform: $PLATFORM"
    echo
    
    case $PLATFORM in
        docker)
            echo "Services:"
            echo "  - DCA Bot: http://localhost:8080"
            echo "  - Grid Bot: http://localhost:8081"
            echo "  - Combo Bot: http://localhost:8082"
            echo "  - pgAdmin: http://localhost:5050"
            echo "  - Redis Commander: http://localhost:8081"
            echo
            echo "To view logs:"
            echo "  docker-compose -f deployments/docker/docker-compose.$ENVIRONMENT.yml logs -f"
            ;;
        kubernetes)
            echo "Services:"
            echo "  - DCA Bot: kubectl port-forward svc/dca-bot-service 8080:8080 -n crypto-trading"
            echo "  - Grid Bot: kubectl port-forward svc/grid-bot-service 8081:8081 -n crypto-trading"
            echo "  - Combo Bot: kubectl port-forward svc/combo-bot-service 8082:8082 -n crypto-trading"
            echo
            echo "To view logs:"
            echo "  kubectl logs -f deployment/dca-bot -n crypto-trading"
            ;;
    esac
    
    echo
    echo "For more information, see:"
    echo "  - README.md"
    echo "  - deployments/README.md"
}

# Main deployment function
main() {
    log "Starting deployment..."
    log "Environment: $ENVIRONMENT"
    log "Platform: $PLATFORM"
    
    check_prerequisites
    validate_env
    build_app
    
    case $PLATFORM in
        docker)
            deploy_docker
            ;;
        kubernetes)
            deploy_kubernetes
            ;;
    esac
    
    setup_monitoring
    health_check
    show_info
}

# Help function
show_help() {
    echo "Usage: $0 [environment] [platform]"
    echo
    echo "Environments:"
    echo "  development  - Development environment (default)"
    echo "  production   - Production environment"
    echo
    echo "Platforms:"
    echo "  docker       - Docker deployment (default)"
    echo "  kubernetes   - Kubernetes deployment"
    echo
    echo "Examples:"
    echo "  $0                    # Deploy development environment with Docker"
    echo "  $0 production         # Deploy production environment with Docker"
    echo "  $0 development k8s    # Deploy development environment with Kubernetes"
    echo "  $0 production k8s     # Deploy production environment with Kubernetes"
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
esac

# Run main function
main "$@"

#!/bin/bash

# Database backup script for Crypto Trading Bots
# This script creates automated backups of the PostgreSQL database

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="$PROJECT_ROOT/backups"
DATE=$(date +%Y%m%d_%H%M%S)
PLATFORM=${1:-docker}

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

# Create backup directory
create_backup_dir() {
    log "Creating backup directory..."
    mkdir -p "$BACKUP_DIR"
    log "Backup directory created: $BACKUP_DIR"
}

# Docker backup
backup_docker() {
    log "Creating Docker database backup..."
    
    # Get database container name
    CONTAINER_NAME="crypto-postgres-prod"
    if [ "$PLATFORM" = "development" ]; then
        CONTAINER_NAME="crypto-postgres-dev"
    fi
    
    # Check if container is running
    if ! docker ps | grep -q "$CONTAINER_NAME"; then
        error "Database container $CONTAINER_NAME is not running"
    fi
    
    # Create backup
    BACKUP_FILE="$BACKUP_DIR/postgres_backup_${DATE}.sql"
    docker exec "$CONTAINER_NAME" pg_dump -U crypto_user crypto_trading > "$BACKUP_FILE"
    
    # Compress backup
    gzip "$BACKUP_FILE"
    
    log "Database backup created: ${BACKUP_FILE}.gz"
    
    # Clean old backups (keep last 7 days)
    find "$BACKUP_DIR" -name "postgres_backup_*.sql.gz" -mtime +7 -delete
    log "Old backups cleaned"
}

# Kubernetes backup
backup_kubernetes() {
    log "Creating Kubernetes database backup..."
    
    # Get database pod name
    POD_NAME=$(kubectl get pods -n crypto-trading -l app=postgres -o jsonpath='{.items[0].metadata.name}')
    
    if [ -z "$POD_NAME" ]; then
        error "PostgreSQL pod not found in crypto-trading namespace"
    fi
    
    # Create backup
    BACKUP_FILE="$BACKUP_DIR/postgres_backup_${DATE}.sql"
    kubectl exec -n crypto-trading "$POD_NAME" -- pg_dump -U crypto_user crypto_trading > "$BACKUP_FILE"
    
    # Compress backup
    gzip "$BACKUP_FILE"
    
    log "Database backup created: ${BACKUP_FILE}.gz"
    
    # Clean old backups (keep last 7 days)
    find "$BACKUP_DIR" -name "postgres_backup_*.sql.gz" -mtime +7 -delete
    log "Old backups cleaned"
}

# Backup configuration files
backup_config() {
    log "Creating configuration backup..."
    
    CONFIG_BACKUP="$BACKUP_DIR/config_backup_${DATE}.tar.gz"
    
    cd "$PROJECT_ROOT"
    tar -czf "$CONFIG_BACKUP" \
        configs/ \
        deployments/ \
        .env \
        go.mod \
        go.sum \
        Makefile \
        Dockerfile \
        docker-compose.yml
    
    log "Configuration backup created: $CONFIG_BACKUP"
}

# Verify backup
verify_backup() {
    log "Verifying backup..."
    
    BACKUP_FILE="$BACKUP_DIR/postgres_backup_${DATE}.sql.gz"
    
    if [ ! -f "$BACKUP_FILE" ]; then
        error "Backup file not found: $BACKUP_FILE"
    fi
    
    # Check file size
    SIZE=$(stat -f%z "$BACKUP_FILE")
    if [ "$SIZE" -lt 1000 ]; then
        warn "Backup file seems too small: $SIZE bytes"
    else
        log "Backup file size: $SIZE bytes"
    fi
    
    # Test backup integrity
    if gunzip -t "$BACKUP_FILE"; then
        log "Backup integrity verified"
    else
        error "Backup integrity check failed"
    fi
}

# Show backup info
show_backup_info() {
    log "Backup completed successfully!"
    echo
    echo "Backup files:"
    echo "  - Database: $BACKUP_DIR/postgres_backup_${DATE}.sql.gz"
    echo "  - Config: $BACKUP_DIR/config_backup_${DATE}.tar.gz"
    echo
    echo "To restore database:"
    echo "  gunzip -c $BACKUP_DIR/postgres_backup_${DATE}.sql.gz | psql -h localhost -U crypto_user -d crypto_trading"
    echo
    echo "To restore configuration:"
    echo "  tar -xzf $BACKUP_DIR/config_backup_${DATE}.tar.gz"
}

# Main backup function
main() {
    log "Starting backup process..."
    log "Platform: $PLATFORM"
    log "Date: $DATE"
    
    create_backup_dir
    
    case $PLATFORM in
        docker)
            backup_docker
            ;;
        kubernetes)
            backup_kubernetes
            ;;
        *)
            error "Unsupported platform: $PLATFORM"
            ;;
    esac
    
    backup_config
    verify_backup
    show_backup_info
}

# Help function
show_help() {
    echo "Usage: $0 [platform]"
    echo
    echo "Platforms:"
    echo "  docker       - Docker backup (default)"
    echo "  kubernetes   - Kubernetes backup"
    echo
    echo "Examples:"
    echo "  $0           # Create Docker backup"
    echo "  $0 kubernetes # Create Kubernetes backup"
    echo
    echo "Backup files will be stored in: $BACKUP_DIR"
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

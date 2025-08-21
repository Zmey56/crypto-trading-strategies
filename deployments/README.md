# Deployment Guide for Crypto Trading Bots

This directory contains deployment configurations for different environments and platforms.

## 📁 Directory Structure

```
deployments/
├── docker/           # Docker and Docker Compose configurations
├── kubernetes/       # Kubernetes manifests and configurations
├── monitoring/       # Monitoring and observability configurations
└── README.md        # This file
```

## 🚀 Quick Start

### Docker Deployment (Recommended for development)

```bash
# Navigate to docker directory
cd deployments/docker

# Development environment
docker-compose -f docker-compose.dev.yml up -d

# Production environment
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes Deployment (Recommended for production)

```bash
# Navigate to kubernetes directory
cd deployments/kubernetes

# Apply all configurations
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f dca-bot-deployment.yaml
kubectl apply -f ingress.yaml
```

## 🔧 Environment Setup

### Prerequisites

#### Docker
- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 4GB RAM available

#### Kubernetes
- Kubernetes cluster 1.20+
- kubectl configured
- Helm 3.0+ (optional)
- Ingress controller (nginx-ingress)
- cert-manager (for SSL certificates)

#### Monitoring
- Prometheus
- Grafana
- AlertManager

### Environment Variables

Create a `.env` file in the root directory:

```bash
# Exchange API Configuration
EXCHANGE_API_KEY=your-api-key-here
EXCHANGE_SECRET_KEY=your-secret-key-here
EXCHANGE_SANDBOX=false

# Database Configuration
POSTGRES_PASSWORD=your-secure-password
POSTGRES_DB=crypto_trading
POSTGRES_USER=crypto_user

# Monitoring Configuration
GRAFANA_ADMIN_PASSWORD=your-grafana-password
PROMETHEUS_RETENTION_DAYS=30

# Security Configuration
JWT_SECRET=your-jwt-secret-key
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
```

## 📊 Monitoring Setup

### Prometheus Configuration

1. Deploy Prometheus with the provided configuration:
```bash
kubectl apply -f monitoring/prometheus-config.yaml
```

2. Apply alerting rules:
```bash
kubectl apply -f monitoring/alerting-rules.yml
```

### Grafana Dashboards

1. Deploy Grafana dashboards:
```bash
kubectl apply -f monitoring/grafana-dashboards.yaml
```

2. Access Grafana at `http://localhost:3000` (default credentials: admin/admin)

## 🔒 Security Considerations

### Production Security Checklist

- [ ] Use strong, unique passwords for all services
- [ ] Enable SSL/TLS encryption
- [ ] Configure firewall rules
- [ ] Set up proper RBAC in Kubernetes
- [ ] Use secrets management (HashiCorp Vault, AWS Secrets Manager)
- [ ] Enable audit logging
- [ ] Regular security updates
- [ ] Network segmentation
- [ ] Backup and disaster recovery

### API Security

- [ ] Rate limiting enabled
- [ ] Authentication required
- [ ] Input validation
- [ ] SQL injection prevention
- [ ] XSS protection
- [ ] CORS properly configured

## 📈 Scaling Considerations

### Horizontal Scaling

- **Docker**: Use Docker Swarm or multiple instances
- **Kubernetes**: Use HorizontalPodAutoscaler
- **Load Balancing**: Configure proper load balancers

### Vertical Scaling

- Monitor resource usage
- Adjust CPU and memory limits
- Optimize application performance

## 🔄 Backup and Recovery

### Database Backup

```bash
# PostgreSQL backup
pg_dump -h localhost -U crypto_user crypto_trading > backup.sql

# Automated backup script
./scripts/backup-database.sh
```

### Configuration Backup

```bash
# Backup Kubernetes configurations
kubectl get all -n crypto-trading -o yaml > backup.yaml

# Backup Docker volumes
docker run --rm -v crypto-trading_postgres-data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz -C /data .
```

## 🚨 Troubleshooting

### Common Issues

1. **Bot not starting**
   - Check environment variables
   - Verify API keys
   - Check logs: `docker logs crypto-dca-bot`

2. **Database connection issues**
   - Verify PostgreSQL is running
   - Check credentials
   - Test connection: `psql -h localhost -U crypto_user -d crypto_trading`

3. **Monitoring not working**
   - Check Prometheus targets
   - Verify metrics endpoints
   - Check Grafana data sources

### Log Locations

- **Docker**: `docker logs <container-name>`
- **Kubernetes**: `kubectl logs -n crypto-trading <pod-name>`
- **Application**: `/app/logs/` directory

## 📞 Support

For deployment issues:

1. Check the logs first
2. Verify configuration files
3. Test connectivity between services
4. Review security settings
5. Check resource limits

## 🔄 Updates and Maintenance

### Rolling Updates

```bash
# Docker
docker-compose pull
docker-compose up -d

# Kubernetes
kubectl set image deployment/dca-bot dca-bot=crypto-trading-bot:latest
```

### Health Checks

- Monitor application health endpoints
- Set up automated health checks
- Configure alerting for failures
- Regular backup verification

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)

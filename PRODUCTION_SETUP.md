# Production Setup Guide

This guide covers the complete production setup for the Collaborative Bucket List application.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Configuration](#environment-configuration)
3. [Security Setup](#security-setup)
4. [Database Setup](#database-setup)
5. [Build and Deployment](#build-and-deployment)
6. [Monitoring and Logging](#monitoring-and-logging)
7. [SSL/HTTPS Setup](#sslhttps-setup)
8. [Backup and Recovery](#backup-and-recovery)
9. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

- Docker and Docker Compose
- Git
- Make (for backend builds)
- Node.js and npm (for frontend builds)
- curl (for health checks)

### System Requirements

- **CPU**: 2+ cores
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 20GB minimum
- **Network**: Stable internet connection

## Environment Configuration

### 1. Backend Environment Setup

```bash
# Copy the production environment template
cp backend/env.production backend/.env

# Edit the environment file with your production values
nano backend/.env
```

**Required Variables:**

- `DATABASE_URL`: PostgreSQL connection string
- `SUPABASE_URL`: Your Supabase project URL
- `SUPABASE_SERVICE_ROLE_KEY`: Supabase service role key
- `SUPABASE_JWT_SECRET`: JWT secret for authentication
- `ALLOWED_ORIGINS`: CORS allowed origins (your domain)
- `FRONTEND_URL`: Your frontend domain

**Production Optimizations:**

- `GIN_MODE=release`: Production mode
- `LOG_LEVEL=info`: Info level logging
- `LOG_FORMAT=json`: JSON structured logging
- `METRICS_ENABLED=true`: Enable metrics collection
- `RATE_LIMIT_REQUESTS_PER_MINUTE=30`: Rate limiting

### 2. Frontend Environment Setup

```bash
# Copy the production environment template
cp frontend/env.production frontend/.env.local

# Edit the environment file with your production values
nano frontend/.env.local
```

**Required Variables:**

- `VITE_SUPABASE_URL`: Your Supabase project URL
- `VITE_SUPABASE_ANON_KEY`: Supabase anonymous key
- `VITE_API_URL`: Your backend API URL
- `VITE_WS_URL`: Your WebSocket server URL

**Production Optimizations:**

- `VITE_ENVIRONMENT=production`: Production environment
- `VITE_ENABLE_ANALYTICS=true`: Enable analytics
- `VITE_ENABLE_ERROR_TRACKING=true`: Enable error tracking
- `VITE_ENABLE_DEBUG_MODE=false`: Disable debug mode

## Security Setup

### 1. Environment Security

```bash
# Set proper file permissions
chmod 600 backend/.env
chmod 600 frontend/.env.local

# Ensure .env files are in .gitignore
echo ".env" >> .gitignore
echo ".env.local" >> .gitignore
```

### 2. Database Security

- Use strong passwords for database users
- Enable SSL connections to database
- Implement connection pooling
- Set up database backups

### 3. Application Security

- Enable HTTPS/SSL
- Configure proper CORS settings
- Implement rate limiting
- Set up security headers
- Use secure JWT secrets

## Database Setup

### PostgreSQL Production Setup

1. **Install PostgreSQL** (if not using managed service)
2. **Create Database and User:**

```sql
CREATE DATABASE bucket_list_db;
CREATE USER bucket_list_user WITH PASSWORD 'strong_password';
GRANT ALL PRIVILEGES ON DATABASE bucket_list_db TO bucket_list_user;
```

3. **Run Migrations:**

```bash
cd backend
# Apply database migrations
# (This will be handled by the application startup)
```

### Supabase Production Setup

1. **Create Production Project** in Supabase dashboard
2. **Configure Row Level Security (RLS)**
3. **Set up Authentication Policies**
4. **Configure API Keys**

## Build and Deployment

### 1. Manual Deployment

```bash
# Build applications
cd backend && make build-prod-with-version && cd ..
cd frontend && npm run build:prod && cd ..

# Build Docker images
docker build -t collaborative-bucket-list-backend:latest backend/
docker build -t collaborative-bucket-list-frontend:latest frontend/

# Deploy with Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

### 2. Automated Deployment

```bash
# Use the production deployment script
./scripts/deploy-production.sh

# Deploy with monitoring
./scripts/deploy-production.sh --monitoring
```

### 3. Health Checks

```bash
# Check application health
curl -f http://localhost:3000/health
curl -f http://localhost:8080/health

# Check Docker containers
docker-compose -f docker-compose.prod.yml ps
```

## Monitoring and Logging

### 1. Application Metrics

The backend exposes metrics at `/metrics` endpoint:

- HTTP request metrics
- Database connection metrics
- Custom business metrics

### 2. Logging Configuration

**Backend Logging:**

- JSON structured logging
- Log levels: debug, info, warn, error
- Log rotation and retention

**Frontend Logging:**

- Error tracking with Sentry
- Performance monitoring
- User analytics

### 3. Monitoring Stack

Deploy the monitoring stack for comprehensive observability:

```bash
# Deploy monitoring services
docker-compose -f docker-compose.monitoring.yml up -d
```

**Available Monitoring Tools:**

- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization
- **Loki**: Log aggregation
- **Alertmanager**: Alert management
- **Jaeger**: Distributed tracing

### 4. Alerting Rules

Configure alerts for:

- High error rates
- Slow response times
- Database connection issues
- Disk space usage
- Memory usage

## SSL/HTTPS Setup

### 1. Using Let's Encrypt

```bash
# Install Certbot
sudo apt-get install certbot

# Obtain SSL certificate
sudo certbot certonly --standalone -d yourdomain.com

# Configure nginx with SSL
# (See nginx/nginx.conf for SSL configuration)
```

### 2. Using Reverse Proxy

Configure nginx as reverse proxy with SSL termination:

```nginx
server {
    listen 443 ssl;
    server_name yourdomain.com;

    ssl_certificate /etc/ssl/certs/your-cert.pem;
    ssl_certificate_key /etc/ssl/private/your-key.pem;

    location / {
        proxy_pass http://frontend:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Backup and Recovery

### 1. Database Backups

```bash
# Create backup script
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="bucket_list_db"

# Create backup
pg_dump $DATABASE_URL > "$BACKUP_DIR/backup_$DATE.sql"

# Compress backup
gzip "$BACKUP_DIR/backup_$DATE.sql"

# Clean old backups (keep last 7 days)
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete
```

### 2. Application Backups

```bash
# Backup environment files
cp backend/.env "backups/env_backup_$(date +%Y%m%d_%H%M%S)"
cp frontend/.env.local "backups/env_local_backup_$(date +%Y%m%d_%H%M%S)"

# Backup Docker volumes
docker run --rm -v collaborative-bucket-list_backend_logs:/data -v $(pwd)/backups:/backup alpine tar czf /backup/logs_backup_$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
```

### 3. Recovery Procedures

**Database Recovery:**

```bash
# Restore from backup
psql $DATABASE_URL < backup_20231201_120000.sql
```

**Application Recovery:**

```bash
# Restore environment files
cp backups/env_backup_20231201_120000 backend/.env
cp backups/env_local_backup_20231201_120000 frontend/.env.local

# Redeploy application
./scripts/deploy-production.sh
```

## Troubleshooting

### Common Issues

1. **Application Won't Start**

   - Check environment variables
   - Verify database connectivity
   - Check Docker logs: `docker-compose logs`

2. **Database Connection Issues**

   - Verify DATABASE_URL format
   - Check network connectivity
   - Ensure database is running

3. **CORS Errors**

   - Verify ALLOWED_ORIGINS includes your domain
   - Check frontend API URL configuration

4. **WebSocket Connection Failed**

   - Verify WebSocket URL configuration
   - Check firewall settings
   - Ensure backend WebSocket endpoint is working

5. **High Memory Usage**
   - Check for memory leaks
   - Optimize database queries
   - Increase container memory limits

### Debug Commands

```bash
# Check container status
docker-compose -f docker-compose.prod.yml ps

# View application logs
docker-compose -f docker-compose.prod.yml logs -f backend
docker-compose -f docker-compose.prod.yml logs -f frontend

# Check resource usage
docker stats

# Access container shell
docker-compose -f docker-compose.prod.yml exec backend sh
docker-compose -f docker-compose.prod.yml exec frontend sh

# Check network connectivity
docker-compose -f docker-compose.prod.yml exec backend wget -qO- http://localhost:8080/health
```

### Performance Optimization

1. **Database Optimization**

   - Add database indexes
   - Optimize queries
   - Configure connection pooling

2. **Application Optimization**

   - Enable gzip compression
   - Optimize static assets
   - Implement caching

3. **Infrastructure Optimization**
   - Use CDN for static assets
   - Implement load balancing
   - Configure auto-scaling

## Maintenance

### Regular Tasks

1. **Daily**

   - Check application health
   - Monitor error rates
   - Review logs for issues

2. **Weekly**

   - Update dependencies
   - Review security patches
   - Check backup integrity

3. **Monthly**
   - Performance review
   - Security audit
   - Capacity planning

### Update Procedures

```bash
# Update application
git pull origin main
./scripts/deploy-production.sh

# Update monitoring stack
docker-compose -f docker-compose.monitoring.yml pull
docker-compose -f docker-compose.monitoring.yml up -d
```

## Support

For additional support:

1. Check the application logs
2. Review monitoring dashboards
3. Consult the troubleshooting section
4. Check the ENVIRONMENT.md file for configuration details

## Security Checklist

- [ ] Environment variables are secure
- [ ] Database connections use SSL
- [ ] HTTPS is enabled
- [ ] CORS is properly configured
- [ ] Rate limiting is enabled
- [ ] Security headers are set
- [ ] JWT secrets are strong
- [ ] Regular security updates
- [ ] Backup procedures tested
- [ ] Monitoring alerts configured

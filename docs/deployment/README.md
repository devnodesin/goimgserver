# Deployment Guide

## Overview

This guide provides comprehensive deployment instructions for goimgserver across different environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Systemd Deployment](#systemd-deployment)
- [Docker Deployment](#docker-deployment)
- [Nginx Reverse Proxy](#nginx-reverse-proxy)
- [Environment Variables](#environment-variables)
- [Security Considerations](#security-considerations)
- [Monitoring](#monitoring)
- [Maintenance](#maintenance)

## Prerequisites

### System Requirements

**Minimum:**
- CPU: 2 cores
- RAM: 4GB
- Storage: 20GB (10GB for application + cache)
- OS: Linux (Ubuntu 20.04+, Debian 11+, or RHEL 8+)

**Recommended:**
- CPU: 4+ cores
- RAM: 8GB+
- Storage: 100GB+ SSD
- OS: Ubuntu 22.04 LTS or Debian 12

**High-Traffic:**
- CPU: 8+ cores
- RAM: 16GB+
- Storage: 500GB+ NVMe SSD
- Network: 1Gbps+ connection
- Load Balancer: Multiple instances

### Software Dependencies

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y libvips-dev pkg-config

# RHEL/CentOS
sudo yum install -y vips-devel pkgconfig

# macOS
brew install vips
```

## Systemd Deployment

### 1. Build the Application

```bash
cd src
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o goimgserver \
    main.go
```

### 2. Install Using Script

```bash
# Copy binary to installation directory
sudo mkdir -p /opt/goimgserver/bin
sudo cp goimgserver /opt/goimgserver/bin/

# Run installation script
cd ../docs/deployment/systemd
sudo ./install.sh
```

### 3. Manual Installation

If you prefer manual installation:

```bash
# Create user
sudo useradd --system --no-create-home --shell /bin/false goimgserver

# Create directories
sudo mkdir -p /opt/goimgserver/{bin,images,cache,logs}
sudo chown -R goimgserver:goimgserver /opt/goimgserver

# Copy binary
sudo cp goimgserver /opt/goimgserver/bin/
sudo chmod 755 /opt/goimgserver/bin/goimgserver

# Copy service file
sudo cp goimgserver.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/goimgserver.service

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable goimgserver
```

### 4. Start the Service

```bash
# Start service
sudo systemctl start goimgserver

# Check status
sudo systemctl status goimgserver

# View logs
sudo journalctl -u goimgserver -f
```

### 5. Service Management

```bash
# Stop service
sudo systemctl stop goimgserver

# Restart service
sudo systemctl restart goimgserver

# Reload configuration (if supported)
sudo systemctl reload goimgserver

# Disable service
sudo systemctl disable goimgserver
```

## Docker Deployment

### 1. Using Docker Compose (Recommended)

```bash
# Navigate to docker directory
cd docs/deployment/docker

# Create images directory
mkdir -p ./images
cp /path/to/your/images/* ./images/

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### 2. Using Docker Directly

```bash
# Build image
docker build -f docs/deployment/docker/Dockerfile -t goimgserver:latest .

# Run container
docker run -d \
    --name goimgserver \
    -p 9000:9000 \
    -v /path/to/images:/app/images:ro \
    -v goimgserver-cache:/app/cache \
    -e GOMAXPROCS=4 \
    -e GOGC=100 \
    --restart unless-stopped \
    goimgserver:latest
```

### 3. Docker Image Management

```bash
# View running containers
docker ps

# Stop container
docker stop goimgserver

# Start container
docker start goimgserver

# View logs
docker logs -f goimgserver

# Remove container
docker rm -f goimgserver

# Remove image
docker rmi goimgserver:latest
```

### 4. Docker Health Checks

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' goimgserver

# View health check logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' goimgserver
```

## Nginx Reverse Proxy

### 1. Install Nginx

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y nginx

# RHEL/CentOS
sudo yum install -y nginx
```

### 2. Configure Nginx

```bash
# Copy configuration
sudo cp docs/deployment/nginx/goimgserver.conf /etc/nginx/sites-available/

# Enable site (Debian/Ubuntu)
sudo ln -s /etc/nginx/sites-available/goimgserver.conf /etc/nginx/sites-enabled/

# For RHEL/CentOS, copy directly to conf.d
sudo cp docs/deployment/nginx/goimgserver.conf /etc/nginx/conf.d/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### 3. SSL/TLS Configuration

```bash
# Install certbot for Let's Encrypt
sudo apt-get install -y certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d img.example.com

# Auto-renewal
sudo certbot renew --dry-run
```

### 4. Cache Configuration

Create cache directory:

```bash
sudo mkdir -p /var/cache/nginx/goimgserver
sudo chown -R www-data:www-data /var/cache/nginx/goimgserver
```

## Environment Variables

### Application Variables

```bash
# Port configuration
export GOIMGSERVER_PORT=9000

# Directory paths
export GOIMGSERVER_IMAGES_DIR=/opt/goimgserver/images
export GOIMGSERVER_CACHE_DIR=/opt/goimgserver/cache

# Logging
export GOIMGSERVER_LOG_LEVEL=info  # debug, info, warn, error

# Pre-cache settings
export GOIMGSERVER_PRECACHE=true
export GOIMGSERVER_PRECACHE_WORKERS=0  # 0 = auto (CPU count)

# Security
export GOIMGSERVER_MAX_FILE_SIZE=100MB
export GOIMGSERVER_MAX_DIMENSIONS=4000
export GOIMGSERVER_RATE_LIMIT=100
```

### Go Runtime Variables

```bash
# CPU cores to use
export GOMAXPROCS=4

# Garbage collection target percentage
export GOGC=100

# Memory limit (Go 1.19+)
export GOMEMLIMIT=6GiB
```

### Setting in Systemd

Edit `/etc/systemd/system/goimgserver.service`:

```ini
[Service]
Environment="GOMAXPROCS=4"
Environment="GOGC=100"
Environment="GOMEMLIMIT=6GiB"
```

### Setting in Docker

Edit `docker-compose.yml`:

```yaml
environment:
  - GOMAXPROCS=4
  - GOGC=100
  - GOMEMLIMIT=6GiB
```

## Security Considerations

### 1. File System Permissions

```bash
# Restrict binary access
sudo chmod 755 /opt/goimgserver/bin/goimgserver

# Images directory (read-only)
sudo chmod 755 /opt/goimgserver/images
sudo chown -R goimgserver:goimgserver /opt/goimgserver/images

# Cache directory (read-write)
sudo chmod 775 /opt/goimgserver/cache
sudo chown -R goimgserver:goimgserver /opt/goimgserver/cache
```

### 2. Firewall Configuration

```bash
# Allow HTTP (if not using nginx)
sudo ufw allow 9000/tcp

# Allow HTTPS (for nginx)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable
```

### 3. SELinux Configuration (RHEL/CentOS)

```bash
# Set context for directories
sudo semanage fcontext -a -t httpd_sys_content_t "/opt/goimgserver/images(/.*)?"
sudo semanage fcontext -a -t httpd_sys_rw_content_t "/opt/goimgserver/cache(/.*)?"
sudo restorecon -Rv /opt/goimgserver
```

### 4. Command Endpoint Security

Restrict access to `/cmd/` endpoints in nginx:

```nginx
location /cmd/ {
    allow 127.0.0.1;
    allow 10.0.0.0/8;  # Internal network
    deny all;
}
```

## Monitoring

### 1. Health Checks

```bash
# Check health endpoint
curl http://localhost:9000/health

# Check readiness
curl http://localhost:9000/health/ready
```

### 2. Systemd Logs

```bash
# View recent logs
sudo journalctl -u goimgserver --since today

# Follow logs
sudo journalctl -u goimgserver -f

# View errors only
sudo journalctl -u goimgserver -p err
```

### 3. Docker Logs

```bash
# View logs
docker logs goimgserver

# Follow logs
docker logs -f goimgserver

# Last 100 lines
docker logs --tail 100 goimgserver
```

### 4. Performance Metrics

Monitor with standard tools:

```bash
# CPU and memory usage
top -p $(pidof goimgserver)

# Disk usage
df -h /opt/goimgserver/cache

# Network connections
netstat -anp | grep :9000
```

## Maintenance

### 1. Cache Management

```bash
# Check cache size
du -sh /opt/goimgserver/cache

# Clear cache via API
curl -X POST http://localhost:9000/cmd/clear

# Manual cache cleanup (older than 30 days)
find /opt/goimgserver/cache -type f -atime +30 -delete
```

### 2. Log Rotation

Create `/etc/logrotate.d/goimgserver`:

```
/opt/goimgserver/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 goimgserver goimgserver
}
```

### 3. Updates

```bash
# Build new version
cd src
go build -o goimgserver main.go

# Stop service
sudo systemctl stop goimgserver

# Backup current binary
sudo cp /opt/goimgserver/bin/goimgserver /opt/goimgserver/bin/goimgserver.bak

# Install new binary
sudo cp goimgserver /opt/goimgserver/bin/

# Start service
sudo systemctl start goimgserver

# Verify
sudo systemctl status goimgserver
```

### 4. Backup

```bash
# Backup images
tar czf images-backup-$(date +%Y%m%d).tar.gz /opt/goimgserver/images

# Backup configuration
tar czf config-backup-$(date +%Y%m%d).tar.gz /etc/systemd/system/goimgserver.service

# Optional: backup cache (usually not necessary)
tar czf cache-backup-$(date +%Y%m%d).tar.gz /opt/goimgserver/cache
```

## Troubleshooting

See the [Troubleshooting Guide](../troubleshooting/README.md) for common issues and solutions.

## Related Documentation

- [API Documentation](../api/README.md)
- [Performance Guide](../performance/README.md)
- [Security Guide](../security/README.md)

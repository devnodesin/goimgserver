# Deployment Guide

## Overview

This document provides comprehensive deployment guidelines for goimgserver across different environments, from development to production-scale deployments.

## Development Deployment

### Local Development Setup

#### Prerequisites
```bash
# Install Go 1.21 or later
go version

# Install libvips development headers (for bimg)
# Ubuntu/Debian:
sudo apt-get install libvips-dev

# macOS:
brew install vips

# Windows:
# Download pre-compiled libvips binaries
```

#### Quick Start
```bash
# Clone repository
git clone <repository-url>
cd goimgserver

# Navigate to source directory
cd src

# Initialize Go module
go mod init goimgserver
go mod tidy

# Install dependencies
go get github.com/gin-gonic/gin
go get github.com/h2non/bimg

# Run application
go run main.go --port 9000 --imagesdir ./images --cachedir ./cache
```

#### Development Configuration
```bash
# Create directory structure
mkdir -p images cache

# Add sample images
cp /path/to/sample/images/* images/

# Optional: Initialize Git in images directory for gitupdate testing
cd images
git init
git add .
git commit -m "Initial images"
cd ..

# Run with debug logging
GOIMGSERVER_LOG_LEVEL=debug go run main.go
```

## Production Deployment

### System Requirements

#### Minimum Requirements
- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Storage**: 20GB (10GB for application + cache)
- **OS**: Linux (Ubuntu 20.04+ recommended)

#### Recommended Requirements
- **CPU**: 4+ cores
- **Memory**: 8GB+ RAM
- **Storage**: 100GB+ SSD (for cache and images)
- **OS**: Linux (Ubuntu 22.04+ recommended)

#### High-Traffic Requirements
- **CPU**: 8+ cores
- **Memory**: 16GB+ RAM
- **Storage**: 500GB+ NVMe SSD
- **Network**: 1Gbps+ connection
- **Load Balancer**: Multiple instances behind load balancer

### Binary Deployment

#### Build for Production
```bash
# Build optimized binary
cd src
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o goimgserver main.go

# Create deployment package
mkdir -p deployment/{bin,config,systemd}
cp goimgserver deployment/bin/
cp config/* deployment/config/
cp systemd/* deployment/systemd/
```

#### System Installation
```bash
# Create system user
sudo useradd --system --home /opt/goimgserver --shell /bin/false goimgserver

# Create directory structure
sudo mkdir -p /opt/goimgserver/{bin,images,cache,logs,config}
sudo chown -R goimgserver:goimgserver /opt/goimgserver

# Install binary
sudo cp goimgserver /opt/goimgserver/bin/
sudo chmod +x /opt/goimgserver/bin/goimgserver

# Set permissions
sudo chmod 755 /opt/goimgserver/images
sudo chmod 700 /opt/goimgserver/cache
sudo chmod 755 /opt/goimgserver/logs
```

#### Systemd Service
```ini
# /etc/systemd/system/goimgserver.service
[Unit]
Description=Go Image Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=goimgserver
Group=goimgserver
WorkingDirectory=/opt/goimgserver
ExecStart=/opt/goimgserver/bin/goimgserver \
    --port 9000 \
    --imagesdir /opt/goimgserver/images \
    --cachedir /opt/goimgserver/cache
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=goimgserver

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/goimgserver/cache /opt/goimgserver/logs

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable goimgserver
sudo systemctl start goimgserver
sudo systemctl status goimgserver
```

### Container Deployment

#### Dockerfile
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev vips-dev

WORKDIR /app

# Copy go mod files
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code
COPY src/ ./

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o goimgserver main.go

# Production image
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache vips ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S goimgserver && \
    adduser -u 1001 -S goimgserver -G goimgserver

# Create directory structure
RUN mkdir -p /app/{images,cache} && \
    chown -R goimgserver:goimgserver /app

# Copy binary from builder
COPY --from=builder --chown=goimgserver:goimgserver /app/goimgserver /app/

# Switch to non-root user
USER goimgserver

WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9000/health || exit 1

EXPOSE 9000

CMD ["./goimgserver", "--port", "9000", "--imagesdir", "./images", "--cachedir", "./cache"]
```

#### Docker Compose
```yaml
# docker-compose.yml
version: '3.8'

services:
  goimgserver:
    build: .
    ports:
      - "9000:9000"
    volumes:
      - ./images:/app/images:ro
      - ./cache:/app/cache
      - ./logs:/app/logs
    environment:
      - GOIMGSERVER_LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9000/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '0.5'
          memory: 1G

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - goimgserver
    restart: unless-stopped
```

#### Nginx Configuration
```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream goimgserver {
        server goimgserver:9000;
        keepalive 32;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=images:10m rate=100r/m;
    limit_req_zone $binary_remote_addr zone=commands:10m rate=10r/m;

    server {
        listen 80;
        server_name your-domain.com;

        # Redirect to HTTPS
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        # SSL configuration
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";

        # Image endpoints
        location /img/ {
            limit_req zone=images burst=20 nodelay;
            proxy_pass http://goimgserver;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            
            # Caching
            proxy_cache_valid 200 1y;
            add_header Cache-Control "public, max-age=31536000, immutable";
        }

        # Command endpoints
        location /cmd/ {
            limit_req zone=commands burst=5 nodelay;
            proxy_pass http://goimgserver;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            
            # No caching for commands
            add_header Cache-Control "no-store";
        }

        # Health check
        location /health {
            proxy_pass http://goimgserver;
            access_log off;
        }
    }
}
```

### Kubernetes Deployment

#### Deployment Manifest
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goimgserver
  labels:
    app: goimgserver
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goimgserver
  template:
    metadata:
      labels:
        app: goimgserver
    spec:
      containers:
      - name: goimgserver
        image: goimgserver:latest
        ports:
        - containerPort: 9000
        env:
        - name: GOIMGSERVER_LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "2"
        volumeMounts:
        - name: images-volume
          mountPath: /app/images
          readOnly: true
        - name: cache-volume
          mountPath: /app/cache
        livenessProbe:
          httpGet:
            path: /health
            port: 9000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 9000
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: images-volume
        persistentVolumeClaim:
          claimName: images-pvc
      - name: cache-volume
        persistentVolumeClaim:
          claimName: cache-pvc
```

#### Service and Ingress
```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: goimgserver-service
spec:
  selector:
    app: goimgserver
  ports:
  - protocol: TCP
    port: 80
    targetPort: 9000
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: goimgserver-ingress
  annotations:
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: tls-secret
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: goimgserver-service
            port:
              number: 80
```

## Configuration Management

### Environment Variables
```bash
# Application settings
export GOIMGSERVER_PORT=9000
export GOIMGSERVER_IMAGES_DIR=/opt/goimgserver/images
export GOIMGSERVER_CACHE_DIR=/opt/goimgserver/cache
export GOIMGSERVER_LOG_LEVEL=info

# Performance tuning
export GOMAXPROCS=4
export GOGC=100

# Security settings
export GOIMGSERVER_MAX_FILE_SIZE=100MB
export GOIMGSERVER_MAX_DIMENSIONS=4000
export GOIMGSERVER_RATE_LIMIT=100
```

### Configuration File
```yaml
# config/goimgserver.yaml
server:
  port: 9000
  timeout: 30s
  max_request_size: 1MB

directories:
  images: /opt/goimgserver/images
  cache: /opt/goimgserver/cache

processing:
  max_dimensions: 4000
  min_dimensions: 10
  default_quality: 75
  max_file_size: 100MB
  concurrent_limit: 10

security:
  rate_limit: 100
  cors_origins: ["*"]
  enable_cors: true

logging:
  level: info
  format: json
  output: /opt/goimgserver/logs/goimgserver.log
```

## Monitoring and Observability

### Health Checks
```bash
# Basic health check
curl -f http://localhost:9000/health || exit 1

# Detailed health check
curl -s http://localhost:9000/health | jq '.status'
```

### Metrics Collection
```bash
# Application metrics (if implemented)
curl -s http://localhost:9000/metrics

# System metrics
ps aux | grep goimgserver
du -sh /opt/goimgserver/cache
df -h /opt/goimgserver
```

### Logging Configuration
```bash
# Systemd journal
sudo journalctl -u goimgserver -f

# Log rotation
# /etc/logrotate.d/goimgserver
/opt/goimgserver/logs/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 goimgserver goimgserver
    postrotate
        systemctl restart goimgserver
    endscript
}
```

## Performance Tuning

### System Optimization
```bash
# Increase file descriptor limits
echo "goimgserver soft nofile 65536" >> /etc/security/limits.conf
echo "goimgserver hard nofile 65536" >> /etc/security/limits.conf

# Optimize kernel parameters
echo 'net.core.somaxconn = 65536' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65536' >> /etc/sysctl.conf
sysctl -p
```

### Application Tuning
```bash
# Go runtime optimization
export GOGC=100          # Garbage collection target
export GOMAXPROCS=4      # CPU cores to use
export GOMEMLIMIT=6GiB   # Memory limit
```

### Cache Optimization
```bash
# SSD optimization for cache directory
sudo mount -o noatime,discard /dev/sdb1 /opt/goimgserver/cache

# Cache cleanup script
#!/bin/bash
find /opt/goimgserver/cache -type f -atime +30 -delete
```

## Backup and Recovery

### Backup Strategy
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backup/goimgserver"
DATE=$(date +%Y%m%d_%H%M%S)

# Backup images
tar -czf "$BACKUP_DIR/images_$DATE.tar.gz" /opt/goimgserver/images

# Backup configuration
cp -r /opt/goimgserver/config "$BACKUP_DIR/config_$DATE"

# Clean old backups (keep 30 days)
find "$BACKUP_DIR" -type f -mtime +30 -delete
```

### Recovery Procedures
```bash
#!/bin/bash
# restore.sh
BACKUP_FILE=$1

# Stop service
sudo systemctl stop goimgserver

# Restore images
tar -xzf "$BACKUP_FILE" -C /

# Clear cache (will be regenerated)
rm -rf /opt/goimgserver/cache/*

# Start service
sudo systemctl start goimgserver
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check logs
sudo journalctl -u goimgserver -n 50

# Check permissions
ls -la /opt/goimgserver/
sudo -u goimgserver ls -la /opt/goimgserver/

# Check dependencies
ldd /opt/goimgserver/bin/goimgserver
```

#### High Memory Usage
```bash
# Monitor memory usage
htop -p $(pgrep goimgserver)

# Check cache size
du -sh /opt/goimgserver/cache

# Clear cache if needed
sudo systemctl stop goimgserver
rm -rf /opt/goimgserver/cache/*
sudo systemctl start goimgserver
```

#### Slow Response Times
```bash
# Check system load
top
iostat -x 1

# Check network connectivity
ss -tulpn | grep :9000

# Monitor request processing
curl -w "%{time_total}\n" -o /dev/null -s http://localhost:9000/img/test.jpg
```

### Log Analysis
```bash
# Error analysis
sudo journalctl -u goimgserver | grep ERROR

# Performance analysis
sudo journalctl -u goimgserver | grep "processing_time" | tail -100

# Security analysis
sudo journalctl -u goimgserver | grep "rate_limit\|security"
```

This deployment guide provides comprehensive instructions for deploying goimgserver across various environments with proper security, monitoring, and maintenance practices.
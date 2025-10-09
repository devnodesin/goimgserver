# Security Guide

## Overview

This guide covers security best practices, hardening, and threat mitigation for goimgserver deployments.

## Security Model

### Threat Model

**Protected Assets:**
- Server infrastructure and resources
- Image files and cache data
- API availability and performance

**Potential Threats:**
- Denial of Service (DoS) attacks
- Resource exhaustion
- Unauthorized access to administrative endpoints
- Path traversal attempts
- Cache poisoning

**Out of Scope:**
- Image content security (watermarking, DRM)
- User authentication (handled by reverse proxy)
- Network security (firewall, IDS/IPS)

## Security Hardening

### 1. File System Security

#### Directory Permissions

```bash
# Application directory (read-only for service user)
sudo chown root:root /opt/goimgserver/bin/goimgserver
sudo chmod 755 /opt/goimgserver/bin/goimgserver

# Images directory (read-only)
sudo chown -R root:goimgserver /opt/goimgserver/images
sudo chmod -R 755 /opt/goimgserver/images

# Cache directory (read-write for service user only)
sudo chown -R goimgserver:goimgserver /opt/goimgserver/cache
sudo chmod 770 /opt/goimgserver/cache

# Logs directory
sudo chown -R goimgserver:goimgserver /opt/goimgserver/logs
sudo chmod 750 /opt/goimgserver/logs
```

#### Restrict Sensitive Files

```bash
# Configuration files (if using config files)
sudo chmod 600 /opt/goimgserver/config.yml
sudo chown goimgserver:goimgserver /opt/goimgserver/config.yml

# Service files
sudo chmod 644 /etc/systemd/system/goimgserver.service
```

### 2. Network Security

#### Firewall Configuration

```bash
# UFW (Ubuntu/Debian)
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH
sudo ufw allow ssh

# Allow only from reverse proxy (if using nginx)
sudo ufw allow from 127.0.0.1 to any port 9000

# For direct access
sudo ufw allow 9000/tcp

# Enable firewall
sudo ufw enable
```

#### iptables Configuration

```bash
# Accept established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Allow HTTP/HTTPS (nginx)
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Block direct access to app port from outside
iptables -A INPUT -p tcp --dport 9000 ! -s 127.0.0.1 -j DROP

# Save rules
iptables-save > /etc/iptables/rules.v4
```

### 3. Systemd Security

Enhanced systemd service configuration:

```ini
[Service]
# Security hardening
NoNewPrivileges=true
PrivateTmp=true
PrivateDevices=true
ProtectSystem=strict
ProtectHome=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true

# Read-write paths
ReadWritePaths=/opt/goimgserver/cache
ReadWritePaths=/opt/goimgserver/logs

# Read-only paths
ReadOnlyPaths=/opt/goimgserver/images

# System call filtering
SystemCallFilter=@system-service
SystemCallFilter=~@privileged @resources
SystemCallArchitectures=native

# Resource limits
LimitNOFILE=65536
LimitNPROC=512
LimitCPU=infinity
LimitMEMLOCK=0

# Capabilities
CapabilityBoundingSet=
AmbientCapabilities=
```

### 4. Docker Security

#### Dockerfile Security

```dockerfile
# Use specific version, not latest
FROM golang:1.21-bookworm AS builder

# Run as non-root user
USER goimgserver

# Drop capabilities
--security-opt="no-new-privileges:true"

# Read-only root filesystem
--read-only
--tmpfs /tmp
```

#### Docker Compose Security

```yaml
services:
  goimgserver:
    security_opt:
      - no-new-privileges:true
      - apparmor=docker-default
    cap_drop:
      - ALL
    read_only: true
    tmpfs:
      - /tmp
    user: "1000:1000"
```

### 5. Nginx Security

#### Security Headers

```nginx
# Security headers
add_header X-Content-Type-Options "nosniff" always;
add_header X-Frame-Options "DENY" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Content-Security-Policy "default-src 'none'; img-src 'self'; style-src 'self'" always;

# HSTS (HTTPS only)
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
```

#### Rate Limiting

```nginx
# Define rate limit zones
limit_req_zone $binary_remote_addr zone=imgapi:10m rate=100r/m;
limit_req_zone $binary_remote_addr zone=imgapi_burst:10m rate=500r/m;

# Apply limits
location /img/ {
    limit_req zone=imgapi burst=10 nodelay;
    limit_req_status 429;
}
```

#### Request Size Limits

```nginx
# Limit request body size (prevent large POST attacks)
client_max_body_size 100M;
client_body_buffer_size 128k;

# Limit header size
large_client_header_buffers 4 16k;
```

#### IP Filtering

```nginx
# Whitelist for administrative endpoints
location /cmd/ {
    # Allow only internal IPs
    allow 127.0.0.1;
    allow 10.0.0.0/8;
    allow 172.16.0.0/12;
    allow 192.168.0.0/16;
    deny all;
}
```

## Rate Limiting

### Application-Level Rate Limiting

The application includes built-in rate limiting:

- **Default**: 100 requests/minute per IP
- **Burst**: 10 requests
- **Response**: HTTP 429 with Retry-After header

### Nginx Rate Limiting

```nginx
# Different limits for different endpoints
limit_req_zone $binary_remote_addr zone=img_normal:10m rate=100r/m;
limit_req_zone $binary_remote_addr zone=img_thumb:10m rate=200r/m;

location ~ ^/img/.+/\d+x\d+/webp$ {
    limit_req zone=img_normal burst=20 nodelay;
}

location ~ ^/img/.+/(150x150|200x200)/webp$ {
    # Higher limits for thumbnails
    limit_req zone=img_thumb burst=50 nodelay;
}
```

### Connection Limits

```nginx
# Limit concurrent connections per IP
limit_conn_zone $binary_remote_addr zone=addr:10m;

location /img/ {
    limit_conn addr 10;
}
```

## Input Validation

### Path Traversal Prevention

The application validates all file paths:

```go
// Built-in validation
// - Rejects paths with ..
// - Rejects absolute paths
// - Restricts to images directory
```

### Dimension Validation

```bash
# Maximum dimensions enforced
--max-dimensions=4000

# Prevents:
# - Memory exhaustion
# - CPU exhaustion
# - Unreasonably large outputs
```

### Format Validation

```go
// Only allowed formats
allowedFormats := []string{"webp", "png", "jpeg", "jpg"}
```

## SSL/TLS Configuration

### Let's Encrypt Setup

```bash
# Install certbot
sudo apt-get install -y certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx \
    -d img.example.com \
    --agree-tos \
    --email admin@example.com \
    --redirect

# Auto-renewal
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
```

### Manual SSL Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name img.example.com;
    
    # Certificates
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    
    # Modern SSL configuration
    ssl_protocols TLSv1.3 TLSv1.2;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256';
    ssl_prefer_server_ciphers off;
    
    # OCSP stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_trusted_certificate /etc/nginx/ssl/chain.pem;
    
    # Session cache
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_session_tickets off;
}
```

## Access Control

### IP Whitelisting

```nginx
# /etc/nginx/snippets/admin-whitelist.conf
allow 203.0.113.0/24;  # Office network
allow 198.51.100.50;   # VPN
deny all;

# Use in locations
location /cmd/ {
    include snippets/admin-whitelist.conf;
    proxy_pass http://goimgserver;
}
```

### Basic Authentication (Nginx)

```bash
# Create password file
sudo apt-get install -y apache2-utils
sudo htpasswd -c /etc/nginx/.htpasswd admin

# Configure nginx
location /cmd/ {
    auth_basic "Restricted";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://goimgserver;
}
```

## Monitoring and Alerting

### Log Monitoring

```bash
# Monitor for suspicious activity
sudo tail -f /var/log/nginx/access.log | grep -E "(cmd|\.\.)"

# Alert on error rate
sudo journalctl -u goimgserver -p err --since "1 minute ago" | wc -l
```

### Automated Alerts

```bash
#!/bin/bash
# /opt/goimgserver/scripts/security-monitor.sh

ERROR_THRESHOLD=10
LOG_FILE="/var/log/nginx/access.log"

# Count 429 errors in last minute
ERROR_COUNT=$(grep -c "\" 429 " "$LOG_FILE" | tail -1000)

if [ "$ERROR_COUNT" -gt "$ERROR_THRESHOLD" ]; then
    echo "High rate limit violations: $ERROR_COUNT" | \
        mail -s "Security Alert: goimgserver" admin@example.com
fi
```

### fail2ban Configuration

```ini
# /etc/fail2ban/filter.d/goimgserver.conf
[Definition]
failregex = ^<HOST> .* "(?:GET|POST) /cmd/ .* 403
            ^<HOST> .* "(?:GET|POST) .* 429

# /etc/fail2ban/jail.local
[goimgserver]
enabled = true
port = http,https
filter = goimgserver
logpath = /var/log/nginx/access.log
maxretry = 5
bantime = 3600
findtime = 600
```

## Security Checklist

### Deployment Checklist

- [ ] Service runs as non-root user
- [ ] File permissions properly configured
- [ ] Firewall rules in place
- [ ] SSL/TLS enabled and configured
- [ ] Security headers configured
- [ ] Rate limiting enabled
- [ ] Access logs enabled
- [ ] Error logs monitored
- [ ] Command endpoints restricted
- [ ] Regular security updates applied

### Nginx Checklist

- [ ] SSL/TLS with modern ciphers
- [ ] Security headers configured
- [ ] Rate limiting configured
- [ ] Connection limits configured
- [ ] Request size limits set
- [ ] IP whitelisting for admin endpoints
- [ ] Access logs enabled
- [ ] HSTS enabled (HTTPS)

### System Checklist

- [ ] OS security updates automatic
- [ ] Unnecessary services disabled
- [ ] SSH key-based authentication
- [ ] Firewall configured
- [ ] SELinux/AppArmor enabled
- [ ] Intrusion detection configured
- [ ] Log rotation configured
- [ ] Backups automated

## Incident Response

### Security Incident Procedure

1. **Detection**: Monitor logs and alerts
2. **Assessment**: Determine severity and scope
3. **Containment**: Block attacking IPs, rate limit
4. **Eradication**: Remove malicious requests from cache
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Update security measures

### Emergency Commands

```bash
# Block IP immediately
sudo ufw deny from 203.0.113.50

# Clear cache
curl -X POST http://localhost:9000/cmd/clear

# Restart service
sudo systemctl restart goimgserver

# Check for suspicious activity
sudo grep "403\|404\|429" /var/log/nginx/access.log | tail -100

# Monitor in real-time
sudo tail -f /var/log/nginx/access.log | grep -E "(cmd|\.\.)"
```

## Compliance

### Data Privacy

- Image files: Ensure compliance with data privacy laws (GDPR, CCPA)
- Logs: Retain only necessary information, anonymize IPs if required
- Cache: Implement retention policies

### Audit Logging

```bash
# Enable detailed audit logging
sudo auditctl -w /opt/goimgserver/bin/goimgserver -p x -k goimgserver_exec
sudo auditctl -w /opt/goimgserver/images -p r -k image_access
sudo auditctl -w /etc/systemd/system/goimgserver.service -p wa -k service_config
```

## Regular Security Tasks

### Daily

- Monitor error logs
- Check rate limit violations
- Review access patterns

### Weekly

- Review security alerts
- Update fail2ban rules
- Check SSL certificate expiry

### Monthly

- Security updates
- Review access control lists
- Audit log analysis
- Security scan

### Quarterly

- Security assessment
- Penetration testing (if applicable)
- Update security documentation
- Team security training

## Related Documentation

- [Deployment Guide](../deployment/README.md)
- [Performance Guide](../performance/README.md)
- [Troubleshooting Guide](../troubleshooting/README.md)

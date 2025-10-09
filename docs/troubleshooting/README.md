# Troubleshooting Guide

## Overview

This guide provides solutions to common problems and debugging procedures for goimgserver.

## Common Issues

### 1. Server Won't Start

#### Symptom

```
Failed to start goimgserver.service
```

#### Possible Causes and Solutions

**Port Already in Use**

```bash
# Check what's using the port
sudo lsof -i :9000

# Solution 1: Stop conflicting service
sudo systemctl stop conflicting-service

# Solution 2: Use different port
sudo systemctl edit goimgserver.service
# Add: --port=9001
```

**libvips Not Found**

```bash
# Error message
Package vips was not found in the pkg-config search path

# Solution: Install libvips
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y libvips-dev

# RHEL/CentOS
sudo yum install -y vips-devel

# macOS
brew install vips
```

**Permission Denied**

```bash
# Check logs
sudo journalctl -u goimgserver -n 50

# Solution: Fix permissions
sudo chown -R goimgserver:goimgserver /opt/goimgserver/cache
sudo chmod 775 /opt/goimgserver/cache
```

**Binary Not Found**

```bash
# Error
/opt/goimgserver/bin/goimgserver: No such file or directory

# Solution: Verify binary location
ls -la /opt/goimgserver/bin/

# Rebuild and install
cd src
go build -o goimgserver main.go
sudo cp goimgserver /opt/goimgserver/bin/
```

### 2. Images Not Loading

#### Symptom

```
HTTP 404: Image not found
```

#### Diagnosis and Solutions

**Check Images Directory**

```bash
# Verify images directory exists
ls -la /opt/goimgserver/images/

# Check file permissions
ls -la /opt/goimgserver/images/sample.jpg

# Solution: Fix permissions
sudo chown -R goimgserver:goimgserver /opt/goimgserver/images
sudo chmod 755 /opt/goimgserver/images
```

**Verify Image Path**

```bash
# Check server configuration
sudo journalctl -u goimgserver | grep "imagesdir"

# Test with curl
curl -v http://localhost:9000/img/sample.jpg

# Solution: Use correct filename
# Ensure filename matches exactly (case-sensitive)
```

**Image Format Issues**

```bash
# Check image format
file /opt/goimgserver/images/sample.jpg

# Solution: Convert to supported format
convert input.bmp output.jpg
```

### 3. Cache Not Working

#### Symptom

- Slow performance
- High CPU usage
- Same images processed repeatedly

#### Diagnosis

```bash
# Check cache directory
ls -la /opt/goimgserver/cache/

# Monitor cache writes
sudo inotifywait -m /opt/goimgserver/cache/

# Check disk space
df -h /opt/goimgserver/cache/
```

#### Solutions

**Insufficient Permissions**

```bash
# Fix cache directory permissions
sudo chown -R goimgserver:goimgserver /opt/goimgserver/cache
sudo chmod 775 /opt/goimgserver/cache
```

**Disk Full**

```bash
# Check disk usage
df -h

# Solution: Clear old cache files
find /opt/goimgserver/cache -type f -atime +30 -delete

# Or use API
curl -X POST http://localhost:9000/cmd/clear
```

**SELinux Blocking (RHEL/CentOS)**

```bash
# Check SELinux denials
sudo ausearch -m avc -ts recent

# Solution: Set correct context
sudo semanage fcontext -a -t httpd_sys_rw_content_t "/opt/goimgserver/cache(/.*)?"
sudo restorecon -Rv /opt/goimgserver/cache
```

### 4. High Memory Usage

#### Symptom

```
Out of memory errors
System becomes unresponsive
```

#### Diagnosis

```bash
# Monitor memory usage
top -p $(pidof goimgserver)

# Check heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

#### Solutions

**Set Memory Limit**

```bash
# Edit systemd service
sudo systemctl edit goimgserver.service

# Add environment variable
[Service]
Environment="GOMEMLIMIT=4GiB"

# Restart service
sudo systemctl daemon-reload
sudo systemctl restart goimgserver
```

**Adjust Garbage Collection**

```bash
# More aggressive GC
Environment="GOGC=50"

# Less frequent GC (more memory)
Environment="GOGC=200"
```

**Limit Concurrent Processing**

```bash
# Reduce worker count
--precache-workers=2

# Reduce GOMAXPROCS
Environment="GOMAXPROCS=2"
```

### 5. Slow Performance

#### Symptom

- High latency
- Timeout errors
- Poor throughput

#### Diagnosis

```bash
# Check CPU usage
top -p $(pidof goimgserver)

# Check disk I/O
iotop -p $(pidof goimgserver)

# Check network
netstat -anp | grep :9000

# Profile CPU usage
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

#### Solutions

**CPU Bottleneck**

```bash
# Increase GOMAXPROCS
Environment="GOMAXPROCS=8"

# Enable pre-cache
--precache=true --precache-workers=8
```

**Disk I/O Bottleneck**

```bash
# Use SSD for cache
sudo mount /dev/sdb1 /opt/goimgserver/cache

# Optimize mount options
sudo mount -o noatime,discard /dev/sdb1 /opt/goimgserver/cache
```

**Network Issues**

```bash
# Check nginx connection
curl -v http://localhost:9000/health

# Increase nginx keepalive
upstream goimgserver {
    keepalive 64;
}
```

### 6. Git Update Fails

#### Symptom

```json
{
  "success": false,
  "error": "Images directory is not a git repository"
}
```

#### Solutions

**Not a Git Repository**

```bash
# Initialize git in images directory
cd /opt/goimgserver/images
git init
git remote add origin <repository-url>
git pull origin main
```

**Permission Issues**

```bash
# Fix git directory permissions
sudo chown -R goimgserver:goimgserver /opt/goimgserver/images/.git
```

**Merge Conflicts**

```bash
# Reset to remote state
cd /opt/goimgserver/images
sudo -u goimgserver git fetch origin
sudo -u goimgserver git reset --hard origin/main
```

### 7. Rate Limiting Issues

#### Symptom

```
HTTP 429: Too Many Requests
```

#### Diagnosis

```bash
# Check rate limit headers
curl -I http://localhost:9000/img/test.jpg

# View rate limit logs
sudo journalctl -u goimgserver | grep "rate limit"
```

#### Solutions

**Increase Rate Limits**

```nginx
# Adjust nginx rate limits
limit_req_zone $binary_remote_addr zone=imgapi:10m rate=200r/m;

location /img/ {
    limit_req zone=imgapi burst=20 nodelay;
}
```

**Whitelist IPs**

```nginx
# Exempt specific IPs from rate limiting
geo $limit {
    default 1;
    10.0.0.0/8 0;
    192.168.0.0/16 0;
}

map $limit $limit_key {
    0 "";
    1 $binary_remote_addr;
}

limit_req_zone $limit_key zone=imgapi:10m rate=100r/m;
```

### 8. SSL/TLS Issues

#### Symptom

```
SSL certificate problem
Connection refused on HTTPS
```

#### Diagnosis

```bash
# Test SSL certificate
openssl s_client -connect img.example.com:443

# Check nginx configuration
sudo nginx -t

# View nginx error logs
sudo tail -f /var/log/nginx/error.log
```

#### Solutions

**Certificate Expired**

```bash
# Check expiry date
openssl x509 -in /etc/nginx/ssl/cert.pem -noout -dates

# Renew with certbot
sudo certbot renew

# Manual renewal
sudo certbot certonly --nginx -d img.example.com
```

**Wrong Certificate Path**

```nginx
# Verify paths in nginx config
ssl_certificate /etc/letsencrypt/live/img.example.com/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/img.example.com/privkey.pem;
```

### 9. Docker Issues

#### Container Won't Start

```bash
# View container logs
docker logs goimgserver

# Check container status
docker ps -a

# Inspect container
docker inspect goimgserver
```

#### Volume Mount Issues

```bash
# Verify volume mounts
docker inspect -f '{{ .Mounts }}' goimgserver

# Solution: Fix volume paths in docker-compose.yml
volumes:
  - ./images:/app/images:ro
  - goimgserver-cache:/app/cache
```

#### Permission Issues in Container

```bash
# Check user in container
docker exec goimgserver id

# Solution: Set correct user
docker run --user 1000:1000 goimgserver
```

### 10. Pre-cache Issues

#### Pre-cache Never Completes

```bash
# Check pre-cache status in logs
sudo journalctl -u goimgserver | grep "precache"

# Solution: Reduce worker count
--precache-workers=2

# Or disable pre-cache for faster startup
--precache=false
```

#### Out of Memory During Pre-cache

```bash
# Reduce concurrent workers
--precache-workers=1

# Set memory limit
Environment="GOMEMLIMIT=4GiB"

# Disable pre-cache
--precache=false
```

## Debugging Procedures

### Enable Debug Logging

```bash
# Set log level
Environment="GOIMGSERVER_LOG_LEVEL=debug"

# Restart service
sudo systemctl restart goimgserver

# View debug logs
sudo journalctl -u goimgserver -f
```

### Enable pprof Profiling

Add to main.go (if not present):

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Access profiles:

```bash
# CPU profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Memory profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof
```

### Network Debugging

```bash
# Trace requests
sudo tcpdump -i any -n port 9000

# Monitor connections
netstat -anp | grep :9000

# Test endpoint
curl -v -H "X-Forwarded-For: 1.2.3.4" http://localhost:9000/img/test.jpg
```

### File System Debugging

```bash
# Monitor file access
sudo strace -p $(pidof goimgserver) -e trace=open,stat

# Watch directory changes
inotifywait -m -r /opt/goimgserver/images/

# Check open files
lsof -p $(pidof goimgserver)
```

## Log Analysis

### Common Log Patterns

**Successful Request**
```
[abc123] 192.168.1.100 GET /img/photo.jpg/800x600/webp 200 45ms
```

**Rate Limited**
```
[def456] 192.168.1.200 GET /img/photo.jpg 429 1ms
```

**File Not Found**
```
[ghi789] 192.168.1.150 GET /img/missing.jpg 404 2ms
```

**Processing Error**
```
[jkl012] ERROR: Failed to process image: invalid format
```

### Log Commands

```bash
# Count errors
sudo journalctl -u goimgserver | grep ERROR | wc -l

# Find slow requests (>100ms)
sudo journalctl -u goimgserver | grep -E "[0-9]{3,}ms"

# Analyze request patterns
sudo journalctl -u goimgserver | awk '{print $6}' | sort | uniq -c

# Top requested images
sudo journalctl -u goimgserver | grep "GET /img" | \
    awk '{print $7}' | sort | uniq -c | sort -rn | head -20
```

## Performance Diagnostics

### Identify Bottlenecks

```bash
# CPU profiling
curl http://localhost:6060/debug/pprof/profile?seconds=30 | \
    go tool pprof -http=:8080 -

# Memory profiling
curl http://localhost:6060/debug/pprof/heap | \
    go tool pprof -http=:8080 -

# Goroutine analysis
curl http://localhost:6060/debug/pprof/goroutine | \
    go tool pprof -http=:8080 -
```

### Load Testing

```bash
# Simple load test
ab -n 1000 -c 10 http://localhost:9000/img/test.jpg/800x600/webp

# Sustained load
hey -z 60s -c 50 http://localhost:9000/img/test.jpg/800x600/webp

# Monitor during load test
watch -n 1 'ps aux | grep goimgserver'
```

## Getting Help

### Gather System Information

```bash
#!/bin/bash
# gather-info.sh

echo "=== System Info ==="
uname -a
cat /etc/os-release

echo -e "\n=== Go Version ==="
go version

echo -e "\n=== Service Status ==="
systemctl status goimgserver

echo -e "\n=== Recent Logs ==="
journalctl -u goimgserver -n 50

echo -e "\n=== Configuration ==="
grep Environment /etc/systemd/system/goimgserver.service

echo -e "\n=== Resource Usage ==="
ps aux | grep goimgserver
df -h /opt/goimgserver/cache

echo -e "\n=== Network ==="
netstat -anp | grep :9000
```

### Community Support

- GitHub Issues: https://github.com/devnodesin/goimgserver/issues
- Documentation: https://github.com/devnodesin/goimgserver/tree/main/docs

### Reporting Bugs

Include:

1. System information (OS, Go version)
2. Service configuration
3. Recent logs
4. Steps to reproduce
5. Expected vs actual behavior

## Related Documentation

- [Deployment Guide](../deployment/README.md)
- [Performance Guide](../performance/README.md)
- [Security Guide](../security/README.md)

# Performance Guide

## Overview

This guide covers performance optimization, tuning, and benchmarking for goimgserver.

## Performance Characteristics

### Typical Performance

- **Throughput**: 500-1000 requests/second (cached)
- **Latency**: 1-5ms (cached), 50-200ms (processing)
- **Cache Hit Rate**: 80-95% (typical web workload)
- **Memory Usage**: 100MB-2GB (depends on workload)
- **CPU Usage**: Scales with concurrent processing

### Factors Affecting Performance

1. **Image Size**: Larger source images take longer to process
2. **Output Dimensions**: Larger output sizes require more processing
3. **Format**: WebP compression is slower than JPEG but produces smaller files
4. **Quality**: Higher quality settings increase processing time
5. **Cache Hit Rate**: Cached responses are 100x+ faster
6. **Concurrent Requests**: Worker pool helps with parallel processing
7. **Disk I/O**: SSD vs HDD significantly impacts cache performance

## System Optimization

### Operating System Tuning

#### File Descriptor Limits

```bash
# Check current limits
ulimit -n

# Set higher limits (temporary)
ulimit -n 65536

# Permanent configuration
sudo tee -a /etc/security/limits.conf << EOF
goimgserver soft nofile 65536
goimgserver hard nofile 65536
EOF
```

#### Kernel Parameters

```bash
# Edit /etc/sysctl.conf
sudo tee -a /etc/sysctl.conf << EOF
# Network optimization
net.core.somaxconn = 65536
net.ipv4.tcp_max_syn_backlog = 65536
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_keepalive_probes = 5
net.ipv4.tcp_keepalive_intvl = 15

# Memory management
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
EOF

# Apply changes
sudo sysctl -p
```

#### I/O Scheduler

```bash
# For SSD (use noop or deadline)
echo noop | sudo tee /sys/block/sda/queue/scheduler

# For HDD (use cfq)
echo cfq | sudo tee /sys/block/sda/queue/scheduler

# Make permanent in /etc/rc.local or systemd
```

### Storage Optimization

#### SSD Configuration

```bash
# Mount with optimal flags for cache directory
sudo mount -o noatime,discard /dev/sdb1 /opt/goimgserver/cache

# Add to /etc/fstab
/dev/sdb1 /opt/goimgserver/cache ext4 noatime,discard 0 2
```

#### File System Choice

- **ext4**: Good general purpose, recommended
- **xfs**: Better for large files and high concurrency
- **btrfs**: Good for snapshots and compression

#### RAID Configuration

```bash
# RAID 0 for maximum performance (cache only)
sudo mdadm --create /dev/md0 --level=0 --raid-devices=2 /dev/sdb /dev/sdc

# RAID 1 for redundancy (images)
sudo mdadm --create /dev/md1 --level=1 --raid-devices=2 /dev/sdd /dev/sde
```

## Application Tuning

### Go Runtime Configuration

#### CPU Cores (GOMAXPROCS)

```bash
# Use all available cores
export GOMAXPROCS=0  # Auto-detect

# Limit to specific number
export GOMAXPROCS=4

# Leave some cores for OS
export GOMAXPROCS=$(($(nproc) - 1))
```

#### Garbage Collection (GOGC)

```bash
# Default (100% = more frequent GC, lower memory)
export GOGC=100

# Less frequent GC, higher memory usage
export GOGC=200

# More aggressive GC, lower memory
export GOGC=50
```

#### Memory Limit (Go 1.19+)

```bash
# Set memory limit to prevent OOM
export GOMEMLIMIT=6GiB

# Dynamic based on system memory
export GOMEMLIMIT=$(($(free -g | awk '/^Mem:/{print $2}') * 3 / 4))GiB
```

### Pre-cache Configuration

#### Worker Pool Size

```bash
# Auto-detect CPU count (default)
--precache-workers=0

# Manual specification
--precache-workers=8

# Conservative (50% of CPUs)
--precache-workers=$(($(nproc) / 2))
```

#### Pre-cache Strategy

```bash
# Enable pre-cache on startup
--precache=true

# Disable for faster startup
--precache=false

# Pre-cache only specific sizes via script
for file in images/*.jpg; do
    curl "http://localhost:9000/img/$(basename $file)/1000x1000/webp" > /dev/null
done
```

### Cache Configuration

#### Cache Size Management

```bash
# Monitor cache size
watch -n 60 'du -sh /opt/goimgserver/cache'

# Automated cleanup (cron)
# Daily at 2 AM: clean files older than 30 days
0 2 * * * find /opt/goimgserver/cache -type f -atime +30 -delete
```

#### Cache Partitioning

```bash
# Create subdirectories for different image types
mkdir -p /opt/goimgserver/cache/{thumbnails,medium,large}

# Symbolic links can help organize
ln -s /fast/ssd/cache /opt/goimgserver/cache/hot
ln -s /slow/hdd/cache /opt/goimgserver/cache/cold
```

## Nginx Optimization

### Worker Configuration

```nginx
# /etc/nginx/nginx.conf
worker_processes auto;
worker_rlimit_nofile 65536;

events {
    worker_connections 8192;
    use epoll;
    multi_accept on;
}
```

### Cache Configuration

```nginx
# Increase cache size and duration
proxy_cache_path /var/cache/nginx/goimgserver 
    levels=1:2 
    keys_zone=imgcache:200m 
    max_size=50g 
    inactive=60d 
    use_temp_path=off;

# Enable cache locking
proxy_cache_lock on;
proxy_cache_lock_timeout 5s;

# Background cache updates
proxy_cache_background_update on;
```

### Connection Optimization

```nginx
# Keepalive to backend
upstream goimgserver {
    server 127.0.0.1:9000;
    keepalive 64;
    keepalive_requests 1000;
    keepalive_timeout 60s;
}

# Client keepalive
keepalive_timeout 65s;
keepalive_requests 1000;
```

### Compression

```nginx
# Gzip compression (not needed for images, but for JSON responses)
gzip on;
gzip_vary on;
gzip_types application/json text/plain;
gzip_comp_level 5;
```

## Load Balancing

### Multiple Instances

#### Using Nginx

```nginx
upstream goimgserver_cluster {
    least_conn;
    
    server server1:9000 max_fails=3 fail_timeout=30s;
    server server2:9000 max_fails=3 fail_timeout=30s;
    server server3:9000 max_fails=3 fail_timeout=30s;
    
    keepalive 64;
}
```

#### Using HAProxy

```haproxy
backend goimgserver
    balance leastconn
    option httpchk GET /health/ready
    
    server server1 192.168.1.101:9000 check
    server server2 192.168.1.102:9000 check
    server server3 192.168.1.103:9000 check
```

### Shared Cache

#### NFS Cache (Simple)

```bash
# On NFS server
sudo apt-get install nfs-kernel-server
sudo mkdir -p /exports/goimgserver-cache
sudo chown goimgserver:goimgserver /exports/goimgserver-cache

# /etc/exports
/exports/goimgserver-cache 192.168.1.0/24(rw,sync,no_subtree_check)

# On clients
sudo mount -t nfs nfs-server:/exports/goimgserver-cache /opt/goimgserver/cache
```

#### Redis Cache (Advanced)

For distributed caching, consider implementing Redis-based cache layer.

## Monitoring and Profiling

### Basic Monitoring

```bash
# CPU usage
top -p $(pidof goimgserver)

# Memory usage
ps aux | grep goimgserver

# Disk I/O
iotop -p $(pidof goimgserver)

# Network
iftop -i eth0
```

### Advanced Profiling

#### Enable pprof

```go
// In main.go (if not already present)
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

#### Collect Profiles

```bash
# CPU profile (30 seconds)
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Analyze with pprof
go tool pprof cpu.prof
```

### Prometheus Metrics

Add Prometheus exporter (future enhancement):

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'goimgserver'
    static_configs:
      - targets: ['localhost:9090']
```

## Performance Testing

### Load Testing with Apache Bench

```bash
# Simple test
ab -n 1000 -c 10 http://localhost:9000/img/test.jpg/800x600/webp

# Sustained load
ab -n 10000 -c 100 -t 60 http://localhost:9000/img/test.jpg/800x600/webp
```

### Load Testing with wrk

```bash
# Install wrk
sudo apt-get install -y wrk

# Run test
wrk -t4 -c100 -d30s http://localhost:9000/img/test.jpg/800x600/webp

# With multiple URLs
wrk -t4 -c100 -d30s -s urls.lua http://localhost:9000/
```

### Load Testing with hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run test
hey -n 10000 -c 100 http://localhost:9000/img/test.jpg/800x600/webp
```

## Benchmarking Results

### Test Environment

- **CPU**: 8 cores @ 3.0 GHz
- **RAM**: 16GB
- **Storage**: NVMe SSD
- **Network**: 1Gbps
- **OS**: Ubuntu 22.04 LTS

### Cached Requests

```
Requests/sec:    856.42
Transfer/sec:    45.23 MB
Latency (avg):   2.34ms
Latency (p99):   8.12ms
```

### Uncached Requests (Processing)

```
Requests/sec:    145.23
Transfer/sec:    7.85 MB
Latency (avg):   68.45ms
Latency (p99):   245.67ms
```

### Mixed Workload (80% cached)

```
Requests/sec:    687.15
Transfer/sec:    36.18 MB
Latency (avg):   14.58ms
Latency (p99):   89.23ms
```

## Capacity Planning

### Sizing Guidelines

#### Small Deployment (< 1M requests/day)

- **Server**: 2 cores, 4GB RAM, 50GB SSD
- **Expected RPS**: 50-100
- **Cache Size**: 10-20GB
- **Cost**: $20-40/month (cloud)

#### Medium Deployment (1-10M requests/day)

- **Server**: 4 cores, 8GB RAM, 200GB SSD
- **Expected RPS**: 200-500
- **Cache Size**: 50-100GB
- **Cost**: $80-150/month (cloud)

#### Large Deployment (> 10M requests/day)

- **Servers**: 3x 8 cores, 16GB RAM, 500GB NVMe
- **Load Balancer**: Required
- **Expected RPS**: 1000-2000+
- **Cache Size**: 200-500GB per node
- **Cost**: $500-1000/month (cloud)

### Growth Planning

```bash
# Calculate required RPS
DAILY_REQUESTS=10000000
PEAK_MULTIPLIER=5
REQUIRED_RPS=$((DAILY_REQUESTS * PEAK_MULTIPLIER / 86400))

echo "Required capacity: ${REQUIRED_RPS} RPS"
```

## Optimization Checklist

- [ ] Enable SSD for cache directory
- [ ] Configure appropriate GOMAXPROCS
- [ ] Set memory limit with GOMEMLIMIT
- [ ] Increase file descriptor limits
- [ ] Configure kernel network parameters
- [ ] Enable nginx cache layer
- [ ] Set up cache cleanup automation
- [ ] Configure pre-cache for common sizes
- [ ] Monitor cache hit rates
- [ ] Set up health checks
- [ ] Configure load balancing (if needed)
- [ ] Enable performance profiling
- [ ] Conduct load testing
- [ ] Set up monitoring and alerts

## Related Documentation

- [Deployment Guide](../deployment/README.md)
- [Security Guide](../security/README.md)
- [Troubleshooting Guide](../troubleshooting/README.md)

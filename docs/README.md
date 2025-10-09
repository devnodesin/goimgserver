# goimgserver Documentation

Welcome to the comprehensive documentation for goimgserver - a dynamic image processing and transformation service.

## Table of Contents

### Getting Started

- [Main README](../README.md) - Project overview and quick start
- [Basic Usage](api/examples/basic_usage.md) - Simple image processing examples
- [Advanced Usage](api/examples/advanced_usage.md) - Advanced patterns and optimization

### API Reference

- [API Documentation](api/README.md) - Complete API reference
- [OpenAPI Specification](api/openapi.yaml) - Machine-readable API spec
- [Basic Examples](api/examples/basic_usage.md) - Common use cases
- [Advanced Examples](api/examples/advanced_usage.md) - Complex scenarios

### Deployment

- [Deployment Guide](deployment/README.md) - Complete deployment instructions
- **Systemd Deployment**
  - [Service File](deployment/systemd/goimgserver.service)
  - [Installation Script](deployment/systemd/install.sh)
- **Docker Deployment**
  - [Dockerfile](deployment/docker/Dockerfile)
  - [Docker Compose](deployment/docker/docker-compose.yml)
- **Reverse Proxy**
  - [Nginx Configuration](deployment/nginx/goimgserver.conf)

### Operations

- [Performance Guide](performance/README.md) - Optimization and tuning
- [Security Guide](security/README.md) - Security hardening
- [Troubleshooting Guide](troubleshooting/README.md) - Common issues

### Design Documents

Located in [../design/](../design/):

- [Overview](../design/00-overview.md) - Architecture overview
- [TDD Methodology](../design/01-tdd-methodology.md) - Testing approach
- [Endpoints](../design/02-endpoints.md) - Endpoint design
- [URL Parsing](../design/03-url-parsing.md) - URL parsing logic
- [File Resolution](../design/04-file-resolution.md) - File handling
- [Default Image](../design/05-default-image.md) - Default image handling
- [API Specification](../design/06-api-specification.md) - Original API spec
- [Security](../design/07-security.md) - Security design
- [Deployment](../design/08-deployment.md) - Deployment design

## Quick Navigation by Task

### I want to...

**Deploy goimgserver**
→ Start with [Deployment Guide](deployment/README.md)

**Optimize performance**
→ Read [Performance Guide](performance/README.md)

**Secure my installation**
→ Follow [Security Guide](security/README.md)

**Fix an issue**
→ Check [Troubleshooting Guide](troubleshooting/README.md)

**Integrate with my app**
→ See [API Documentation](api/README.md) and [Basic Usage](api/examples/basic_usage.md)

**Use Docker**
→ Go to [Docker Deployment](deployment/docker/)

**Set up systemd**
→ See [Systemd Deployment](deployment/systemd/)

**Configure nginx**
→ Use [Nginx Configuration](deployment/nginx/goimgserver.conf)

## Documentation Structure

```
docs/
├── README.md                    # This file
├── api/
│   ├── README.md               # API reference
│   ├── openapi.yaml            # OpenAPI 3.0 specification
│   └── examples/
│       ├── basic_usage.md      # Basic usage examples
│       └── advanced_usage.md   # Advanced usage examples
├── deployment/
│   ├── README.md               # Deployment guide
│   ├── systemd/
│   │   ├── goimgserver.service # Systemd service file
│   │   └── install.sh         # Installation script
│   ├── docker/
│   │   ├── Dockerfile         # Docker image definition
│   │   └── docker-compose.yml # Docker Compose config
│   └── nginx/
│       └── goimgserver.conf   # Nginx reverse proxy config
├── performance/
│   └── README.md               # Performance optimization
├── security/
│   └── README.md               # Security hardening
└── troubleshooting/
    └── README.md               # Troubleshooting guide
```

## Testing

All documentation is validated through automated tests. See [docs_test.go](docs_test.go) for the test suite that ensures:

- API examples work correctly
- Configuration examples are valid
- Deployment scripts execute properly
- Code examples compile and run
- Troubleshooting scenarios are accurate

Run tests:

```bash
cd docs
go test -v
```

## Contributing

When contributing documentation:

1. **Follow TDD**: Write tests first for new examples
2. **Test examples**: Ensure all code examples work
3. **Validate configs**: Test configuration examples
4. **Keep current**: Update docs when code changes
5. **Be clear**: Use simple language and examples

See [../design/01-tdd-methodology.md](../design/01-tdd-methodology.md) for the project's TDD approach.

## Getting Help

- **Issues**: https://github.com/devnodesin/goimgserver/issues
- **Discussions**: Check existing issues and design documents
- **Documentation**: Search this documentation
- **Troubleshooting**: See [troubleshooting guide](troubleshooting/README.md)

## License

This documentation is part of goimgserver and is licensed under the MIT License.
See [../LICENSE](../LICENSE) for details.

---

**Quick Links:**
[API](api/README.md) | 
[Deploy](deployment/README.md) | 
[Performance](performance/README.md) | 
[Security](security/README.md) | 
[Troubleshoot](troubleshooting/README.md)

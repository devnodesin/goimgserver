# Development Container Configuration

This directory contains the configuration for VS Code Development Containers (devcontainers), providing a fully configured development environment for the goimgserver project.

## Features

The devcontainer provides:

- **Go 1.24** development environment (Debian Bookworm-based)
- **libvips** development libraries for image processing (required by bimg)
- **Essential Go tools**: gopls, delve debugger, staticcheck, goimports
- **VS Code extensions**: Go, GitHub Copilot, Docker, GitLens, and more
- **Pre-configured settings**: Format on save, organize imports, linting
- **Port forwarding**: Automatic forwarding of port 9000 for the server
- **Persistent volumes**: Mounts for cache and images directories

## Prerequisites

### For VS Code:
1. Install [Visual Studio Code](https://code.visualstudio.com/)
2. Install the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
3. Install [Docker Desktop](https://www.docker.com/products/docker-desktop) (Windows/Mac) or Docker Engine (Linux)

### For GitHub Codespaces:
- No local installation required! Just click "Code" → "Codespaces" → "Create codespace on main"

## Getting Started

### Option 1: VS Code (Local)

1. Open the project folder in VS Code
2. When prompted, click **"Reopen in Container"**
   - Or press `F1` and select "Dev Containers: Reopen in Container"
3. Wait for the container to build and configure (first time takes 3-5 minutes)
4. The development environment is ready when you see "Development environment ready!" in the terminal

### Option 2: GitHub Codespaces (Cloud)

1. Navigate to the repository on GitHub
2. Click the **"Code"** button
3. Select the **"Codespaces"** tab
4. Click **"Create codespace on main"** (or your current branch)
5. Wait for the environment to initialize

## What Happens During Setup

The devcontainer automatically performs these steps:

1. **Container Creation** (`onCreateCommand`):
   - Updates apt package lists
   - Installs libvips-dev and pkg-config (required for image processing)
   - Installs Go development tools (gopls, delve, staticcheck, goimports)

2. **Post-Creation** (`postCreateCommand`):
   - Changes to the `src` directory
   - Downloads Go module dependencies
   - **Builds the project** to verify everything works

3. **Post-Start** (`postStartCommand`):
   - Displays "Development environment ready!" message

## Building and Running

Once inside the devcontainer:

```bash
# Navigate to source directory
cd src

# Build the application
go build

# Run the application
go run main.go --port 9000 --imagesdir ./images --cachedir ./cache

# Run tests
go test ./...

# Run tests with coverage
../run_test.sh --coverage
```

## Accessing the Server

The devcontainer automatically forwards port 9000. Once the server is running:

- **Local VS Code**: Access at `http://localhost:9000`
- **GitHub Codespaces**: VS Code will show a notification with the forwarded URL

## Environment Variables

The devcontainer sets these environment variables:

- `GOMAXPROCS=4`: Limits Go to 4 CPU cores
- `GOGC=100`: Standard Go garbage collection target
- `CGO_ENABLED=1`: Enables C bindings (required for libvips/bimg)

## Persistent Data

The following directories are mounted for persistent storage:

- `src/cache`: Image cache directory
- `src/images`: Source images directory

These directories persist across container rebuilds.

## Troubleshooting

### Build Fails with "vips.pc not found"

This should not happen if using the devcontainer, but if it does:

```bash
sudo apt-get update
sudo apt-get install -y libvips-dev pkg-config
```

### Go Tools Not Working

Reinstall Go tools:

```bash
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install golang.org/x/tools/cmd/goimports@latest
```

### Container Rebuild

If you need to rebuild the container from scratch:

1. Press `F1`
2. Select "Dev Containers: Rebuild Container"
3. Wait for rebuild to complete

## Customization

You can customize the devcontainer by editing `.devcontainer/devcontainer.json`:

- **Add more VS Code extensions**: Update the `extensions` array
- **Change VS Code settings**: Modify the `settings` object
- **Add more tools**: Update the `onCreateCommand` section
- **Adjust environment variables**: Modify the `containerEnv` section

## VS Code Extensions Included

- **golang.go**: Go language support with IntelliSense
- **GitHub.copilot**: AI-powered code completion
- **GitHub.copilot-chat**: AI chat assistant
- **ms-azuretools.vscode-docker**: Docker container management
- **eamodio.gitlens**: Enhanced Git capabilities
- **streetsidesoftware.code-spell-checker**: Spell checking

## Additional Resources

- [VS Code Dev Containers Documentation](https://code.visualstudio.com/docs/devcontainers/containers)
- [GitHub Codespaces Documentation](https://docs.github.com/en/codespaces)
- [Project README](../README.md)
- [Deployment Guide](../docs/deployment/README.md)

## Support

If you encounter issues with the devcontainer setup:

1. Check the container logs in VS Code's terminal
2. Try rebuilding the container
3. Verify Docker is running properly
4. Check the [troubleshooting section](#troubleshooting) above

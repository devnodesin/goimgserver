#!/bin/bash
set -e

# goimgserver Installation Script
# This script installs goimgserver as a systemd service

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo_error "Please run as root (use sudo)"
    exit 1
fi

# Configuration
INSTALL_DIR="/opt/goimgserver"
SERVICE_USER="goimgserver"
SERVICE_GROUP="goimgserver"
BINARY_PATH="/opt/goimgserver/bin/goimgserver"
SERVICE_FILE="/etc/systemd/system/goimgserver.service"

echo_info "Starting goimgserver installation..."

# Create service user
if ! id "$SERVICE_USER" &>/dev/null; then
    echo_info "Creating service user: $SERVICE_USER"
    useradd --system --no-create-home --shell /bin/false "$SERVICE_USER"
else
    echo_info "Service user already exists: $SERVICE_USER"
fi

# Create installation directories
echo_info "Creating installation directories..."
mkdir -p "$INSTALL_DIR"/{bin,images,cache,logs}

# Set ownership
echo_info "Setting directory permissions..."
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
chmod 755 "$INSTALL_DIR"
chmod 755 "$INSTALL_DIR/bin"
chmod 775 "$INSTALL_DIR/images"
chmod 775 "$INSTALL_DIR/cache"
chmod 775 "$INSTALL_DIR/logs"

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo_error "Binary not found at: $BINARY_PATH"
    echo_info "Please build and copy the binary to: $BINARY_PATH"
    exit 1
fi

# Make binary executable
chmod 755 "$BINARY_PATH"

# Install systemd service file
echo_info "Installing systemd service file..."
if [ -f "goimgserver.service" ]; then
    cp goimgserver.service "$SERVICE_FILE"
    chmod 644 "$SERVICE_FILE"
else
    echo_error "Service file not found: goimgserver.service"
    exit 1
fi

# Reload systemd
echo_info "Reloading systemd daemon..."
systemctl daemon-reload

# Enable service
echo_info "Enabling goimgserver service..."
systemctl enable goimgserver.service

# Check libvips installation
if ! pkg-config --exists vips; then
    echo_warn "libvips not found. Installing..."
    if command -v apt-get &> /dev/null; then
        apt-get update
        apt-get install -y libvips-dev
    elif command -v yum &> /dev/null; then
        yum install -y vips-devel
    else
        echo_error "Could not install libvips automatically. Please install manually."
        exit 1
    fi
fi

echo_info "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Copy your images to: $INSTALL_DIR/images/"
echo "2. Start the service: systemctl start goimgserver"
echo "3. Check status: systemctl status goimgserver"
echo "4. View logs: journalctl -u goimgserver -f"
echo ""
echo "Service will start automatically on boot."

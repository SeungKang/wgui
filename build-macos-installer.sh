#!/bin/bash

set -e

# Configuration
APP_NAME="wgui"
VERSION="${VERSION:-0.0.1}"
GROUP="com.github.SeungKang"
APPLICATION_ID="${GROUP}.${APP_NAME}"

# Directories
SCRIPT_PATH=$(realpath "$0")
SCRIPT_DIR=$(dirname "${SCRIPT_PATH}")
BUILD_DIR="${SCRIPT_DIR}/build"
RESOURCES_DIR="${SCRIPT_DIR}/resources"
INSTALLER_RESOURCES_DIR="${RESOURCES_DIR}/installer"
APP_CLEANROOM_DIR="${BUILD_DIR}/app-cleanroom"
INSTALLER_CLEANROOM_DIR="${BUILD_DIR}/installer-cleanroom"
MACOS_APP_DIR="${APP_CLEANROOM_DIR}/${APP_NAME}.app"

# Output files
INSTALLER_PREFIX="${APP_NAME}-${VERSION}"
FINAL_PKG="${BUILD_DIR}/${INSTALLER_PREFIX}.pkg"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cleanup() {
    if [ -f "${INSTALLER_CLEANROOM_DIR}/application-files.pkg" ]; then
        rm -f "${INSTALLER_CLEANROOM_DIR}/application-files.pkg"
    fi
}

trap cleanup EXIT

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    log_error "This script must be run on macOS"
    exit 1
fi

# Check for required tools
for tool in pkgbuild productbuild; do
    if ! command -v $tool &> /dev/null; then
        log_error "$tool is required but not installed"
        exit 1
    fi
done

# Verify .app bundle exists
if [ ! -d "$MACOS_APP_DIR" ]; then
    log_error ".app bundle not found: $MACOS_APP_DIR"
    log_error "Please run build-app.sh first to create the .app bundle"
    exit 1
fi

# Create installer cleanroom directory
log_step "Creating installer directories..."
mkdir -p "$INSTALLER_CLEANROOM_DIR"

# Build installer package
log_step "Preparing installer resources..."

# Verify installer resources exist
MACOS_INSTALLER_FILES="${INSTALLER_RESOURCES_DIR}/macos"
if [ ! -d "$MACOS_INSTALLER_FILES" ]; then
    log_error "Installer resources directory not found: $MACOS_INSTALLER_FILES"
    exit 1
fi

# Copy installer resources
cp -r "${MACOS_INSTALLER_FILES}/"* "$INSTALLER_CLEANROOM_DIR/"

# Verify required files exist
COMPONENT_PLIST="${INSTALLER_CLEANROOM_DIR}/component.plist"
DISTRIBUTION_XML="${INSTALLER_CLEANROOM_DIR}/distribution.xml"

if [ ! -f "$COMPONENT_PLIST" ]; then
    log_error "component.plist not found in installer resources"
    exit 1
fi

if [ ! -f "$DISTRIBUTION_XML" ]; then
    log_error "distribution.xml not found in installer resources"
    exit 1
fi

# Update component.plist
log_step "Updating installer configuration..."
sed -i '' "s/APPLICATION_NAME/${APP_NAME}/g" "$COMPONENT_PLIST"

# Update distribution.xml
sed -i '' "s/INSTALLER_TITLE/${APP_NAME} ${VERSION}/g" "$DISTRIBUTION_XML"
sed -i '' "s/PACKAGE_REF_ID/${APPLICATION_ID}/g" "$DISTRIBUTION_XML"

# Build component package
APPLICATION_FILES_PKG="${INSTALLER_CLEANROOM_DIR}/application-files.pkg"

log_step "Running pkgbuild..."
pkgbuild \
    --version "$VERSION" \
    --root "$APP_CLEANROOM_DIR" \
    --component-plist "$COMPONENT_PLIST" \
    --install-location "/Applications" \
    "$APPLICATION_FILES_PKG"

if [ ! -f "$APPLICATION_FILES_PKG" ]; then
    log_error "Failed to create component package"
    exit 1
fi

# Build final product package
log_step "Running productbuild..."
productbuild \
    --version "$VERSION" \
    --distribution "$DISTRIBUTION_XML" \
    --resources "$INSTALLER_CLEANROOM_DIR" \
    --package-path "$INSTALLER_CLEANROOM_DIR" \
    "$FINAL_PKG"

if [ ! -f "$FINAL_PKG" ]; then
    log_error "Failed to create installer package"
    exit 1
fi

log_info "✓ Installer package created successfully"

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Build Complete!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "Installer: ${FINAL_PKG}"
echo "Version:   ${VERSION}"
echo ""
echo "To install, run:"
echo "  sudo installer -pkg \"${FINAL_PKG}\" -target /"
echo ""

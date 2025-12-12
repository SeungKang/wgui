#!/bin/bash

set -e

# Configuration
APP_NAME="wgui"
VERSION="${VERSION:0.0.1}"
GROUP="com.github.SeungKang"
APPLICATION_ID="${GROUP}.${APP_NAME}"
PUBLISHER="Seung Kang"
PROJECT_URL="https://github.com/SeungKang/${APP_NAME}"
BUNDLE_SIGNATURE="${APP_NAME}"

# Directories
SCRIPT_PATH=$(realpath "$0")
SCRIPT_DIR=$(dirname "${SCRIPT_PATH}")
BUILD_DIR="${SCRIPT_DIR}/build"
RESOURCES_DIR="${SCRIPT_DIR}/resources"
APP_RESOURCES_DIR="${RESOURCES_DIR}/application"
INSTALLER_RESOURCES_DIR="${RESOURCES_DIR}/installer"
APP_CLEANROOM_DIR="${BUILD_DIR}/app-cleanroom"
INSTALLER_CLEANROOM_DIR="${BUILD_DIR}/installer-cleanroom"
MACOS_APP_DIR="${APP_CLEANROOM_DIR}/${APP_NAME}.app"

# Output files
INSTALLER_PREFIX="${APP_NAME}-${VERSION}"
FINAL_PKG="${BUILD_DIR}/${INSTALLER_PREFIX}.pkg"

# Go build flags
LDFLAGS="-X main.version=${VERSION}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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
for tool in go pkgbuild productbuild; do
    if ! command -v $tool &> /dev/null; then
        log_error "$tool is required but not installed"
        exit 1
    fi
done

# Create build directories
log_step "Creating build directories..."
mkdir -p "$BUILD_DIR"
mkdir -p "$APP_CLEANROOM_DIR"
mkdir -p "$INSTALLER_CLEANROOM_DIR"

# Build Go executable for macOS
log_step "Building Go executable for macOS (amd64)..."
cd "$SCRIPT_DIR"
GOOS=darwin GOARCH=arm64 go build \
    -ldflags "$LDFLAGS" \
    -o "${BUILD_DIR}/${APP_NAME}-darwin-arm64"

if [ ! -f "${BUILD_DIR}/${APP_NAME}-darwin-arm64" ]; then
    log_error "Failed to build executable"
    exit 1
fi

log_info "✓ Executable built successfully"

# Create .app bundle structure
log_step "Creating macOS .app bundle..."
CONTENTS_DIR="${MACOS_APP_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR_APP="${CONTENTS_DIR}/Resources"

mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR_APP"

# Create PkgInfo
echo -n "APPL${BUNDLE_SIGNATURE}" > "${CONTENTS_DIR}/PkgInfo"

# Copy executable
cp "${BUILD_DIR}/${APP_NAME}-darwin-arm64" "${MACOS_DIR}/${APP_NAME}"
chmod +x "${MACOS_DIR}/${APP_NAME}"

# Copy and update Info.plist
log_step "Updating Info.plist..."
INFO_PLIST_SOURCE="${APP_RESOURCES_DIR}/macos/Info.plist"

if [ ! -f "$INFO_PLIST_SOURCE" ]; then
    log_error "Info.plist not found: $INFO_PLIST_SOURCE"
    exit 1
fi

cp "$INFO_PLIST_SOURCE" "${CONTENTS_DIR}/Info.plist"

# Replace placeholders in Info.plist
sed -i '' "s/APP_NAME/${APP_NAME}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/IDENTIFIER/${APPLICATION_ID}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/EXECUTABLE_NAME/${APP_NAME}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/VERSION/${VERSION}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/BUNDLE_SIGNATURE/${BUNDLE_SIGNATURE}/g" "${CONTENTS_DIR}/Info.plist"

log_info "✓ .app bundle created successfully"

# Build installer package
log_step "Building installer package..."

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

# Cleanup
rm -f "$APPLICATION_FILES_PKG"
rm -rf "$MACOS_APP_DIR"

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

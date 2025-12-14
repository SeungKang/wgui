#!/bin/bash

if [[ -z "${VERSION}" ]]
then
    echo 'the VERSION environment variable must be set'
    exit 1
fi

set -eux

# Configuration
APP_NAME="wgui"
VERSION="${VERSION##*v}"
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

cleanup() {
    if [ -f "${INSTALLER_CLEANROOM_DIR}/application-files.pkg" ]; then
        rm -f "${INSTALLER_CLEANROOM_DIR}/application-files.pkg"
    fi
}

trap cleanup EXIT

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    exit 1
fi

# Check for required tools
for tool in pkgbuild productbuild; do
    if ! command -v $tool &> /dev/null; then
        exit 1
    fi
done

# Verify .app bundle exists
if [ ! -d "$MACOS_APP_DIR" ]; then
    exit 1
fi

# Create installer cleanroom directory
mkdir -p "$INSTALLER_CLEANROOM_DIR"

# Build installer package
# Verify installer resources exist
MACOS_INSTALLER_FILES="${INSTALLER_RESOURCES_DIR}/macos"
if [ ! -d "$MACOS_INSTALLER_FILES" ]; then
    exit 1
fi

# Copy installer resources
cp -r "${MACOS_INSTALLER_FILES}/"* "$INSTALLER_CLEANROOM_DIR/"

# Verify required files exist
COMPONENT_PLIST="${INSTALLER_CLEANROOM_DIR}/component.plist"
DISTRIBUTION_XML="${INSTALLER_CLEANROOM_DIR}/distribution.xml"

if [ ! -f "$COMPONENT_PLIST" ]; then
    exit 1
fi

if [ ! -f "$DISTRIBUTION_XML" ]; then
    exit 1
fi

# Update component.plist
sed -i '' "s/APPLICATION_NAME/${APP_NAME}/g" "$COMPONENT_PLIST"

# Update distribution.xml
sed -i '' "s/INSTALLER_TITLE/${APP_NAME} ${VERSION}/g" "$DISTRIBUTION_XML"
sed -i '' "s/PACKAGE_REF_ID/${APPLICATION_ID}/g" "$DISTRIBUTION_XML"

# Build component package
APPLICATION_FILES_PKG="${INSTALLER_CLEANROOM_DIR}/application-files.pkg"

pkgbuild \
    --version "$VERSION" \
    --root "$APP_CLEANROOM_DIR" \
    --component-plist "$COMPONENT_PLIST" \
    --install-location "/Applications" \
    "$APPLICATION_FILES_PKG" 1>&2

if [ ! -f "$APPLICATION_FILES_PKG" ]; then
    exit 1
fi

# Build final product package
productbuild \
    --version "$VERSION" \
    --distribution "$DISTRIBUTION_XML" \
    --resources "$INSTALLER_CLEANROOM_DIR" \
    --package-path "$INSTALLER_CLEANROOM_DIR" \
    "$FINAL_PKG" 1>&2

if [ ! -f "$FINAL_PKG" ]; then
    exit 1
fi

echo "${FINAL_PKG}"

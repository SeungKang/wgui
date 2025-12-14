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
BUNDLE_SIGNATURE="${APP_NAME}"

# Directories
SCRIPT_PATH=$(realpath "$0")
SCRIPT_DIR=$(dirname "${SCRIPT_PATH}")
BUILD_DIR="${SCRIPT_DIR}/build"
RESOURCES_DIR="${SCRIPT_DIR}/resources"
APP_RESOURCES_DIR="${RESOURCES_DIR}/application"
APP_CLEANROOM_DIR="${BUILD_DIR}/app-cleanroom"
MACOS_APP_DIR="${APP_CLEANROOM_DIR}/${APP_NAME}.app"

# Go build flags
LDFLAGS="-X main.version=${VERSION}"

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    exit 1
fi

# Check for required tools
for tool in go; do
    if ! command -v $tool &> /dev/null; then
        exit 1
    fi
done

# Create build directories
mkdir -p "$BUILD_DIR"
mkdir -p "$APP_CLEANROOM_DIR"

# Build Go executable for macOS
cd "$SCRIPT_DIR"
GOOS=darwin GOARCH=arm64 go build \
    -ldflags "$LDFLAGS" \
    -o "${BUILD_DIR}/${APP_NAME}-darwin-arm64"

if [ ! -f "${BUILD_DIR}/${APP_NAME}-darwin-arm64" ]; then
    exit 1
fi

# Install wgu dependency
GOOS=darwin GOARCH=arm64 GOBIN="${BUILD_DIR}" go install gitlab.com/stephen-fox/wgu@v0.0.11

if [ ! -f "${BUILD_DIR}/wgu" ]; then
    exit 1
fi

# Create .app bundle structure
CONTENTS_DIR="${MACOS_APP_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR_APP="${CONTENTS_DIR}/Resources"

mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR_APP"

# Create PkgInfo
echo -n "APPL${BUNDLE_SIGNATURE}" > "${CONTENTS_DIR}/PkgInfo"

# Copy wgui executable
cp "${BUILD_DIR}/${APP_NAME}-darwin-arm64" "${MACOS_DIR}/${APP_NAME}"
chmod +x "${MACOS_DIR}/${APP_NAME}"

# Copy wgu executable
cp "${BUILD_DIR}/wgu" "${MACOS_DIR}/wgu"
chmod +x "${MACOS_DIR}/wgu"

# Copy icon if it exists
ICON_FILE="${APP_RESOURCES_DIR}/macos/app.icns"
if [ -f "$ICON_FILE" ]; then
    cp "$ICON_FILE" "$RESOURCES_DIR_APP/"
fi

# Copy and update Info.plist
INFO_PLIST_SOURCE="${APP_RESOURCES_DIR}/macos/Info.plist"

if [ ! -f "$INFO_PLIST_SOURCE" ]; then
    exit 1
fi

cp "$INFO_PLIST_SOURCE" "${CONTENTS_DIR}/Info.plist"

# Replace placeholders in Info.plist
sed -i '' "s/APP_NAME/${APP_NAME}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/IDENTIFIER/${APPLICATION_ID}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/EXECUTABLE_NAME/${APP_NAME}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/VERSION/${VERSION}/g" "${CONTENTS_DIR}/Info.plist"
sed -i '' "s/BUNDLE_SIGNATURE/${BUNDLE_SIGNATURE}/g" "${CONTENTS_DIR}/Info.plist"

# Add icon name to Info.plist if icon exists
if [ -f "$ICON_FILE" ]; then
    sed -i '' "s/ICON_NAME/app.icns/g" "${CONTENTS_DIR}/Info.plist"
else
    sed -i '' "s/ICON_NAME//g" "${CONTENTS_DIR}/Info.plist"
fi

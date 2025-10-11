#!/bin/bash

# Example
# $ VERSION=0.0.1 ./build.sh

if [[ -z "${VERSION}" ]]
then
    echo 'the VERSION environment variable must be set'
    exit 1
fi

set -ex

VERSION="${VERSION##*v}"

script_path=$(realpath "$0")
script_dir=$(dirname "${script_path}")

build_dir="${script_dir}/build/app"
mkdir -p "${build_dir}"

GOOS="windows" GOARCH="amd64" go build -o "${build_dir}/wgui.exe" -ldflags "-H=windowsgui -X main.version=${VERSION}"
GOOS="windows" GOARCH="amd64" GOBIN="${build_dir}" go install -ldflags -H=windowsgui gitlab.com/stephen-fox/wgu@v0.0.8

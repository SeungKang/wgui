#!/bin/bash

set -ex

script_path=$(realpath "$0")
script_dir=$(dirname "${script_path}")

build_dir="${script_dir}/build/app"
mkdir -p "${build_dir}"

GOOS="windows" GOARCH="amd64" go build -o "${build_dir}/wgui.exe" -ldflags -H=windowsgui
GOOS="windows" GOARCH="amd64" GOBIN="${build_dir}" go install -ldflags -H=windowsgui gitlab.com/stephen-fox/wgu@v0.0.8

#!/bin/bash

# Example
# $ VERSION=0.0.1 ./package.sh

if [[ -z "${VERSION}" ]]
then
    echo 'the VERSION environment variable must be set'
    exit 1
fi

set -eux

VERSION="${VERSION##*v}"

build_dir='build'

cp resources/windows/installer/installer.iss "${build_dir}/"

# TODO: Add license for Windows installer.
# '/DLicenseFileOverride=license_en_US.txt'
innosetup_args=$(cat <<-END
/DApplicationFilesPath=app
/DOutputPath=.
/DAppNameOverride=wgui
/DAppPublisherOverride=Buh
/DAppURLOverride=https://github.com/SeungKang/wgui
/DAppExeNameOverride=wgui.exe
/DAppVersionOverride=${VERSION}
/DOutputBaseFilenameOverride=wgui-installer-${VERSION}
installer.iss
END
)

innosetup_args="\"${innosetup_args//$'\n'/\" \"}\""

bat_file_path="${build_dir}/package.bat"
echo "ISCC.exe ${innosetup_args}" > "${bat_file_path}"
(cd "${build_dir}" && cmd //c "${bat_file_path##*/}") 1>&2

find "${build_dir}" -maxdepth 1 -type f -iname '*-installer*'

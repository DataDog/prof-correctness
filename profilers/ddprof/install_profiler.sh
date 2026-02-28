#!/bin/bash

set -euo pipefail

echoerr() { echo "$@" 1>&2; }

look_in_folder() {
    local folder=$1
    if [ ! -d ${folder} ]; then
        echoerr "No ${folder} available"
        return 1
    fi
    echoerr ${folder}
    ls ${folder}
    ddprof_name=$(ls -1 ${folder}/ddprof*.xz  2> /dev/null || true)
    if [ "$(echo $ddprof_name | wc -l)" -ge "2" ]; then
        echoerr "Clean up the folder in ${folder}"
        exit 1
    fi
    # Look for unpacked version
    if [ -z "${ddprof_name}" ] || [ ! -e "${ddprof_name}" ]; then
        ddprof_name=$(ls -1 ${folder}/ddprof  2> /dev/null || true)
    fi
    echoerr "using ${ddprof_name}"
    if [ -z "${ddprof_name}" ] || [ ! -e "${ddprof_name}" ]; then
        return 1
    fi
    cp ${ddprof_name} ./
    return 0
}

look_in_s3() {
    local ddprof_path=${1-""}
    local binaries_url="https://binaries.ddbuild.io/"
    if curl --output /dev/null --connect-timeout 2 --silent --head --fail "${binaries_url}"; then
        echoerr "${binaries_url} is reachable. Fetching main..."
        # Main fetch
        if [ -z ${ddprof_name-:""} ]; then
            ddprof_name="ddprof-main-amd64-alpine-linux-musl.tar.xz"
        fi

        if [ -z ${ddprof_path-:""} ]; then
            # TODO: This requires appgate. Should we put a warning somewhere ?
            ddprof_path=${binaries_url}"ddprof-build/"
        fi
        cmd="curl -L -o ${ddprof_name} --insecure ${ddprof_path}/${ddprof_name}"
        echoerr ${cmd}
        eval $cmd
        retVal=$?
        if [ $retVal -eq 1 ]; then
            echoerr "Error downloading from s3"
        fi
        return 0
    else
        echoerr "Unable to reach s3"
        return 1
    fi
}

download_from_github() {
    # Download the latest release candidate and store it in the current directory
    ddprof_name="ddprof-amd64-linux.tar.xz"
    url_release_candidate="https://github.com/DataDog/ddprof/releases/download/latest-rc/${ddprof_name}"

    echo "Downloading from ${url_release_candidate}..."
    curl -fsSL -O "${url_release_candidate}"
}

# Takes name and path to binary
ddprof_install_path=${1-""}
if [ -z ${ddprof_install_path} ] || [ ! -d ${ddprof_install_path-:""} ]; then
    echo "Specify install path"
    exit 1
fi
ddprof_name=${2-""}
binaries_path=${3-"/app/binaries"}
s3_path=${4-""}

mkdir -p /tmp
pushd /tmp

if look_in_folder ${binaries_path}; then
    echo "Success finding file in binaries"
elif look_in_s3 ${s3_path}; then
    echo "Success fetching from s3..."
else
    echo "Fetching latest GH release..."
    # This should not fail
    download_from_github
fi

echo ${ddprof_name}
if [[ ${ddprof_name} == ${binaries_path}/ddprof ]]; then
    cp ${ddprof_name} ${ddprof_install_path}/ddprof
else
    tar xvf ${ddprof_name} ddprof/bin/ddprof -O > ${ddprof_install_path}/ddprof
    rm -f ./${ddprof_name}
fi
chmod 755 ${ddprof_install_path}/ddprof
popd

PROFILER_VERSION=$(${ddprof_install_path}/ddprof --version)
echo "Profiler version: $(echo ${PROFILER_VERSION})"
exit 0

#!/usr/bin/env bash

set -euo pipefail

BUILD_VERSION=$(cat version)
BUILD_DATE=$(date "+%F %T")
COMMIT_SHA1=$(git rev-parse HEAD)

TARGET_DIR="dist"
TARGET_NAME="mmh"
DARWIN_ARCHES="amd64 arm64"
COMMANDS="mcp mcs mcx mec mgo mping mtun"

LDFLAGS="-X 'github.com/mritd/mmh/cmd.Version=${BUILD_VERSION}' \
    -X 'github.com/mritd/mmh/cmd.BuildDate=${BUILD_DATE}' \
    -X 'github.com/mritd/mmh/cmd.CommitID=${COMMIT_SHA1}'"

BIN_DIR="$(go env GOPATH)/bin"

build_darwin() {
    local arch=$1
    local target="${TARGET_DIR}/${TARGET_NAME}_darwin_${arch}"
    echo "build => ${target} (darwin/${arch})"
    GOOS=darwin GOARCH="${arch}" CGO_ENABLED=0 go build -trimpath -o "${target}" \
        -ldflags "${LDFLAGS} -w -s"
}

build_universal() {
    local universal="${TARGET_DIR}/${TARGET_NAME}_darwin_universal"
    if ! command -v lipo >/dev/null 2>&1; then
        echo "WARN: lipo not found, skip universal binary (install Xcode Command Line Tools)" >&2
        return
    fi
    echo "lipo  => ${universal}"
    lipo -create -output "${universal}" \
        "${TARGET_DIR}/${TARGET_NAME}_darwin_amd64" \
        "${TARGET_DIR}/${TARGET_NAME}_darwin_arm64"
}

checksums() {
    (cd "${TARGET_DIR}" && shasum -a 256 ${TARGET_NAME}_darwin_* > checksums.txt)
    echo "checksums => ${TARGET_DIR}/checksums.txt"
}

install_bin() {
    mkdir -p "${BIN_DIR}"
    echo "install => ${BIN_DIR}/mmh"
    GOOS=darwin GOARCH="$(go env GOARCH)" CGO_ENABLED=0 go build -trimpath \
        -o "${BIN_DIR}/mmh" -ldflags "${LDFLAGS}"
    for cmd in ${COMMANDS}; do
        echo "install => ${BIN_DIR}/${cmd}"
        ln -sf "${BIN_DIR}/mmh" "${BIN_DIR}/${cmd}"
    done
}

uninstall_bin() {
    echo "remove => ${BIN_DIR}/mmh"
    rm -f "${BIN_DIR}/mmh"
    for cmd in ${COMMANDS}; do
        echo "remove => ${BIN_DIR}/${cmd}"
        rm -f "${BIN_DIR}/${cmd}"
    done
}

gen_completion() {
    local mmh_bin="${BIN_DIR}/mmh"
    if [ ! -x "${mmh_bin}" ]; then
        echo "mmh binary not found at ${mmh_bin}, run 'make install' first" >&2
        exit 1
    fi
    for s in bash zsh fish powershell; do
        "${mmh_bin}" --completion "${s}" > "docs/completions/mmh.${s}"
    done
    cat docs/completions/mmh.zsh > docs/completions/mmh.ohmyzsh
    echo 'compdef _mmh mmh' >> docs/completions/mmh.ohmyzsh
    for cmd in ${COMMANDS}; do
        echo "compdef _mmh ${cmd}" >> docs/completions/mmh.ohmyzsh
    done
}

case "${1:-build}" in
install)
    install_bin
    ;;
uninstall)
    uninstall_bin
    ;;
completion)
    gen_completion
    ;;
build | *)
    rm -rf "${TARGET_DIR}"
    mkdir -p "${TARGET_DIR}"
    for arch in ${DARWIN_ARCHES}; do
        build_darwin "${arch}"
    done
    build_universal
    checksums
    ;;
esac

#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
#
# Install script for the ovp8xx CLI (https://github.com/graugans/go-ovp8xx)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/graugans/go-ovp8xx/v2/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/graugans/go-ovp8xx/v2/install.sh | sh -s -- --version v2.10.0 --bin-dir /usr/local/bin
#
# The script downloads the release archive matching the current OS and CPU
# architecture, verifies it against the checksums.txt of the release and
# installs the binary. It never calls sudo on its own.

set -eu

REPO="graugans/go-ovp8xx"
BINARY="ovp8xx"
VARIANTS="ovp9xx o3cxxx"

usage() {
    cat <<EOF
Install the ${BINARY} CLI from the GitHub releases of ${REPO}.

Options:
  -b, --bin-dir DIR   Installation directory (default: \$HOME/.local/bin)
  -v, --version TAG   Release to install, e.g. v2.10.0 (default: latest)
      --no-links      Do not create the ${VARIANTS} links
  -h, --help          Show this help

Environment variables:
  OVP8XX_INSTALL_DIR  Same as --bin-dir
  OVP8XX_VERSION      Same as --version
EOF
}

log() {
    printf '%s\n' "$*" >&2
}

fail() {
    log "error: $*"
    exit 1
}

need() {
    command -v "$1" >/dev/null 2>&1 || fail "required command '$1' not found"
}

detect_os() {
    case "$(uname -s)" in
    Linux) echo "Linux" ;;
    Darwin) echo "Darwin" ;;
    *) fail "unsupported OS '$(uname -s)', please download the binary from https://github.com/${REPO}/releases" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
    x86_64 | amd64) echo "x86_64" ;;
    aarch64 | arm64) echo "arm64" ;;
    i386 | i686) echo "i386" ;;
    *) fail "unsupported architecture '$(uname -m)', please download the binary from https://github.com/${REPO}/releases" ;;
    esac
}

download() {
    # download URL FILE
    curl -fsSL --proto '=https' --tlsv1.2 -o "$2" "$1" || fail "download of $1 failed"
}

sha256() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | cut -d ' ' -f 1
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | cut -d ' ' -f 1
    else
        fail "neither sha256sum nor shasum found, cannot verify the download"
    fi
}

main() {
    bin_dir="${OVP8XX_INSTALL_DIR:-${HOME}/.local/bin}"
    version="${OVP8XX_VERSION:-latest}"
    links=1

    while [ $# -gt 0 ]; do
        case "$1" in
        -b | --bin-dir)
            [ $# -ge 2 ] || fail "$1 requires an argument"
            bin_dir="$2"
            shift 2
            ;;
        -v | --version)
            [ $# -ge 2 ] || fail "$1 requires an argument"
            version="$2"
            shift 2
            ;;
        --no-links)
            links=0
            shift
            ;;
        -h | --help)
            usage
            exit 0
            ;;
        *) fail "unknown option '$1', see --help" ;;
        esac
    done

    need curl
    need tar
    need uname

    os="$(detect_os)"
    arch="$(detect_arch)"
    [ "${os}/${arch}" != "Darwin/i386" ] || fail "there is no 32 bit macOS build"

    archive="go-ovp8xx_${os}_${arch}.tar.gz"
    if [ "${version}" = "latest" ]; then
        base_url="https://github.com/${REPO}/releases/latest/download"
    else
        case "${version}" in
        v*) ;;
        *) version="v${version}" ;;
        esac
        base_url="https://github.com/${REPO}/releases/download/${version}"
    fi

    tmp_dir="$(mktemp -d)"
    # shellcheck disable=SC2064 # expand tmp_dir now
    trap "rm -rf '${tmp_dir}'" EXIT INT TERM

    log "Downloading ${archive} (${version})"
    download "${base_url}/${archive}" "${tmp_dir}/${archive}"
    download "${base_url}/checksums.txt" "${tmp_dir}/checksums.txt"

    log "Verifying checksum"
    expected="$(awk -v f="${archive}" '$2 == f { print $1 }' "${tmp_dir}/checksums.txt")"
    [ -n "${expected}" ] || fail "${archive} is not listed in checksums.txt"
    actual="$(sha256 "${tmp_dir}/${archive}")"
    [ "${expected}" = "${actual}" ] || fail "checksum mismatch for ${archive}: expected ${expected}, got ${actual}"

    tar -xzf "${tmp_dir}/${archive}" -C "${tmp_dir}" "${BINARY}" || fail "cannot extract ${BINARY} from ${archive}"

    mkdir -p "${bin_dir}" 2>/dev/null || true
    if [ ! -d "${bin_dir}" ] || [ ! -w "${bin_dir}" ]; then
        fail "${bin_dir} is not writable. Choose another directory with --bin-dir or run the script with sudo, e.g. '... | sudo sh -s -- --bin-dir /usr/local/bin'"
    fi

    # Install via a temporary file and rename it, so a running binary is not overwritten in place.
    cp "${tmp_dir}/${BINARY}" "${bin_dir}/.${BINARY}.new"
    chmod 0755 "${bin_dir}/.${BINARY}.new"
    mv -f "${bin_dir}/.${BINARY}.new" "${bin_dir}/${BINARY}"
    log "Installed ${bin_dir}/${BINARY}"

    if [ "${links}" -eq 1 ]; then
        for variant in ${VARIANTS}; do
            ln -sf "${BINARY}" "${bin_dir}/${variant}"
            log "Linked    ${bin_dir}/${variant} -> ${BINARY}"
        done
    fi

    case ":${PATH}:" in
    *":${bin_dir}:"*) ;;
    *) log "Note: ${bin_dir} is not in your PATH. Add it e.g. with: export PATH=\"${bin_dir}:\$PATH\"" ;;
    esac

    "${bin_dir}/${BINARY}" --version >&2 || true
}

# Everything is wrapped in main so a partially downloaded script does nothing.
main "$@"

#!/bin/bash
set -euo pipefail

GOLANGCI_LINT_VERSION="v2.15.0"

# The base image may ship a yarn apt source with an expired signing key,
# which breaks `apt-get update`. Yarn is not needed for this project.
sudo rm -f /etc/apt/sources.list.d/yarn.list

# The volume holding the GitHub CLI configuration is created owned by root
sudo chown "$(id -un):$(id -gn)" "${HOME}/.config" "${HOME}/.config/gh"

# install missing packages
sudo apt-get update
sudo apt-get install -q -y \
    git-lfs \
    pre-commit \
    shfmt \
    shellcheck
sudo apt-get clean
sudo rm -rf /var/lib/apt/lists/*

# Fetch git large files (the PCIC test data is stored in LFS)
git lfs install --local
git lfs pull

# Go CI Lint
curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/${GOLANGCI_LINT_VERSION}/install.sh" |
    sh -s -- -b "$(go env GOPATH)/bin" "${GOLANGCI_LINT_VERSION}"

# Go tools
go install golang.org/x/tools/cmd/goimports@latest

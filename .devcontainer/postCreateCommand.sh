#!/bin/bash

# install missing packages
sudo apt-get update &&
    sudo apt-get install git-lfs \
        python3-pip \
        python3-venv \
        shfmt \
        shellcheck &&
    sudo apt-get clean &&
    sudo rm -rf /var/lib/apt/lists/*

# Fetch git large files
git lfs fetch --all

# Install Python packages
export PATH=~/.local/bin:$PATH
pip3 install --break-system-packages --upgrade pip
pip3 install --break-system-packages -r requirements.txt

# Go CI Lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.54.0

# Go tools
go install golang.org/x/tools/cmd/goimports@latest

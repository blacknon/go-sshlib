#!/usr/bin/env bash
set -euo pipefail

mkdir -p "${GOCACHE:-/tmp/go-sshlib-gocache}"
mkdir -p "${GOTMPDIR:-/tmp/go-sshlib-gotmp}"

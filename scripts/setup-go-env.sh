#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
gocache_posix="${repo_root}/.tmp/go-build"
gotmpdir_posix="${repo_root}/.tmp/go-tmp"

mkdir -p "${gocache_posix}" "${gotmpdir_posix}"

export GOCACHE="${gocache_posix}"
export GOTMPDIR="${gotmpdir_posix}"

case "$(uname -s)" in
MINGW* | MSYS* | CYGWIN*)
	if command -v cygpath >/dev/null 2>&1; then
		export GOCACHE="$(cygpath -w "${gocache_posix}")"
		export GOTMPDIR="$(cygpath -w "${gotmpdir_posix}")"
	fi
	;;
esac

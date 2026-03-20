#!/usr/bin/env bash
set -euo pipefail

mise run test-unit
mise run test-integration

#!/usr/bin/env bash
set -euo pipefail

docker info >/dev/null
docker compose -f docker-compose.test.yml up -d --build sshd

cleanup() {
  docker compose -f docker-compose.test.yml down -v
}
trap cleanup EXIT

container_id="$(docker compose -f docker-compose.test.yml ps -q sshd)"
ready=0
for _ in $(seq 1 30); do
  if [ -n "$container_id" ] && docker compose -f docker-compose.test.yml exec -T sshd pgrep -x sshd >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
[ "$ready" -eq 1 ]

SSHLIB_INTEGRATION=1 \
SSHLIB_TEST_HOST=127.0.0.1 \
SSHLIB_TEST_PORT=2222 \
SSHLIB_TEST_USER=testuser \
SSHLIB_TEST_PASSWORD=testpass \
go test -v ./...

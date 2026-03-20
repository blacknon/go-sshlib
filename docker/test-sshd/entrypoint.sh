#!/bin/sh
set -eu

mkdir -p /var/run/sshd
ssh-keygen -A

exec /usr/sbin/sshd -D -e -f /etc/ssh/sshd_config

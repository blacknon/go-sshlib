FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends openssh-server procps \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /var/run/sshd \
    && useradd -m -s /bin/bash testuser \
    && echo 'testuser:testpass' | chpasswd

COPY docker/test-sshd/sshd_config /etc/ssh/sshd_config
COPY docker/test-sshd/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 2222

CMD ["/usr/local/bin/entrypoint.sh"]

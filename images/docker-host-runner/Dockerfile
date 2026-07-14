FROM ghcr.io/actions/actions-runner:2.332.0@sha256:8c3f5970b8ceb90cbd3e89b80c6806bb74d9c31686e9177c743323a4539d12f5

USER root

ARG TARGETARCH
ARG DOCKER_COMPOSE_VERSION=2.40.3
ARG DOCKER_COMPOSE_AMD64_SHA256=dba9d98e1ba5bfe11d88c99b9bd32fc4a0624a30fafe68eea34d61a3e42fd372

RUN set -eux; \
    test "$TARGETARCH" = "amd64"; \
    asset="docker-compose-linux-x86_64"; \
    curl -fsSL \
      "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/${asset}" \
      -o "/tmp/${asset}"; \
    echo "${DOCKER_COMPOSE_AMD64_SHA256}  /tmp/${asset}" | sha256sum -c -; \
    install -m 0755 "/tmp/${asset}" /usr/local/lib/docker/cli-plugins/docker-compose; \
    rm -f "/tmp/${asset}"; \
    rm -f \
      /usr/bin/dockerd \
      /usr/bin/docker-proxy \
      /usr/bin/containerd \
      /usr/bin/containerd-shim \
      /usr/bin/containerd-shim-runc-v2 \
      /usr/bin/ctr \
      /usr/bin/runc

USER runner
WORKDIR /home/runner

ENV RUNNER_DISABLEUPDATE=1 \
    RUNNER_TOOL_CACHE=/opt/hostedtoolcache

CMD ["/home/runner/run.sh"]

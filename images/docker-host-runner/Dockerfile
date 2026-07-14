FROM ghcr.io/actions/actions-runner:2.332.0@sha256:8c3f5970b8ceb90cbd3e89b80c6806bb74d9c31686e9177c743323a4539d12f5

USER root

ARG TARGETARCH
ARG DOCKER_COMPOSE_VERSION=2.40.3
ARG DOCKER_COMPOSE_AMD64_SHA256=dba9d98e1ba5bfe11d88c99b9bd32fc4a0624a30fafe68eea34d61a3e42fd372
ARG DOCKER_COMPOSE_ARM64_SHA256=d26373b19e89160546d15407516cc59f453030d9bc5b43ba7faf16f7b4980137
ARG GH_VERSION=2.96.0
ARG GH_AMD64_SHA256=83d5c2ccad5498f58bf6368acb1ab32588cf43ab3a4b1c301bf36328b1c8bd60
ARG GH_ARM64_SHA256=06f86ec7103d41993b76cd78072f43595c34aaa56506d971d9860e67140bf909
ARG RUSTUP_VERSION=1.29.0
ARG RUSTUP_AMD64_SHA256=4acc9acc76d5079515b46346a485974457b5a79893cfb01112423c89aeb5aa10
ARG RUSTUP_ARM64_SHA256=9732d6c5e2a098d3521fca8145d826ae0aaa067ef2385ead08e6feac88fa5792

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      build-essential \
      cmake \
      pkg-config \
    && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    case "$TARGETARCH" in \
      amd64) \
        compose_asset="docker-compose-linux-x86_64"; \
        compose_checksum="$DOCKER_COMPOSE_AMD64_SHA256"; \
        gh_asset="gh_${GH_VERSION}_linux_amd64"; \
        gh_checksum="$GH_AMD64_SHA256"; \
        rustup_target="x86_64-unknown-linux-gnu"; \
        rustup_checksum="$RUSTUP_AMD64_SHA256"; \
        ;; \
      arm64) \
        compose_asset="docker-compose-linux-aarch64"; \
        compose_checksum="$DOCKER_COMPOSE_ARM64_SHA256"; \
        gh_asset="gh_${GH_VERSION}_linux_arm64"; \
        gh_checksum="$GH_ARM64_SHA256"; \
        rustup_target="aarch64-unknown-linux-gnu"; \
        rustup_checksum="$RUSTUP_ARM64_SHA256"; \
        ;; \
      *) \
        echo "unsupported Docker host runner architecture: $TARGETARCH" >&2; \
        exit 1; \
        ;; \
    esac; \
    curl -fsSL \
      "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/${compose_asset}" \
      -o "/tmp/${compose_asset}"; \
    echo "${compose_checksum}  /tmp/${compose_asset}" | sha256sum -c -; \
    install -m 0755 "/tmp/${compose_asset}" /usr/local/lib/docker/cli-plugins/docker-compose; \
    curl -fsSL \
      "https://github.com/cli/cli/releases/download/v${GH_VERSION}/${gh_asset}.tar.gz" \
      -o "/tmp/${gh_asset}.tar.gz"; \
    echo "${gh_checksum}  /tmp/${gh_asset}.tar.gz" | sha256sum -c -; \
    tar -xzf "/tmp/${gh_asset}.tar.gz" -C /tmp; \
    install -m 0755 "/tmp/${gh_asset}/bin/gh" /usr/local/bin/gh; \
    curl -fsSL \
      "https://static.rust-lang.org/rustup/archive/${RUSTUP_VERSION}/${rustup_target}/rustup-init" \
      -o /tmp/rustup-init; \
    echo "${rustup_checksum}  /tmp/rustup-init" | sha256sum -c -; \
    install -m 0755 /tmp/rustup-init /usr/local/bin/rustup-init; \
    rm -rf "/tmp/${compose_asset}" "/tmp/${gh_asset}" "/tmp/${gh_asset}.tar.gz" /tmp/rustup-init; \
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

ENV CARGO_HOME=/home/runner/.cargo \
    RUSTUP_HOME=/home/runner/.rustup \
    PATH=/home/runner/.cargo/bin:${PATH} \
    RUNNER_DISABLEUPDATE=1 \
    RUNNER_TOOL_CACHE=/opt/hostedtoolcache

CMD ["/home/runner/run.sh"]

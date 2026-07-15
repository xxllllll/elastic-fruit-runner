---
title: Docker host runners with OrbStack
description: Run one ephemeral Linux AMD64 or ARM64 JIT runner through the active OrbStack Docker context.
---

The `docker-host` backend is the first-phase Mac mini MVP. It creates no idle
runner containers, does not use privileged DinD, and uses the active Docker
Context Unix socket.

## Requirements

- Apple Silicon Mac with OrbStack running
- Active Docker Context with a `unix://` endpoint
- Docker support for `linux/amd64` or `linux/arm64`
- One private repository runner set with `max_runners: 1`

Verify the active context without copying its path into configuration:

```sh
docker context show
docker context inspect --format '{{.Endpoints.docker.Host}}'
docker info --format '{{.Architecture}}'
```

`DOCKER_HOST=unix:///absolute/path/to/docker.sock` is also supported. TCP and
SSH endpoints are rejected in this phase.

## Resolve the published runner image

The Fork publishes a public multi-platform OCI Index. Resolve its version tag
to an immutable digest before configuring a production runner:

```sh
images/docker-host-runner/resolve-release.sh 2.332.0-r1
```

The command requires Docker Buildx and jq. It accepts no implicit `latest`
version and verifies that the index contains exactly `linux/amd64` and
`linux/arm64`. Use the printed `image=...@sha256:...` value in configuration.

Verify both remote architectures with:

```sh
RUNNER_IMAGE="ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner"
EFR_IMAGE_TEST_AMD64_REFERENCE="${RUNNER_IMAGE}@sha256:<amd64-digest>" \
EFR_IMAGE_TEST_ARM64_REFERENCE="${RUNNER_IMAGE}@sha256:<arm64-digest>" \
  images/docker-host-runner/test.sh
```

## Build the image locally

Use local architecture tags while changing the Dockerfile:

```sh
docker build \
  --platform linux/amd64 \
  -t elastic-fruit-runner/docker-host-runner:2.332.0-amd64 \
  images/docker-host-runner

docker build \
  --platform linux/arm64 \
  -t elastic-fruit-runner/docker-host-runner:2.332.0-arm64 \
  images/docker-host-runner
```

The image pins GitHub Actions Runner `2.332.0`, GitHub CLI `2.96.0`, Rustup
initializer `1.29.0`, and Docker Compose `2.40.3`. It includes GCC, G++, Make,
CMake, pkg-config, Docker CLI, and Buildx. It removes Docker daemon executables
and contains no repository credentials or project-specific Rust toolchain.

Verify both locally built architectures with:

```sh
images/docker-host-runner/test.sh
```

## Configure one repository

```yaml
api_addr: 127.0.0.1:8080
cache_root: /path/to/elastic-fruit-runner/cache

repos:
  - repo: your-org/your-private-repo
    auth:
      pat_token: github_pat_xxx
    runner_sets:
      - name: repo-orbstack-amd64
        backend: docker-host
        image: ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner@sha256:<verified-index-digest>
        labels: [self-hosted, linux, amd64]
        max_runners: 1
        platform: linux/amd64
        cache_namespace: your-private-repo

idle_timeout: 15m
```

For a native Apple Silicon runner, keep the same multi-platform index digest,
set `platform: linux/arm64`, replace the `amd64` label with `arm64`, and use a
distinct Runner Set name such as `repo-orbstack-arm64`. Only one `docker-host`
Runner Set can be configured at a time.

When publishing an update, use a new `<runner-version>-r<revision>` tag, resolve
its new digest, preserve the current configuration as a rollback point, and
replace only the `image` value. Restart the Controller and complete a real JIT
smoke before removing the backup. Never point a production Runner Set at
`latest` or at a version tag without its digest.

Keep the real PAT or GitHub App private key outside the repository. Do not put
JIT configuration, runner tokens, or credentials under `cache_root`.

## Cache layout

The backend creates these host directories:

```text
cache_root/
  shared/pnpm-store/
  shared/tool-cache/
  <cache_namespace>/linux-amd64/cargo-home/
  <cache_namespace>/linux-amd64/rustup-home/
  <cache_namespace>/linux-amd64/cargo-target/
  <cache_namespace>/linux-amd64/sccache/
  <cache_namespace>/linux-arm64/cargo-home/
  <cache_namespace>/linux-arm64/rustup-home/
  <cache_namespace>/linux-arm64/cargo-target/
  <cache_namespace>/linux-arm64/sccache/
```

Cargo Home, Rustup Home, Cargo Target, and sccache are project- and
Platform-scoped. This prevents Cargo configuration, Rustup proxy executables,
and installed toolchains from crossing repository or CPU architecture
boundaries. The old `shared/cargo-home` directory is not copied or deleted, so
the first job after this change has a cold Cargo and Rustup cache. The pnpm
store and Runner Tool Cache remain shared. The runner work directory remains
ephemeral. BuildKit cache remains in the host Docker Engine and is never
deleted by runner cleanup.

The image places `/home/runner/.cargo/bin` on `PATH`. Initialize an empty Rust
cache before the first `rustup` command:

```sh
if ! command -v rustup >/dev/null 2>&1; then
  rustup-init -y --profile minimal --default-toolchain none --no-modify-path
fi
rustup show active-toolchain
```

The repository's `rust-toolchain.toml` remains responsible for selecting the
actual Rust version.

## Validate with a JIT smoke job

Use a private test repository and a fine-grained PAT restricted to that
repository. Repository-level JIT configuration requires `Administration: Read
and write`; other repository permissions can remain disabled. Keep the token in
the local EFR configuration only.

Create a temporary workflow that targets the runner set name exactly:

```yaml
name: EFR JIT Smoke

on: workflow_dispatch

permissions:
  contents: read

jobs:
  smoke:
    runs-on: repo-orbstack-amd64
    timeout-minutes: 10
    steps:
      - name: Verify host-socket runner
        shell: bash
        run: |
          set -euo pipefail
          test "$(uname -m)" = "x86_64"
          test -S /var/run/docker.sock
          ! command -v dockerd
          docker version
          docker buildx version
          docker compose version
          docker run --rm --platform linux/amd64 alpine:3.22 uname -m
```

A successful smoke must also leave zero EFR runner containers after the job,
preserve the configured cache directories, and leave unrelated containers and
named volumes unchanged.

## Cleanup behavior

Containers receive explicit EFR ownership, runner-set, and runner-name labels.
Startup cleanup only matches these labels. Job cleanup uses `docker rm -f -v`,
which removes the runner container and anonymous volumes but does not delete
named volumes, host cache directories, or BuildKit state.

## Current limits

- One repository-level runner set
- One concurrent runner
- Linux AMD64 or ARM64, one Platform at a time
- Production service installation and automatic startup are not included
- No remote wake after the Mac enters full sleep

Later phases can add multiple repositories, concurrency, and production service
deployment after the single-runner path is validated.

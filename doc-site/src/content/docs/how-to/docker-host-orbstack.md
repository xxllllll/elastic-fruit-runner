---
title: Docker host runners with OrbStack
description: Run one ephemeral Linux AMD64 JIT runner through the active OrbStack Docker context.
---

The `docker-host` backend is the first-phase Mac mini MVP. It creates no idle
runner containers, does not use privileged DinD, and uses the active Docker
Context Unix socket.

## Requirements

- Apple Silicon Mac with OrbStack running
- Active Docker Context with a `unix://` endpoint
- Docker support for `linux/amd64`
- One private repository runner set with `max_runners: 1`

Verify the active context without copying its path into configuration:

```sh
docker context show
docker context inspect --format '{{.Endpoints.docker.Host}}'
docker info --format '{{.Architecture}}'
```

`DOCKER_HOST=unix:///absolute/path/to/docker.sock` is also supported. TCP and
SSH endpoints are rejected in this phase.

## Build the runner image

```sh
docker build \
  --platform linux/amd64 \
  -t elastic-fruit-runner/docker-host-runner:2.332.0 \
  images/docker-host-runner
```

The image pins GitHub Actions Runner `2.332.0` and Docker Compose `2.40.3`.
It includes Docker CLI and Buildx, removes Docker daemon executables, and
contains no repository credentials.

## Configure one repository

```yaml
api_addr: 127.0.0.1:8080
cache_root: /path/to/elastic-fruit-runner/cache

repos:
  - repo: your-org/your-private-repo
    auth:
      pat_token: ghp_xxx
    runner_sets:
      - name: repo-orbstack-amd64
        backend: docker-host
        image: elastic-fruit-runner/docker-host-runner:2.332.0
        labels: [self-hosted, linux, amd64]
        max_runners: 1
        platform: linux/amd64
        cache_namespace: your-private-repo

idle_timeout: 15m
```

Keep the real PAT or GitHub App private key outside the repository. Do not put
JIT configuration, runner tokens, or credentials under `cache_root`.

## Cache layout

The backend creates these host directories:

```text
cache_root/
  shared/cargo-home/
  shared/pnpm-store/
  shared/tool-cache/
  <cache_namespace>/cargo-target/
  <cache_namespace>/sccache/
```

Cargo Target and sccache are project-scoped. Cargo Home, pnpm store, and the
Runner Tool Cache are shared. The runner work directory remains ephemeral.
BuildKit cache remains in the host Docker Engine and is never deleted by runner
cleanup.

## Cleanup behavior

Containers receive explicit EFR ownership, runner-set, and runner-name labels.
Startup cleanup only matches these labels. Job cleanup uses `docker rm -f -v`,
which removes the runner container and anonymous volumes but does not delete
named volumes, host cache directories, or BuildKit state.

## Current limits

- One repository-level runner set
- One concurrent runner
- Linux AMD64 only
- No real workflow smoke test is performed until credentials are supplied
- No remote wake after the Mac enters full sleep

Later phases can add ARM64, multiple repositories, concurrency, and production
service deployment after the single-runner path is validated.

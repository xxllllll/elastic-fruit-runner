# Docker Host Runner Image

This image is the runner environment for the `docker-host` backend.

- Base: multi-architecture `ghcr.io/actions/actions-runner:2.332.0`, pinned by
  OCI Index digest.
- Includes the GitHub Actions runner, Git, curl, jq, Python 3, CA certificates,
  Docker CLI, Buildx, and Docker Compose `2.40.3`.
- Removes Docker daemon and container runtime executables.
- Runs `/home/runner/run.sh` as the non-root `runner` user.
- Disables in-place runner updates so the image version remains reproducible.
- Contains no GitHub or repository credentials.

Build architecture-specific local images:

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

The Dockerfile selects the matching Docker Compose asset and verifies its
architecture-specific SHA-256. Unsupported `TARGETARCH` values fail the build.

The backend mounts the active Docker Context Unix socket at runtime. Do not add
credentials or host-specific socket paths to this image.

# Docker Host Runner Image

This image is the runner environment for the `docker-host` backend.

- Base: `ghcr.io/actions/actions-runner:2.332.0`, pinned by digest.
- Includes the GitHub Actions runner, Git, curl, jq, Python 3, CA certificates,
  Docker CLI, Buildx, and Docker Compose `2.40.3`.
- Removes Docker daemon and container runtime executables.
- Runs `/home/runner/run.sh` as the non-root `runner` user.
- Disables in-place runner updates so the image version remains reproducible.
- Contains no GitHub or repository credentials.

Build the first-phase AMD64 image:

```sh
docker build \
  --platform linux/amd64 \
  -t elastic-fruit-runner/docker-host-runner:2.332.0 \
  images/docker-host-runner
```

The backend mounts the active Docker Context Unix socket at runtime. Do not add
credentials or host-specific socket paths to this image.

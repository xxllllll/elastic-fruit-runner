# Docker Host Runner Image

This image is the runner environment for the `docker-host` backend.

- Base: multi-architecture `ghcr.io/actions/actions-runner:2.332.0`, pinned by
  OCI Index digest.
- Includes the GitHub Actions runner, GitHub CLI `2.96.0`, Git, curl, jq,
  Python 3, CA certificates, GCC, G++, Make, CMake, pkg-config, Docker CLI,
  Buildx, and Docker Compose `2.40.3`.
- Includes the pinned Rustup `1.29.0` initializer, but no project-specific Rust
  toolchain. Cargo Home is already on `PATH`.
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

The Dockerfile selects matching Docker Compose, GitHub CLI, and Rustup assets
and verifies their architecture-specific SHA-256 values. Unsupported
`TARGETARCH` values fail the build.

Run the complete AMD64 and ARM64 image contract test:

```sh
images/docker-host-runner/test.sh
```

The test builds both architectures, verifies the installed tools and removed
daemon executables, then proves that a Rustup initialization can be reused by a
second container through persistent Cargo and Rustup bind mounts.

At runtime the backend mounts project- and Platform-scoped Cargo and Rustup
homes. A workflow can initialize an empty cache without pinning a toolchain in
the image:

```sh
if ! command -v rustup >/dev/null 2>&1; then
  rustup-init -y --profile minimal --default-toolchain none --no-modify-path
fi
rustup show active-toolchain
```

The backend mounts the active Docker Context Unix socket at runtime. Do not add
credentials or host-specific socket paths to this image.

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

## Published image

The Fork publishes one multi-platform OCI Index at:

```text
ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner
```

Release tags use the Actions Runner version and an image recipe revision, for
example `2.332.0-r1`. A source tag named `sha-<git-commit>` points to the same
index. The workflow refuses to replace either existing tag, and `latest` is
intentionally not published.

Resolve and validate a release before using it:

```sh
images/docker-host-runner/resolve-release.sh 2.332.0-r1
```

The command verifies that the index contains exactly `linux/amd64` and
`linux/arm64`, then prints the immutable index reference and both platform
digests:

```text
image=ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner@sha256:<index-digest>
amd64_digest=sha256:<amd64-manifest-digest>
arm64_digest=sha256:<arm64-manifest-digest>
```

Production configuration must use the printed index digest and an explicit
platform. Tags are only discovery inputs for an intentional update.

## Release a new image

1. Update and test `Dockerfile` and this directory's scripts in a pull request.
2. Merge the pull request into the Fork's default branch.
3. Run the `Docker Host Runner Image` workflow from the default branch with a
   positive image revision.
4. Check that the version and source tags resolve to the same index digest.
5. Run `resolve-release.sh`, then test the returned digest with
   `EFR_IMAGE_TEST_REFERENCE`.
6. Back up the host configuration, replace only its image digest, and restart
   the Controller.
7. Run a real JIT smoke before treating the new digest as active.

To revert an update, restore the previous digest or the configuration backup.
Do not delete the old Registry version until every host has moved away from it.

The publish job uses the repository `GITHUB_TOKEN` with `packages: write`.
The public image contains no repository credentials, so hosts do not need a
long-lived Package token to pull it.

## Local development

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

Run the complete AMD64 and ARM64 local image contract test:

```sh
images/docker-host-runner/test.sh
```

The test builds both architectures, verifies the installed tools and removed
daemon executables, then proves that a Rustup initialization can be reused by a
second container through persistent Cargo and Rustup bind mounts.

The same contract can verify a published index digest without rebuilding it:

```sh
RUNNER_IMAGE="ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner"
EFR_IMAGE_TEST_REFERENCE="${RUNNER_IMAGE}@sha256:<index-digest>" \
  images/docker-host-runner/test.sh
```

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

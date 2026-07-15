#!/usr/bin/env bash
set -euo pipefail

image_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
resource_suffix="${EFR_IMAGE_TEST_SUFFIX:-$(date +%s)-$$}"
compose_version="$(sed -n 's/^ARG DOCKER_COMPOSE_VERSION=//p' "$image_dir/Dockerfile")"
gh_version="$(sed -n 's/^ARG GH_VERSION=//p' "$image_dir/Dockerfile")"
rustup_version="$(sed -n 's/^ARG RUSTUP_VERSION=//p' "$image_dir/Dockerfile")"
remote_amd64_image="${EFR_IMAGE_TEST_AMD64_REFERENCE:-}"
remote_arm64_image="${EFR_IMAGE_TEST_ARM64_REFERENCE:-}"
declare -a test_images=()
declare -a test_directories=()

[[ -n "$compose_version" && -n "$gh_version" && -n "$rustup_version" ]]
[[ ( -z "$remote_amd64_image" && -z "$remote_arm64_image" ) ||
  ( -n "$remote_amd64_image" && -n "$remote_arm64_image" ) ]] || {
  echo "both EFR_IMAGE_TEST_AMD64_REFERENCE and EFR_IMAGE_TEST_ARM64_REFERENCE are required" >&2
  exit 1
}

cleanup() {
  local resource
  set +eu
  for resource in "${test_directories[@]}"; do
    rm -rf "$resource"
  done
  for resource in "${test_images[@]}"; do
    docker image rm -f "$resource" >/dev/null 2>&1
  done
}
trap cleanup EXIT

verify_image_contract() {
  local image="$1" platform="$2" expected_arch="$3"
  local image_user image_volumes

  image_user="$(docker image inspect --format '{{.Config.User}}' "$image")"
  image_volumes="$(docker image inspect --format '{{json .Config.Volumes}}' "$image")"
  [[ "$image_user" == "runner" ]]
  [[ "$image_volumes" == "null" ]]

  docker run --rm --platform "$platform" \
    --env "EXPECTED_ARCH=$expected_arch" \
    --env "EXPECTED_COMPOSE_VERSION=$compose_version" \
    --env "EXPECTED_GH_VERSION=$gh_version" \
    --env "EXPECTED_RUSTUP_VERSION=$rustup_version" \
    --entrypoint bash "$image" -lc '
      set -euo pipefail
      [[ "$(uname -m)" == "$EXPECTED_ARCH" ]]
      [[ "$(id -un)" == runner ]]
      [[ ":$PATH:" == *:/home/runner/.cargo/bin:* ]]
      for command_name in gh make gcc g++ pkg-config cmake rustup-init docker; do
        command -v "$command_name" >/dev/null
      done
      gh --version | grep -q "gh version $EXPECTED_GH_VERSION"
      rustup-init --version | grep -q "rustup-init $EXPECTED_RUSTUP_VERSION"
      docker buildx version >/dev/null
      [[ "$(docker compose version --short)" == "$EXPECTED_COMPOSE_VERSION" ]]
      ! command -v cargo
      ! command -v rustup
      ! command -v dockerd
      ! command -v containerd
      ! command -v runc
    '
}

verify_rustup_persistence() {
  local image="$1" platform="$2" arch_name="$3"
  local cache_root cargo_home rustup_home

  cache_root="$(mktemp -d "${TMPDIR:-/tmp}/efr-rust-${arch_name}.XXXXXX")"
  cargo_home="$cache_root/cargo-home"
  rustup_home="$cache_root/rustup-home"
  mkdir -p "$cargo_home" "$rustup_home"
  chmod 0777 "$cargo_home" "$rustup_home"
  test_directories+=("$cache_root")

  docker run --rm --platform "$platform" \
    --mount "type=bind,src=${cargo_home},dst=/home/runner/.cargo" \
    --mount "type=bind,src=${rustup_home},dst=/home/runner/.rustup" \
    --entrypoint bash "$image" -lc '
      set -euo pipefail
      rustup-init -y --profile minimal --default-toolchain none --no-modify-path
      rustup --version
      [[ -x "$CARGO_HOME/bin/rustup" ]]
    '

  docker run --rm --platform "$platform" \
    --mount "type=bind,src=${cargo_home},dst=/home/runner/.cargo" \
    --mount "type=bind,src=${rustup_home},dst=/home/runner/.rustup" \
    --entrypoint bash "$image" -lc '
      set -euo pipefail
      rustup --version
      [[ -x "$CARGO_HOME/bin/rustup" ]]
    '
}

test_platform() {
  local platform="$1" arch_name="$2" expected_arch="$3"
  local image

  if [[ -n "$remote_amd64_image" ]]; then
    if [[ "$arch_name" == "amd64" ]]; then
      image="$remote_amd64_image"
    else
      image="$remote_arm64_image"
    fi
    docker pull --platform "$platform" "$image"
  else
    image="elastic-fruit-runner/docker-host-runner:test-${arch_name}-${resource_suffix}"
    test_images+=("$image")
    docker build --platform "$platform" -t "$image" "$image_dir"
  fi
  verify_image_contract "$image" "$platform" "$expected_arch"
  verify_rustup_persistence "$image" "$platform" "$arch_name"
  printf 'Verified docker-host runner image for %s\n' "$platform"
}

test_platform linux/amd64 amd64 x86_64
test_platform linux/arm64 arm64 aarch64

#!/usr/bin/env bash
set -euo pipefail

readonly default_repository="ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner"
readonly docker_bin="${DOCKER_BIN:-docker}"
readonly repository="${RUNNER_IMAGE_REPOSITORY:-$default_repository}"

fail() {
  printf 'docker-host runner release resolution failed: %s\n' "$*" >&2
  exit 1
}

validate_input() {
  local tag="$1"

  [[ "$tag" =~ ^[0-9]+\.[0-9]+\.[0-9]+-r[1-9][0-9]*$ ]] ||
    fail "release tag must match <runner-version>-r<positive-revision>: value=$tag"
  [[ "$repository" != *[[:space:]@]* ]] ||
    fail "repository must not contain whitespace or a digest: value=$repository"
  command -v "$docker_bin" >/dev/null 2>&1 ||
    fail "Docker CLI is unavailable: command=$docker_bin"
  command -v jq >/dev/null 2>&1 || fail "jq is unavailable"
}

inspect_manifest() {
  local reference="$1"
  local manifest

  if ! manifest="$($docker_bin buildx imagetools inspect "$reference" \
    --format '{{json .Manifest}}')"; then
    fail "cannot inspect remote image: reference=$reference"
  fi
  jq -e . >/dev/null 2>&1 <<<"$manifest" ||
    fail "registry returned invalid manifest JSON: reference=$reference"
  printf '%s\n' "$manifest"
}

validate_manifest() {
  local reference="$1" manifest="$2"
  local media_type platforms

  media_type="$(jq -r '.mediaType // empty' <<<"$manifest")"
  [[ "$media_type" == "application/vnd.oci.image.index.v1+json" ]] ||
    fail "expected OCI image index: reference=$reference mediaType=$media_type"
  platforms="$(jq -r \
    '[.manifests[] | "\(.platform.os // "missing")/\(.platform.architecture // "missing")"] | sort | join(",")' \
    <<<"$manifest")"
  [[ "$(jq '.manifests | length' <<<"$manifest")" == "2" &&
    "$platforms" == "linux/amd64,linux/arm64" ]] ||
    fail "expected exactly linux/amd64 and linux/arm64: reference=$reference platforms=$platforms"
}

print_release() {
  local manifest="$1"
  local index_digest amd64_digest arm64_digest

  index_digest="$(jq -r '.digest // empty' <<<"$manifest")"
  amd64_digest="$(jq -r '.manifests[] | select(.platform.os == "linux" and .platform.architecture == "amd64") | .digest' <<<"$manifest")"
  arm64_digest="$(jq -r '.manifests[] | select(.platform.os == "linux" and .platform.architecture == "arm64") | .digest' <<<"$manifest")"
  for digest in "$index_digest" "$amd64_digest" "$arm64_digest"; do
    [[ "$digest" =~ ^sha256:[0-9a-f]{64}$ ]] ||
      fail "manifest contains an invalid digest: value=$digest"
  done
  printf 'image=%s@%s\n' "$repository" "$index_digest"
  printf 'amd64_digest=%s\n' "$amd64_digest"
  printf 'arm64_digest=%s\n' "$arm64_digest"
}

main() {
  [[ "$#" == "1" ]] || fail "usage: $0 <runner-version-rN>"
  local tag="$1" reference manifest

  validate_input "$tag"
  reference="$repository:$tag"
  manifest="$(inspect_manifest "$reference")"
  validate_manifest "$reference" "$manifest"
  print_release "$manifest"
}

main "$@"

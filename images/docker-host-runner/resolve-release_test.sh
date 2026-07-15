#!/usr/bin/env bash
set -euo pipefail

image_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
resolver="$image_dir/resolve-release.sh"
temp_dir="$(mktemp -d "${TMPDIR:-/tmp}/efr-release-test.XXXXXX")"
trap 'rm -rf "$temp_dir"' EXIT

fake_docker="$temp_dir/docker"
cat >"$fake_docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
[[ "$*" == *"buildx imagetools inspect"* ]]
cat "$FAKE_MANIFEST"
EOF
chmod 0755 "$fake_docker"

write_manifest() {
  local platforms="$1"
  cat >"$temp_dir/manifest.json" <<EOF
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "digest": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "manifests": [
$platforms
  ]
}
EOF
}

run_resolver() {
  FAKE_MANIFEST="$temp_dir/manifest.json" \
    DOCKER_BIN="$fake_docker" \
    RUNNER_IMAGE_REPOSITORY="ghcr.io/example/project/docker-host-runner" \
    "$resolver" "$@"
}

expect_failure() {
  local expected="$1"
  shift
  if output="$(run_resolver "$@" 2>&1)"; then
    echo "expected resolver failure for: $*" >&2
    exit 1
  fi
  grep -Fq "$expected" <<<"$output"
}

amd64='    {
      "digest": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      "platform": {"os": "linux", "architecture": "amd64"}
    }'
arm64='    {
      "digest": "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
      "platform": {"os": "linux", "architecture": "arm64"}
    }'
unknown='    {
      "digest": "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
      "platform": {"os": "unknown", "architecture": "unknown"}
    }'

write_manifest "$amd64,$(printf '\n')$arm64"
output="$(run_resolver 2.332.0-r1)"
grep -Fxq \
  'image=ghcr.io/example/project/docker-host-runner@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' \
  <<<"$output"
grep -Fxq \
  'amd64_image=ghcr.io/example/project/docker-host-runner@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb' \
  <<<"$output"
grep -Fxq \
  'arm64_image=ghcr.io/example/project/docker-host-runner@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc' \
  <<<"$output"
grep -Fxq \
  'amd64_digest=sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb' \
  <<<"$output"
grep -Fxq \
  'arm64_digest=sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc' \
  <<<"$output"

expect_failure 'release tag must match' latest

write_manifest "$amd64"
expect_failure 'exactly linux/amd64 and linux/arm64' 2.332.0-r1

write_manifest "$amd64,$(printf '\n')$arm64,$(printf '\n')$unknown"
expect_failure 'exactly linux/amd64 and linux/arm64' 2.332.0-r1

printf '%s\n' 'resolve-release tests passed'

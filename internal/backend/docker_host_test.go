package backend

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestDefaultDockerHostRunnerImageUsesPublishedDigest(t *testing.T) {
	t.Parallel()
	const want = "ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner@sha256:24d7af1adc02c8c5d21306752d3d31df1d693eeea0c9c59be4c3f481dc9911a8"
	if defaultDockerHostRunnerImage != want {
		t.Fatalf("defaultDockerHostRunnerImage = %q, want %q", defaultDockerHostRunnerImage, want)
	}
}

func TestDockerHostRunArguments(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: successfulRunResults()}
	backend, createdDirs := newTestDockerHostBackend(fake, "linux/amd64")
	const jit = "secret-jit-config"
	if err := backend.Run(context.Background(), "repo-amd64-abc12", jit); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	call := fake.calls[2]
	assertContainsSequence(t, call.args, "--context", "orbstack", "run", "-d")
	assertContainsSequence(t, call.args, "--platform", "linux/amd64")
	assertContainsSequence(t, call.args, "--group-add", "20", "--group-add", "0")
	assertContainsSequence(t, call.args, "--label", managedLabel+"=true")
	assertContainsSequence(t, call.args, "--label", runnerSetLabel+"=repo-amd64")
	assertContainsSequence(t, call.args, "--label", runnerNameLabel+"=repo-amd64-abc12")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/tmp/orbstack/docker.sock,dst=/var/run/docker.sock")
	assertContainsSequence(t, call.args, "--env", "CARGO_HOME=/home/runner/.cargo")
	assertContainsSequence(t, call.args, "--env", "RUSTUP_HOME=/home/runner/.rustup")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-amd64/cargo-home,dst=/home/runner/.cargo")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-amd64/rustup-home,dst=/home/runner/.rustup")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-amd64/cargo-target,dst=/home/runner/.cache/efr/cargo-target")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-amd64/sccache,dst=/home/runner/.cache/sccache")
	if slices.Contains(call.args, "--privileged") {
		t.Fatal("docker run arguments unexpectedly contain --privileged")
	}
	if strings.Contains(strings.Join(call.args, " "), jit) {
		t.Fatal("JIT config leaked into Docker command arguments")
	}
	if !slices.Contains(call.env, "ACTIONS_RUNNER_INPUT_JITCONFIG="+jit) {
		t.Fatalf("Docker command environment does not contain JIT config: %v", call.env)
	}
	assertCacheDirectories(t, createdDirs, "linux-amd64")
}

func TestDockerHostRunArgumentsARM64(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: successfulRunResults()}
	backend, createdDirs := newTestDockerHostBackend(fake, "linux/arm64")
	const jit = "secret-jit-config"
	if err := backend.Run(context.Background(), "repo-arm64-abc12", jit); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	call := fake.calls[2]
	assertContainsSequence(t, call.args, "--platform", "linux/arm64")
	assertContainsSequence(t, call.args, "--label", runnerSetLabel+"=repo-arm64")
	assertContainsSequence(t, call.args, "--label", runnerNameLabel+"=repo-arm64-abc12")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-arm64/cargo-home,dst=/home/runner/.cargo")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-arm64/rustup-home,dst=/home/runner/.rustup")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-arm64/cargo-target,dst=/home/runner/.cache/efr/cargo-target")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/linux-arm64/sccache,dst=/home/runner/.cache/sccache")
	if slices.Contains(call.args, "--privileged") {
		t.Fatal("docker run arguments unexpectedly contain --privileged")
	}
	assertCacheDirectories(t, createdDirs, "linux-arm64")
}

func TestDockerHostCachePlatformSegment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		platform string
		want     string
		wantErr  bool
	}{
		{platform: "linux/amd64", want: "linux-amd64"},
		{platform: "linux/arm64", want: "linux-arm64"},
		{platform: "linux/riscv64", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			got, err := dockerHostCachePlatformSegment(tt.platform)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("dockerHostCachePlatformSegment(%q) expected error", tt.platform)
				}
				return
			}
			if err != nil {
				t.Fatalf("dockerHostCachePlatformSegment(%q) error: %v", tt.platform, err)
			}
			if got != tt.want {
				t.Fatalf("dockerHostCachePlatformSegment(%q) = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestDockerHostRunRejectsEmptyJITConfig(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{}
	backend, _ := newTestDockerHostBackend(fake, "linux/amd64")
	if err := backend.Run(context.Background(), "repo-amd64-abc12", ""); err == nil {
		t.Fatal("Run() expected error for an empty JIT config")
	}
	if len(fake.calls) != 0 {
		t.Fatalf("Docker calls = %d, want 0", len(fake.calls))
	}
}

func TestDockerHostRunRedactsJITConfig(t *testing.T) {
	t.Parallel()
	const jit = "secret-jit-config"
	fake := &fakeDockerCommandRunner{results: []fakeDockerResult{
		{output: "orbstack\n"},
		{output: "unix:///tmp/orbstack/docker.sock\n"},
		{output: "invalid environment " + jit, err: errors.New("exit 125")},
	}}
	backend, _ := newTestDockerHostBackend(fake, "linux/amd64")
	var logs bytes.Buffer
	backend.logger = slog.New(slog.NewTextHandler(&logs, nil))
	err := backend.Run(context.Background(), "repo-amd64-abc12", jit)
	if err == nil {
		t.Fatal("Run() expected error")
	}
	if strings.Contains(err.Error(), jit) || strings.Contains(logs.String(), jit) {
		t.Fatalf("JIT config leaked: error=%q logs=%q", err, logs.String())
	}
	for _, required := range []string{"repo-amd64-abc12", defaultDockerHostRunnerImage, "linux/amd64", "orbstack"} {
		if !strings.Contains(err.Error(), required) {
			t.Fatalf("Run() error missing %q: %v", required, err)
		}
	}
}

func TestDockerHostCleanupRemovesOnlyLabeledRunner(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: []fakeDockerResult{
		{output: "orbstack\n"},
		{output: "unix:///tmp/orbstack/docker.sock\n"},
		{output: "container-id\n"},
		{},
	}}
	backend, _ := newTestDockerHostBackend(fake, "linux/amd64")
	backend.Cleanup(context.Background(), "repo-amd64-abc12")

	filters := strings.Join(fake.calls[2].args, " ")
	for _, label := range []string{managedLabel + "=true", runnerSetLabel + "=repo-amd64", runnerNameLabel + "=repo-amd64-abc12"} {
		if !strings.Contains(filters, label) {
			t.Fatalf("cleanup filters missing %q: %s", label, filters)
		}
	}
	assertContainsSequence(t, fake.calls[3].args, "rm", "-f", "-v", "container-id")
}

func TestDockerHostCleanupIsIdempotent(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: []fakeDockerResult{
		{output: "orbstack\n"},
		{output: "unix:///tmp/orbstack/docker.sock\n"},
		{output: "\n"},
	}}
	backend, _ := newTestDockerHostBackend(fake, "linux/amd64")
	backend.Cleanup(context.Background(), "missing-runner")
	if len(fake.calls) != 3 {
		t.Fatalf("Docker calls = %d, want no rm call", len(fake.calls))
	}
}

func TestDockerHostCleanupAllUsesManagementLabels(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: []fakeDockerResult{
		{output: "orbstack\n"},
		{output: "unix:///tmp/orbstack/docker.sock\n"},
		{output: "one\ntwo\n"},
		{},
	}}
	backend, _ := newTestDockerHostBackend(fake, "linux/amd64")
	backend.CleanupAll(context.Background(), "repo-amd64")
	filters := strings.Join(fake.calls[2].args, " ")
	if !strings.Contains(filters, managedLabel+"=true") || !strings.Contains(filters, runnerSetLabel+"=repo-amd64") {
		t.Fatalf("CleanupAll filters = %s", filters)
	}
	if strings.Contains(filters, "name=") || strings.Contains(filters, runnerNameLabel) {
		t.Fatalf("CleanupAll must not rely on name filters: %s", filters)
	}
	assertContainsSequence(t, fake.calls[3].args, "rm", "-f", "-v", "one", "two")
}

func successfulRunResults() []fakeDockerResult {
	return []fakeDockerResult{
		{output: "orbstack\n"},
		{output: "unix:///tmp/orbstack/docker.sock\n"},
		{output: "container-id\n"},
	}
}

func newTestDockerHostBackend(fake *fakeDockerCommandRunner, platform string) (result *DockerHostBackend, createdDirs *[]string) {
	runnerSet := "repo-amd64"
	if platform == "linux/arm64" {
		runnerSet = "repo-arm64"
	}
	result = NewDockerHostBackend(DockerHostOptions{
		RunnerSet:      runnerSet,
		Platform:       platform,
		CacheRoot:      "/cache/efr",
		CacheNamespace: "cloudspine",
	})
	result.runner = fake
	result.resolver = dockerContextResolver{runner: fake, getenv: func(string) string { return "" }}
	result.socketGroups = func(string) ([]string, error) { return []string{"20", "0"}, nil }
	created := []string{}
	result.mkdirAll = func(path string, _ os.FileMode) error {
		created = append(created, path)
		return nil
	}
	return result, &created
}

func assertCacheDirectories(t *testing.T, created *[]string, platformSegment string) {
	t.Helper()
	want := []string{
		"/cache/efr/cloudspine/" + platformSegment + "/cargo-home",
		"/cache/efr/cloudspine/" + platformSegment + "/rustup-home",
		"/cache/efr/cloudspine/" + platformSegment + "/cargo-target",
		"/cache/efr/cloudspine/" + platformSegment + "/sccache",
		"/cache/efr/shared/pnpm-store",
		"/cache/efr/shared/tool-cache",
	}
	if !equalStrings(*created, want) {
		t.Fatalf("cache directories = %v, want %v", *created, want)
	}
}

func assertContainsSequence(t *testing.T, values []string, sequence ...string) {
	t.Helper()
	for i := 0; i+len(sequence) <= len(values); i++ {
		if equalStrings(values[i:i+len(sequence)], sequence) {
			return
		}
	}
	t.Fatalf("%v does not contain sequence %v", values, sequence)
}

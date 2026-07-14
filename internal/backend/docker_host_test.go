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

func TestDockerHostRunArguments(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: successfulRunResults()}
	backend, createdDirs := newTestDockerHostBackend(fake)
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
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/shared/cargo-home,dst=/home/runner/.cargo")
	assertContainsSequence(t, call.args, "--mount", "type=bind,src=/cache/efr/cloudspine/cargo-target,dst=/home/runner/.cache/efr/cargo-target")
	if slices.Contains(call.args, "--privileged") {
		t.Fatal("docker run arguments unexpectedly contain --privileged")
	}
	if strings.Contains(strings.Join(call.args, " "), jit) {
		t.Fatal("JIT config leaked into Docker command arguments")
	}
	if !slices.Contains(call.env, "ACTIONS_RUNNER_INPUT_JITCONFIG="+jit) {
		t.Fatalf("Docker command environment does not contain JIT config: %v", call.env)
	}
	assertCacheDirectories(t, createdDirs)
}

func TestDockerHostRunRejectsEmptyJITConfig(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{}
	backend, _ := newTestDockerHostBackend(fake)
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
	backend, _ := newTestDockerHostBackend(fake)
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
	backend, _ := newTestDockerHostBackend(fake)
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
	backend, _ := newTestDockerHostBackend(fake)
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
	backend, _ := newTestDockerHostBackend(fake)
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

func newTestDockerHostBackend(fake *fakeDockerCommandRunner) (result *DockerHostBackend, createdDirs *[]string) {
	result = NewDockerHostBackend(DockerHostOptions{
		RunnerSet:      "repo-amd64",
		Platform:       "linux/amd64",
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

func assertCacheDirectories(t *testing.T, created *[]string) {
	t.Helper()
	want := []string{
		"/cache/efr/shared/cargo-home",
		"/cache/efr/cloudspine/cargo-target",
		"/cache/efr/cloudspine/sccache",
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

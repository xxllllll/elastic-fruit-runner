package backend

import (
	"context"
	"errors"
	"testing"
)

func TestDockerContextResolverUsesDockerHost(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{}
	resolver := dockerContextResolver{
		runner: fake,
		getenv: func(key string) string {
			if key == "DOCKER_HOST" {
				return "unix:///tmp/docker.sock"
			}
			return ""
		},
	}
	endpoint, err := resolver.resolve(context.Background())
	if err != nil {
		t.Fatalf("resolve() error: %v", err)
	}
	if endpoint.contextName != "DOCKER_HOST" || endpoint.socketPath != "/tmp/docker.sock" {
		t.Fatalf("endpoint = %+v", endpoint)
	}
	if len(fake.calls) != 0 {
		t.Fatalf("Docker CLI calls = %d, want 0", len(fake.calls))
	}
}

func TestDockerContextResolverUsesActiveContext(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: []fakeDockerResult{
		{output: "orbstack\n"},
		{output: "unix:///Users/test/.orbstack/run/docker.sock\n"},
	}}
	resolver := dockerContextResolver{runner: fake, getenv: func(string) string { return "" }}
	endpoint, err := resolver.resolve(context.Background())
	if err != nil {
		t.Fatalf("resolve() error: %v", err)
	}
	if endpoint.contextName != "orbstack" || !endpoint.useContext {
		t.Fatalf("endpoint = %+v", endpoint)
	}
	want := []string{"--context", "orbstack", "run"}
	if got := endpoint.commandArgs("run"); !equalStrings(got, want) {
		t.Fatalf("commandArgs = %v, want %v", got, want)
	}
}

func TestDockerContextResolverRejectsNonUnixEndpoint(t *testing.T) {
	t.Parallel()
	fake := &fakeDockerCommandRunner{results: []fakeDockerResult{
		{output: "remote\n"},
		{output: "tcp://docker.example:2376\n"},
	}}
	resolver := dockerContextResolver{runner: fake, getenv: func(string) string { return "" }}
	if _, err := resolver.resolve(context.Background()); err == nil {
		t.Fatal("resolve() expected error for TCP endpoint")
	}
}

func TestParseUnixDockerEndpoint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"unix:///tmp/docker.sock", "/tmp/docker.sock", false},
		{"unix://relative.sock", "", true},
		{"tcp://localhost:2375", "", true},
		{"/tmp/docker.sock", "", true},
	}
	for _, tt := range tests {
		got, err := parseUnixDockerEndpoint(tt.input)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Fatalf("parseUnixDockerEndpoint(%q) = %q, %v", tt.input, got, err)
		}
	}
}

type fakeDockerCall struct {
	env  []string
	args []string
}

type fakeDockerResult struct {
	output string
	err    error
}

type fakeDockerCommandRunner struct {
	calls   []fakeDockerCall
	results []fakeDockerResult
}

func (f *fakeDockerCommandRunner) CombinedOutput(_ context.Context, env []string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, fakeDockerCall{env: append([]string(nil), env...), args: append([]string(nil), args...)})
	if len(f.results) == 0 {
		return nil, errors.New("unexpected Docker command")
	}
	result := f.results[0]
	f.results = f.results[1:]
	return []byte(result.output), result.err
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

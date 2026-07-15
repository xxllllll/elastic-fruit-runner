package backend

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultDockerHostRunnerImage = "ghcr.io/xxllllll/elastic-fruit-runner/docker-host-runner@sha256:24d7af1adc02c8c5d21306752d3d31df1d693eeea0c9c59be4c3f481dc9911a8"
	managedLabel                 = "io.github.boring-design.elastic-fruit-runner.managed"
	runnerSetLabel               = "io.github.boring-design.elastic-fruit-runner.runner-set"
	runnerNameLabel              = "io.github.boring-design.elastic-fruit-runner.runner-name"
)

var dockerHostTracer = otel.Tracer("github.com/boring-design/elastic-fruit-runner/internal/backend/docker-host")

var _ Backend = (*DockerHostBackend)(nil)

type DockerHostOptions struct {
	RunnerSet      string
	Image          string
	Platform       string
	CacheRoot      string
	CacheNamespace string
}

type DockerHostBackend struct {
	options      DockerHostOptions
	runner       dockerCommandRunner
	resolver     dockerContextResolver
	socketGroups func(string) ([]string, error)
	mkdirAll     func(string, os.FileMode) error
	logger       *slog.Logger
}

type dockerRunInput struct {
	endpoint dockerEndpoint
	name     string
	groups   []string
	mounts   []string
}

type dockerCleanupRequest struct {
	endpoint dockerEndpoint
	ids      []string
	logKey   string
	logValue string
}

type dockerRunError struct {
	name        string
	contextName string
	err         error
}

func NewDockerHostBackend(options DockerHostOptions) *DockerHostBackend {
	runner := execDockerCommandRunner{}
	if options.Image == "" {
		options.Image = defaultDockerHostRunnerImage
	}
	return &DockerHostBackend{
		options:      options,
		runner:       runner,
		resolver:     dockerContextResolver{runner: runner, getenv: os.Getenv},
		socketGroups: dockerSocketGroupIDs,
		mkdirAll:     os.MkdirAll,
		logger:       slog.Default().With("backend", "docker-host", "runnerSet", options.RunnerSet),
	}
}

func (b *DockerHostBackend) Run(ctx context.Context, name, jitConfig string) error {
	ctx, span := dockerHostTracer.Start(ctx, "backend.docker_host.run",
		trace.WithAttributes(attribute.String("container.name", name)))
	defer span.End()
	if jitConfig == "" {
		err := fmt.Errorf("docker-host runner %q received an empty JIT config", name)
		return b.recordRunError(span, dockerRunError{name: name, contextName: "unresolved", err: err})
	}

	endpoint, err := b.resolver.resolve(ctx)
	if err != nil {
		return b.recordRunError(span, dockerRunError{name: name, contextName: "resolve Docker context", err: err})
	}
	groups, err := b.socketGroups(endpoint.socketPath)
	if err != nil {
		return b.recordRunError(span, dockerRunError{name: name, contextName: endpoint.contextName, err: err})
	}
	mounts, err := b.prepareCacheMounts()
	if err != nil {
		return b.recordRunError(span, dockerRunError{name: name, contextName: endpoint.contextName, err: err})
	}

	args := b.runArgs(dockerRunInput{endpoint: endpoint, name: name, groups: groups, mounts: mounts})
	env := []string{"ACTIONS_RUNNER_INPUT_JITCONFIG=" + jitConfig}
	out, err := b.runner.CombinedOutput(ctx, env, args...)
	if err == nil {
		return nil
	}
	output := redactSecret(string(out), jitConfig)
	err = fmt.Errorf("docker-host run failed: runner=%q image=%q platform=%q context=%q endpoint=%q output=%q: %w",
		name, b.options.Image, b.options.Platform, endpoint.contextName, endpoint.host, output, err)
	return b.recordRunError(span, dockerRunError{name: name, contextName: endpoint.contextName, err: err})
}

func (b *DockerHostBackend) runArgs(input dockerRunInput) []string {
	args := []string{"run", "-d", "--name", input.name, "--platform", b.options.Platform}
	args = append(args,
		"--label", managedLabel+"=true",
		"--label", runnerSetLabel+"="+b.options.RunnerSet,
		"--label", runnerNameLabel+"="+input.name,
		"--env", "ACTIONS_RUNNER_INPUT_JITCONFIG",
		"--env", "CARGO_HOME=/home/runner/.cargo",
		"--env", "RUSTUP_HOME=/home/runner/.rustup",
		"--env", "CARGO_TARGET_DIR=/home/runner/.cache/efr/cargo-target",
		"--env", "SCCACHE_DIR=/home/runner/.cache/sccache",
		"--env", "NPM_CONFIG_STORE_DIR=/home/runner/.cache/pnpm-store",
		"--env", "RUNNER_TOOL_CACHE=/opt/hostedtoolcache",
	)
	for _, group := range input.groups {
		args = append(args, "--group-add", group)
	}
	args = append(args, "--mount", bindMount(input.endpoint.socketPath, "/var/run/docker.sock"))
	for _, mount := range input.mounts {
		args = append(args, "--mount", mount)
	}
	args = append(args, b.options.Image)
	return input.endpoint.commandArgs(args...)
}

func (b *DockerHostBackend) Cleanup(ctx context.Context, name string) {
	endpoint, err := b.resolver.resolve(ctx)
	if err != nil {
		b.logger.Warn("resolve Docker context for cleanup", "runner", name, "err", err)
		return
	}
	ids, err := b.listManagedContainers(ctx, endpoint, name)
	if err != nil || len(ids) == 0 {
		if err != nil {
			b.logger.Warn("list runner containers for cleanup", "runner", name, "context", endpoint.contextName, "err", err)
		}
		return
	}
	b.removeContainers(ctx, dockerCleanupRequest{endpoint: endpoint, ids: ids, logKey: "runner", logValue: name})
}

func (b *DockerHostBackend) CleanupAll(ctx context.Context, prefix string) {
	if prefix != b.options.RunnerSet {
		b.logger.Warn("refusing cleanup for mismatched runner set", "requested", prefix)
		return
	}
	endpoint, err := b.resolver.resolve(ctx)
	if err != nil {
		b.logger.Warn("resolve Docker context for cleanup all", "err", err)
		return
	}
	ids, err := b.listManagedContainers(ctx, endpoint, "")
	if err != nil {
		b.logger.Warn("list managed containers for cleanup all", "context", endpoint.contextName, "err", err)
		return
	}
	b.removeContainers(ctx, dockerCleanupRequest{
		endpoint: endpoint,
		ids:      ids,
		logKey:   "runnerSet",
		logValue: b.options.RunnerSet,
	})
}

func (b *DockerHostBackend) listManagedContainers(ctx context.Context, endpoint dockerEndpoint, name string) ([]string, error) {
	args := []string{"ps", "-a", "-q",
		"--filter", "label=" + managedLabel + "=true",
		"--filter", "label=" + runnerSetLabel + "=" + b.options.RunnerSet,
	}
	if name != "" {
		args = append(args, "--filter", "label="+runnerNameLabel+"="+name)
	}
	out, err := b.runner.CombinedOutput(ctx, nil, endpoint.commandArgs(args...)...)
	if err != nil {
		return nil, fmt.Errorf("docker ps failed for context %q: output=%q: %w", endpoint.contextName, string(out), err)
	}
	return strings.Fields(string(out)), nil
}

func (b *DockerHostBackend) removeContainers(ctx context.Context, request dockerCleanupRequest) {
	if len(request.ids) == 0 {
		return
	}
	args := append([]string{"rm", "-f", "-v"}, request.ids...)
	out, err := b.runner.CombinedOutput(ctx, nil, request.endpoint.commandArgs(args...)...)
	if err != nil {
		b.logger.Warn("remove docker-host containers", request.logKey, request.logValue,
			"context", request.endpoint.contextName, "output", string(out), "err", err)
	}
}

func (b *DockerHostBackend) prepareCacheMounts() ([]string, error) {
	platformSegment, err := dockerHostCachePlatformSegment(b.options.Platform)
	if err != nil {
		return nil, err
	}
	projectCacheRoot := filepath.Join(b.options.CacheRoot, b.options.CacheNamespace, platformSegment)
	paths := []struct{ source, target string }{
		{filepath.Join(projectCacheRoot, "cargo-home"), "/home/runner/.cargo"},
		{filepath.Join(projectCacheRoot, "rustup-home"), "/home/runner/.rustup"},
		{filepath.Join(projectCacheRoot, "cargo-target"), "/home/runner/.cache/efr/cargo-target"},
		{filepath.Join(projectCacheRoot, "sccache"), "/home/runner/.cache/sccache"},
		{filepath.Join(b.options.CacheRoot, "shared", "pnpm-store"), "/home/runner/.cache/pnpm-store"},
		{filepath.Join(b.options.CacheRoot, "shared", "tool-cache"), "/opt/hostedtoolcache"},
	}
	mounts := make([]string, 0, len(paths))
	for _, item := range paths {
		if err := b.mkdirAll(item.source, 0o750); err != nil {
			return nil, fmt.Errorf("create docker-host cache directory %q: %w", item.source, err)
		}
		mounts = append(mounts, bindMount(item.source, item.target))
	}
	return mounts, nil
}

func dockerHostCachePlatformSegment(platform string) (string, error) {
	switch platform {
	case "linux/amd64":
		return "linux-amd64", nil
	case "linux/arm64":
		return "linux-arm64", nil
	default:
		return "", fmt.Errorf("unsupported docker-host cache platform %q", platform)
	}
}

func bindMount(source, target string) string {
	return "type=bind,src=" + source + ",dst=" + target
}

func redactSecret(value, secret string) string {
	if secret == "" {
		return value
	}
	return strings.ReplaceAll(value, secret, "[REDACTED]")
}

func dockerSocketGroupIDs(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat Docker Unix socket %q: %w", path, err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		return nil, fmt.Errorf("docker Unix endpoint %q is not a socket", path)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("read group ownership for Docker Unix socket %q", path)
	}
	groups := []string{strconv.FormatUint(uint64(stat.Gid), 10)}
	if runtime.GOOS == "darwin" && stat.Gid != 0 {
		groups = append(groups, "0")
	}
	return groups, nil
}

func (b *DockerHostBackend) recordRunError(span trace.Span, runError dockerRunError) error {
	span.RecordError(runError.err)
	span.SetStatus(codes.Error, runError.err.Error())
	b.logger.Error("docker-host runner start failed", "runner", runError.name,
		"context", runError.contextName, "err", runError.err)
	return runError.err
}

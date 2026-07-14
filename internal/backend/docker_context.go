package backend

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

const dockerEndpointFormat = "{{.Endpoints.docker.Host}}"

type dockerEndpoint struct {
	contextName string
	host        string
	socketPath  string
	useContext  bool
}

func (e dockerEndpoint) commandArgs(args ...string) []string {
	if !e.useContext {
		return args
	}
	result := make([]string, 0, len(args)+2)
	result = append(result, "--context", e.contextName)
	return append(result, args...)
}

type dockerContextResolver struct {
	runner dockerCommandRunner
	getenv func(string) string
}

func (r dockerContextResolver) resolve(ctx context.Context) (dockerEndpoint, error) {
	if host := strings.TrimSpace(r.getenv("DOCKER_HOST")); host != "" {
		path, err := parseUnixDockerEndpoint(host)
		if err != nil {
			return dockerEndpoint{}, fmt.Errorf("resolve Docker endpoint from DOCKER_HOST: %w", err)
		}
		return dockerEndpoint{contextName: "DOCKER_HOST", host: host, socketPath: path}, nil
	}

	name, err := r.currentContext(ctx)
	if err != nil {
		return dockerEndpoint{}, err
	}
	host, err := r.contextHost(ctx, name)
	if err != nil {
		return dockerEndpoint{}, err
	}
	path, err := parseUnixDockerEndpoint(host)
	if err != nil {
		return dockerEndpoint{}, fmt.Errorf("docker context %q endpoint %q: %w", name, host, err)
	}
	return dockerEndpoint{contextName: name, host: host, socketPath: path, useContext: true}, nil
}

func (r dockerContextResolver) currentContext(ctx context.Context) (string, error) {
	out, err := r.runner.CombinedOutput(ctx, nil, "context", "show")
	name := strings.TrimSpace(string(out))
	if err != nil {
		return "", fmt.Errorf("resolve active Docker context: output=%q: %w", name, err)
	}
	if name == "" {
		return "", fmt.Errorf("resolve active Docker context: Docker CLI returned an empty context name")
	}
	return name, nil
}

func (r dockerContextResolver) contextHost(ctx context.Context, name string) (string, error) {
	out, err := r.runner.CombinedOutput(ctx, nil, "context", "inspect", name, "--format", dockerEndpointFormat)
	host := strings.TrimSpace(string(out))
	if err != nil {
		return "", fmt.Errorf("inspect Docker context %q: output=%q: %w", name, host, err)
	}
	if host == "" {
		return "", fmt.Errorf("inspect Docker context %q: Docker CLI returned an empty endpoint", name)
	}
	return host, nil
}

func parseUnixDockerEndpoint(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse endpoint %q: %w", raw, err)
	}
	if u.Scheme != "unix" || u.Host != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("endpoint must use unix:///absolute/path, got %q", raw)
	}
	if !filepath.IsAbs(u.Path) {
		return "", fmt.Errorf("unix socket path must be absolute, got %q", u.Path)
	}
	return filepath.Clean(u.Path), nil
}

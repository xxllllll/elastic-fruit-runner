package backend

import (
	"context"
	"os"
	"os/exec"

	"github.com/boring-design/elastic-fruit-runner/internal/binpath"
)

type dockerCommandRunner interface {
	CombinedOutput(ctx context.Context, env []string, args ...string) ([]byte, error)
}

type execDockerCommandRunner struct{}

func (execDockerCommandRunner) CombinedOutput(ctx context.Context, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binpath.Lookup("docker"), args...)
	cmd.Env = mergeCommandEnvironment(os.Environ(), env)
	return cmd.CombinedOutput()
}

func mergeCommandEnvironment(base, overrides []string) []string {
	keys := make(map[string]struct{}, len(overrides))
	for _, value := range overrides {
		keys[environmentKey(value)] = struct{}{}
	}
	result := make([]string, 0, len(base)+len(overrides))
	for _, value := range base {
		if _, overridden := keys[environmentKey(value)]; !overridden {
			result = append(result, value)
		}
	}
	return append(result, overrides...)
}

func environmentKey(value string) string {
	for i := range value {
		if value[i] == '=' {
			return value[:i]
		}
	}
	return value
}

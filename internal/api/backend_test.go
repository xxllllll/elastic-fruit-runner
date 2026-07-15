package api

import (
	"testing"

	controlplanev1 "github.com/boring-design/elastic-fruit-runner/gen/controlplane/v1"
)

func TestToProtoBackendDockerHost(t *testing.T) {
	t.Parallel()
	if got := toProtoBackend("docker-host"); got != controlplanev1.Backend_BACKEND_DOCKER_HOST {
		t.Fatalf("toProtoBackend(docker-host) = %v", got)
	}
}

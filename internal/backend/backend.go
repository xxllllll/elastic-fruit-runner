package backend

import "context"

// Backend abstracts the runner execution environment.
// Implementations handle the full lifecycle: start the GitHub Actions runner
// with a JIT config, and clean up afterwards.
type Backend interface {
	// Run starts a runner instance with the given JIT config.
	// It sets up the execution environment and launches the GitHub Actions
	// runner process. It returns after the backend launch command succeeds;
	// GitHub JobStarted events are the authoritative confirmation of assignment.
	Run(ctx context.Context, name, jitConfig string) error

	// Cleanup tears down the execution environment.
	// Must be safe to call even if Run failed.
	Cleanup(ctx context.Context, name string)

	// CleanupAll removes resources owned by the specified runner set. Backend
	// implementations may use names, labels, or native resource metadata.
	// Called once at controller startup to remove resources from previous runs.
	CleanupAll(ctx context.Context, prefix string)
}

package backend

import (
	"slices"
	"testing"
)

func TestMergeCommandEnvironmentOverridesExistingValue(t *testing.T) {
	t.Parallel()
	got := mergeCommandEnvironment(
		[]string{"PATH=/usr/bin", "ACTIONS_RUNNER_INPUT_JITCONFIG=old"},
		[]string{"ACTIONS_RUNNER_INPUT_JITCONFIG=new"},
	)
	if slices.Contains(got, "ACTIONS_RUNNER_INPUT_JITCONFIG=old") {
		t.Fatalf("environment retained overridden JIT config: %v", got)
	}
	if !slices.Contains(got, "ACTIONS_RUNNER_INPUT_JITCONFIG=new") {
		t.Fatalf("environment missing new JIT config: %v", got)
	}
}

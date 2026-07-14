package config

import (
	"fmt"
	"path/filepath"
)

func validateDockerHostRunnerSet(rs *RunnerSetConfig, validation runnerSetValidation) error {
	prefix := validation.prefix
	if !validation.dockerHostAllowed {
		return fmt.Errorf("%s.backend docker-host is only supported for repository runner sets in the first phase", prefix)
	}
	if rs.MaxRunners != 1 {
		return fmt.Errorf("%s.max_runners must be 1 for docker-host in the first phase", prefix)
	}
	if rs.Platform == "" {
		rs.Platform = "linux/amd64"
	}
	if rs.Platform != "linux/amd64" {
		return fmt.Errorf("%s.platform must be linux/amd64 for docker-host in the first phase, got %q", prefix, rs.Platform)
	}
	if validation.cacheRoot == "" || !filepath.IsAbs(validation.cacheRoot) {
		return fmt.Errorf("cache_root must be an absolute path when docker-host is configured, got %q", validation.cacheRoot)
	}
	if !validCacheNamespace(rs.CacheNamespace) {
		return fmt.Errorf("%s.cache_namespace must contain only letters, digits, '.', '_' or '-', got %q", prefix, rs.CacheNamespace)
	}
	return nil
}

func validCacheNamespace(namespace string) bool {
	if namespace == "" || namespace == "." || namespace == ".." {
		return false
	}
	for _, r := range namespace {
		if !validCacheNamespaceRune(r) {
			return false
		}
	}
	return true
}

func validCacheNamespaceRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' ||
		r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-'
}

func countDockerHostRunnerSets(repos []RepoConfig) int {
	count := 0
	for _, repo := range repos {
		for _, runnerSet := range repo.RunnerSets {
			if runnerSet.Backend == "docker-host" {
				count++
			}
		}
	}
	return count
}

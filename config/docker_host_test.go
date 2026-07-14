package config

import (
	"strings"
	"testing"
	"time"
)

func validDockerHostConfig() Config {
	repo := validRepoConfig()
	repo.RunnerSets[0] = RunnerSetConfig{
		Name:           "repo-amd64",
		Backend:        "docker-host",
		MaxRunners:     1,
		CacheNamespace: "cloudspine",
	}
	return Config{
		Repos:       []RepoConfig{repo},
		IdleTimeout: 15 * time.Minute,
		CacheRoot:   "/cache/efr",
	}
}

func TestValidateDockerHost(t *testing.T) {
	t.Parallel()
	cfg := validDockerHostConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
	if got := cfg.Repos[0].RunnerSets[0].Platform; got != "linux/amd64" {
		t.Fatalf("Platform = %q, want linux/amd64 default", got)
	}
}

func TestValidateDockerHostARM64(t *testing.T) {
	t.Parallel()
	cfg := validDockerHostConfig()
	cfg.Repos[0].RunnerSets[0].Platform = "linux/arm64"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
}

func TestValidateDockerHostRestrictions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{"max runners", func(c *Config) { c.Repos[0].RunnerSets[0].MaxRunners = 2 }, "max_runners must be 1"},
		{"platform", func(c *Config) { c.Repos[0].RunnerSets[0].Platform = "linux/riscv64" }, "platform must be linux/amd64 or linux/arm64"},
		{"cache root", func(c *Config) { c.CacheRoot = "relative/cache" }, "cache_root must be an absolute path"},
		{"empty namespace", func(c *Config) { c.Repos[0].RunnerSets[0].CacheNamespace = "" }, "cache_namespace"},
		{"traversal namespace", func(c *Config) { c.Repos[0].RunnerSets[0].CacheNamespace = "../shared" }, "cache_namespace"},
		{"multiple sets", addSecondDockerHostSet, "at most one repository runner set"},
		{"organization scope", moveDockerHostToOrg, "only supported for repository runner sets"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validDockerHostConfig()
			tt.mutate(&cfg)
			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func addSecondDockerHostSet(cfg *Config) {
	second := validRepoConfig()
	second.Repo = "owner/other"
	second.RunnerSets[0] = cfg.Repos[0].RunnerSets[0]
	second.RunnerSets[0].Name = "other-amd64"
	second.RunnerSets[0].CacheNamespace = "other"
	cfg.Repos = append(cfg.Repos, second)
}

func moveDockerHostToOrg(cfg *Config) {
	org := validOrgConfig()
	org.RunnerSets[0] = cfg.Repos[0].RunnerSets[0]
	cfg.Orgs = []OrgConfig{org}
	cfg.Repos = nil
}

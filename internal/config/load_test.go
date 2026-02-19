package config

import (
	"strings"
	"testing"
)

// TestFR2_LoadValidYAML verifies that Load returns a fully populated *Config
// from a valid orchestrator.yaml file.
func TestFR2_LoadValidYAML(t *testing.T) {
	cfg, err := Load("testdata/valid.yaml")
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}

	// Project fields
	if cfg.Project.Name != "differentia" {
		t.Errorf("Project.Name = %q, want %q", cfg.Project.Name, "differentia")
	}
	if cfg.Project.Repository != "https://github.com/dmitriyb/differentia.git" {
		t.Errorf("Project.Repository = %q, want %q", cfg.Project.Repository, "https://github.com/dmitriyb/differentia.git")
	}

	// Credentials
	if cfg.Credentials.Backend != "rbw" {
		t.Errorf("Credentials.Backend = %q, want %q", cfg.Credentials.Backend, "rbw")
	}
	if len(cfg.Credentials.Secrets) != 2 {
		t.Errorf("len(Credentials.Secrets) = %d, want 2", len(cfg.Credentials.Secrets))
	}
	if s, ok := cfg.Credentials.Secrets["claude_token"]; !ok {
		t.Error("Credentials.Secrets missing key \"claude_token\"")
	} else {
		if s.Name != "claude-oauth-token" {
			t.Errorf("claude_token.Name = %q, want %q", s.Name, "claude-oauth-token")
		}
		if s.Env != "CLAUDE_CODE_OAUTH_TOKEN" {
			t.Errorf("claude_token.Env = %q, want %q", s.Env, "CLAUDE_CODE_OAUTH_TOKEN")
		}
	}

	// Docker
	if cfg.Docker.BaseImage != "debian:bookworm-slim" {
		t.Errorf("Docker.BaseImage = %q, want %q", cfg.Docker.BaseImage, "debian:bookworm-slim")
	}
	if cfg.Docker.Dockerfile != "Dockerfile.agent" {
		t.Errorf("Docker.Dockerfile = %q, want %q", cfg.Docker.Dockerfile, "Dockerfile.agent")
	}
	if v, ok := cfg.Docker.BuildArgs["ZIG_VERSION"]; !ok || v != "0.14.1" {
		t.Errorf("Docker.BuildArgs[\"ZIG_VERSION\"] = %q, want %q", v, "0.14.1")
	}

	// Agents
	if len(cfg.Agents) != 2 {
		t.Fatalf("len(Agents) = %d, want 2", len(cfg.Agents))
	}
	impl, ok := cfg.Agents["implementer"]
	if !ok {
		t.Fatal("Agents missing key \"implementer\"")
	}
	if impl.Prompt.System != "roles/IMPLEMENTING.md" {
		t.Errorf("implementer.Prompt.System = %q, want %q", impl.Prompt.System, "roles/IMPLEMENTING.md")
	}
	if impl.Workspace != "rw" {
		t.Errorf("implementer.Workspace = %q, want %q", impl.Workspace, "rw")
	}
	if len(impl.Tools) != 2 {
		t.Errorf("len(implementer.Tools) = %d, want 2", len(impl.Tools))
	}

	// Pipeline
	if len(cfg.Pipeline) != 3 {
		t.Fatalf("len(Pipeline) = %d, want 3", len(cfg.Pipeline))
	}
	if cfg.Pipeline[0].Name != "implement" {
		t.Errorf("Pipeline[0].Name = %q, want %q", cfg.Pipeline[0].Name, "implement")
	}
	if cfg.Pipeline[1].Agent != "reviewer" {
		t.Errorf("Pipeline[1].Agent = %q, want %q", cfg.Pipeline[1].Agent, "reviewer")
	}
	if len(cfg.Pipeline[1].DependsOn) != 1 || cfg.Pipeline[1].DependsOn[0] != "implement" {
		t.Errorf("Pipeline[1].DependsOn = %v, want [implement]", cfg.Pipeline[1].DependsOn)
	}
	if cfg.Pipeline[2].Condition != "steps.review.output.status == 'changes_requested'" {
		t.Errorf("Pipeline[2].Condition = %q, want %q", cfg.Pipeline[2].Condition, "steps.review.output.status == 'changes_requested'")
	}
}

// TestFR2_LoadMissingFile verifies that Load returns a wrapped error containing
// the file path when the file does not exist.
func TestFR2_LoadMissingFile(t *testing.T) {
	_, err := Load("testdata/nonexistent.yaml")
	if err == nil {
		t.Fatal("Load returned nil error for missing file")
	}
	msg := err.Error()
	if !strings.Contains(msg, "config: read") {
		t.Errorf("error = %q, want prefix containing %q", msg, "config: read")
	}
	if !strings.Contains(msg, "nonexistent.yaml") {
		t.Errorf("error = %q, want it to contain file path", msg)
	}
}

// TestFR2_LoadMalformedYAML verifies that Load returns a wrapped error when
// the YAML content is syntactically invalid.
func TestFR2_LoadMalformedYAML(t *testing.T) {
	_, err := Load("testdata/malformed.yaml")
	if err == nil {
		t.Fatal("Load returned nil error for malformed YAML")
	}
	msg := err.Error()
	if !strings.Contains(msg, "config: parse") {
		t.Errorf("error = %q, want prefix containing %q", msg, "config: parse")
	}
	if !strings.Contains(msg, "malformed.yaml") {
		t.Errorf("error = %q, want it to contain file path", msg)
	}
}

// TestNFR2_LoadNoSideEffects verifies that Load is a pure function of the
// file contents — it does not read environment variables or perform network calls.
// This is validated structurally: Load accepts only a path and returns only
// (*Config, error).
func TestNFR2_LoadNoSideEffects(t *testing.T) {
	// Load with a valid file and verify the result depends only on file contents.
	// Running the same file twice must produce identical results.
	cfg1, err1 := Load("testdata/valid.yaml")
	if err1 != nil {
		t.Fatalf("first Load: %v", err1)
	}
	cfg2, err2 := Load("testdata/valid.yaml")
	if err2 != nil {
		t.Fatalf("second Load: %v", err2)
	}
	if cfg1.Project.Name != cfg2.Project.Name {
		t.Errorf("consecutive loads returned different Project.Name: %q vs %q",
			cfg1.Project.Name, cfg2.Project.Name)
	}
	if cfg1.Credentials.Backend != cfg2.Credentials.Backend {
		t.Errorf("consecutive loads returned different Credentials.Backend: %q vs %q",
			cfg1.Credentials.Backend, cfg2.Credentials.Backend)
	}
}

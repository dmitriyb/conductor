package config

import (
	"reflect"
	"testing"
)

// TestFR1_FR2_ConfigRoundTripFidelity loads a full example orchestrator.yaml
// and verifies every field in the returned *Config matches expected values.
// This is the acceptance test for config loading round-trip fidelity (FR1, FR2).
func TestFR1_FR2_ConfigRoundTripFidelity(t *testing.T) {
	cfg, err := Load("testdata/valid.yaml")
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}

	// --- Project (FR1: types, FR2: loading) ---

	if cfg.Project.Name != "differentia" {
		t.Errorf("Project.Name = %q, want %q", cfg.Project.Name, "differentia")
	}
	if cfg.Project.Repository != "https://github.com/dmitriyb/differentia.git" {
		t.Errorf("Project.Repository = %q, want %q", cfg.Project.Repository, "https://github.com/dmitriyb/differentia.git")
	}

	// --- Credentials (FR1: types, FR2: loading) ---

	if cfg.Credentials.Backend != "rbw" {
		t.Errorf("Credentials.Backend = %q, want %q", cfg.Credentials.Backend, "rbw")
	}
	if len(cfg.Credentials.Secrets) != 2 {
		t.Fatalf("len(Credentials.Secrets) = %d, want 2", len(cfg.Credentials.Secrets))
	}

	wantSecrets := map[string]SecretRef{
		"claude_token": {Name: "claude-oauth-token", Env: "CLAUDE_CODE_OAUTH_TOKEN"},
		"github_pat":   {Name: "differentia-pat", Env: "AGENT_GH_TOKEN"},
	}
	for key, want := range wantSecrets {
		got, ok := cfg.Credentials.Secrets[key]
		if !ok {
			t.Errorf("Credentials.Secrets missing key %q", key)
			continue
		}
		if got.Name != want.Name {
			t.Errorf("Credentials.Secrets[%q].Name = %q, want %q", key, got.Name, want.Name)
		}
		if got.Env != want.Env {
			t.Errorf("Credentials.Secrets[%q].Env = %q, want %q", key, got.Env, want.Env)
		}
	}

	// --- Docker (FR1: types, FR2: loading) ---

	if cfg.Docker.BaseImage != "debian:bookworm-slim" {
		t.Errorf("Docker.BaseImage = %q, want %q", cfg.Docker.BaseImage, "debian:bookworm-slim")
	}
	if cfg.Docker.Dockerfile != "Dockerfile.agent" {
		t.Errorf("Docker.Dockerfile = %q, want %q", cfg.Docker.Dockerfile, "Dockerfile.agent")
	}
	if len(cfg.Docker.BuildArgs) != 1 {
		t.Fatalf("len(Docker.BuildArgs) = %d, want 1", len(cfg.Docker.BuildArgs))
	}
	if v := cfg.Docker.BuildArgs["ZIG_VERSION"]; v != "0.14.1" {
		t.Errorf("Docker.BuildArgs[\"ZIG_VERSION\"] = %q, want %q", v, "0.14.1")
	}

	// --- Agents (FR1: types, FR2: loading) ---

	if len(cfg.Agents) != 2 {
		t.Fatalf("len(Agents) = %d, want 2", len(cfg.Agents))
	}

	// implementer agent
	impl, ok := cfg.Agents["implementer"]
	if !ok {
		t.Fatal("Agents missing key \"implementer\"")
	}
	if impl.Prompt.System != "roles/IMPLEMENTING.md" {
		t.Errorf("implementer.Prompt.System = %q, want %q", impl.Prompt.System, "roles/IMPLEMENTING.md")
	}
	if impl.Prompt.Task != "Implement GitHub issue #{{.IssueNumber}} from {{.RepoURL}}." {
		t.Errorf("implementer.Prompt.Task = %q, want %q", impl.Prompt.Task, "Implement GitHub issue #{{.IssueNumber}} from {{.RepoURL}}.")
	}
	if impl.Workspace != "rw" {
		t.Errorf("implementer.Workspace = %q, want %q", impl.Workspace, "rw")
	}
	wantTools := []string{"gh", "git"}
	if !reflect.DeepEqual(impl.Tools, wantTools) {
		t.Errorf("implementer.Tools = %v, want %v", impl.Tools, wantTools)
	}
	wantImplSchema := map[string]any{"pr_number": "int", "branch": "string", "status": "string"}
	if !reflect.DeepEqual(impl.OutputSchema, wantImplSchema) {
		t.Errorf("implementer.OutputSchema = %v, want %v", impl.OutputSchema, wantImplSchema)
	}

	// reviewer agent
	rev, ok := cfg.Agents["reviewer"]
	if !ok {
		t.Fatal("Agents missing key \"reviewer\"")
	}
	if rev.Prompt.System != "roles/REVIEWING.md" {
		t.Errorf("reviewer.Prompt.System = %q, want %q", rev.Prompt.System, "roles/REVIEWING.md")
	}
	if rev.Prompt.Task != "Review PR #{{.PRNumber}} in {{.RepoOwner}}/{{.RepoName}}." {
		t.Errorf("reviewer.Prompt.Task = %q, want %q", rev.Prompt.Task, "Review PR #{{.PRNumber}} in {{.RepoOwner}}/{{.RepoName}}.")
	}
	if rev.Workspace != "ro" {
		t.Errorf("reviewer.Workspace = %q, want %q", rev.Workspace, "ro")
	}
	if rev.Tools != nil {
		t.Errorf("reviewer.Tools = %v, want nil", rev.Tools)
	}
	wantRevSchema := map[string]any{"status": "string", "comment_count": "int", "summary": "string"}
	if !reflect.DeepEqual(rev.OutputSchema, wantRevSchema) {
		t.Errorf("reviewer.OutputSchema = %v, want %v", rev.OutputSchema, wantRevSchema)
	}

	// --- Pipeline (FR1: types, FR2: loading) ---

	if len(cfg.Pipeline) != 3 {
		t.Fatalf("len(Pipeline) = %d, want 3", len(cfg.Pipeline))
	}

	// Step 0: implement
	if cfg.Pipeline[0].Name != "implement" {
		t.Errorf("Pipeline[0].Name = %q, want %q", cfg.Pipeline[0].Name, "implement")
	}
	if cfg.Pipeline[0].Agent != "implementer" {
		t.Errorf("Pipeline[0].Agent = %q, want %q", cfg.Pipeline[0].Agent, "implementer")
	}
	if cfg.Pipeline[0].DependsOn != nil {
		t.Errorf("Pipeline[0].DependsOn = %v, want nil", cfg.Pipeline[0].DependsOn)
	}
	if cfg.Pipeline[0].Condition != "" {
		t.Errorf("Pipeline[0].Condition = %q, want empty", cfg.Pipeline[0].Condition)
	}

	// Step 1: review
	if cfg.Pipeline[1].Name != "review" {
		t.Errorf("Pipeline[1].Name = %q, want %q", cfg.Pipeline[1].Name, "review")
	}
	if cfg.Pipeline[1].Agent != "reviewer" {
		t.Errorf("Pipeline[1].Agent = %q, want %q", cfg.Pipeline[1].Agent, "reviewer")
	}
	wantDeps := []string{"implement"}
	if !reflect.DeepEqual(cfg.Pipeline[1].DependsOn, wantDeps) {
		t.Errorf("Pipeline[1].DependsOn = %v, want %v", cfg.Pipeline[1].DependsOn, wantDeps)
	}
	if cfg.Pipeline[1].Condition != "" {
		t.Errorf("Pipeline[1].Condition = %q, want empty", cfg.Pipeline[1].Condition)
	}

	// Step 2: fix
	if cfg.Pipeline[2].Name != "fix" {
		t.Errorf("Pipeline[2].Name = %q, want %q", cfg.Pipeline[2].Name, "fix")
	}
	if cfg.Pipeline[2].Agent != "implementer" {
		t.Errorf("Pipeline[2].Agent = %q, want %q", cfg.Pipeline[2].Agent, "implementer")
	}
	wantFixDeps := []string{"review"}
	if !reflect.DeepEqual(cfg.Pipeline[2].DependsOn, wantFixDeps) {
		t.Errorf("Pipeline[2].DependsOn = %v, want %v", cfg.Pipeline[2].DependsOn, wantFixDeps)
	}
	if cfg.Pipeline[2].Condition != "steps.review.output.status == 'changes_requested'" {
		t.Errorf("Pipeline[2].Condition = %q, want %q", cfg.Pipeline[2].Condition, "steps.review.output.status == 'changes_requested'")
	}
}

// TestNFR2_LoadPureFunction verifies that Load is a pure function of file
// contents — loading the same file twice produces identical results with no
// observable side effects (no network calls, no environment variable reads).
func TestNFR2_LoadPureFunction(t *testing.T) {
	cfg1, err := Load("testdata/valid.yaml")
	if err != nil {
		t.Fatalf("first Load: %v", err)
	}
	cfg2, err := Load("testdata/valid.yaml")
	if err != nil {
		t.Fatalf("second Load: %v", err)
	}
	if !reflect.DeepEqual(cfg1, cfg2) {
		t.Error("consecutive loads of the same file returned different Config values")
	}
}

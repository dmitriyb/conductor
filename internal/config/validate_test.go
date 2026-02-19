package config

import (
	"strings"
	"testing"
)

// validConfig returns a minimal Config that passes Validate.
func validConfig() Config {
	return Config{
		Project: Project{
			Name:       "test-project",
			Repository: "https://github.com/test/repo.git",
		},
		Credentials: Credentials{
			Backend: "env",
		},
		Docker: Docker{
			BaseImage: "debian:bookworm-slim",
		},
		Agents: map[string]AgentDef{
			"worker": {
				Prompt:    PromptDef{System: "system.md", Task: "do the thing"},
				Workspace: "rw",
			},
		},
		Pipeline: []StepDef{
			{Name: "build", Agent: "worker"},
		},
	}
}

// TestFR3_ValidConfigReturnsNil verifies that a fully valid config produces no errors.
func TestFR3_ValidConfigReturnsNil(t *testing.T) {
	cfg := validConfig()
	err := Validate(&cfg)
	if err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
}

// TestFR3_RequiredFields verifies that missing required fields produce
// field-path errors.
func TestFR3_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name:    "missing project.name",
			mutate:  func(c *Config) { c.Project.Name = "" },
			wantErr: "project.name: required",
		},
		{
			name:    "missing project.repository",
			mutate:  func(c *Config) { c.Project.Repository = "" },
			wantErr: "project.repository: required",
		},
		{
			name:    "missing docker.base_image",
			mutate:  func(c *Config) { c.Docker.BaseImage = "" },
			wantErr: "docker.base_image: required",
		},
		{
			name:    "no agents defined",
			mutate:  func(c *Config) { c.Agents = nil },
			wantErr: "agents: at least one agent must be defined",
		},
		{
			name:    "empty pipeline",
			mutate:  func(c *Config) { c.Pipeline = nil },
			wantErr: "pipeline: at least one step required",
		},
		{
			name: "missing pipeline step name",
			mutate: func(c *Config) {
				c.Pipeline = []StepDef{{Agent: "worker"}}
			},
			wantErr: "pipeline[0].name: required",
		},
		{
			name: "missing pipeline step agent",
			mutate: func(c *Config) {
				c.Pipeline = []StepDef{{Name: "build"}}
			},
			wantErr: "pipeline[0].agent: required",
		},
		{
			name: "missing agent prompt.system",
			mutate: func(c *Config) {
				c.Agents = map[string]AgentDef{
					"worker": {Prompt: PromptDef{Task: "do"}, Workspace: "rw"},
				}
			},
			wantErr: "agents.worker.prompt.system: required",
		},
		{
			name: "missing agent prompt.task",
			mutate: func(c *Config) {
				c.Agents = map[string]AgentDef{
					"worker": {Prompt: PromptDef{System: "s.md"}, Workspace: "rw"},
				}
			},
			wantErr: "agents.worker.prompt.task: required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.mutate(&cfg)
			err := Validate(&cfg)
			if err == nil {
				t.Fatalf("Validate returned nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// TestFR3_InvalidCredentialBackend verifies that an unrecognised
// credentials.backend value produces a descriptive error.
func TestFR3_InvalidCredentialBackend(t *testing.T) {
	cfg := validConfig()
	cfg.Credentials.Backend = "vault"
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("Validate returned nil for invalid backend")
	}
	msg := err.Error()
	if !strings.Contains(msg, "credentials.backend") {
		t.Errorf("error = %q, want it to contain %q", msg, "credentials.backend")
	}
	if !strings.Contains(msg, `"vault"`) {
		t.Errorf("error = %q, want it to contain the invalid value %q", msg, "vault")
	}
}

// TestFR3_ValidCredentialBackends verifies all three allowed backends pass validation.
func TestFR3_ValidCredentialBackends(t *testing.T) {
	for _, backend := range []string{"rbw", "env", "file"} {
		t.Run(backend, func(t *testing.T) {
			cfg := validConfig()
			cfg.Credentials.Backend = backend
			err := Validate(&cfg)
			if err != nil {
				t.Fatalf("Validate returned error for valid backend %q: %v", backend, err)
			}
		})
	}
}

// TestFR3_UndefinedAgentReference verifies that a pipeline step referencing
// an agent not in the agents map produces a descriptive error.
func TestFR3_UndefinedAgentReference(t *testing.T) {
	cfg := validConfig()
	cfg.Pipeline = []StepDef{
		{Name: "step1", Agent: "nonexistent"},
	}
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("Validate returned nil for undefined agent reference")
	}
	msg := err.Error()
	if !strings.Contains(msg, "references undefined agent") {
		t.Errorf("error = %q, want it to contain %q", msg, "references undefined agent")
	}
	if !strings.Contains(msg, `"nonexistent"`) {
		t.Errorf("error = %q, want it to contain the agent name %q", msg, "nonexistent")
	}
	if !strings.Contains(msg, "pipeline[0].agent") {
		t.Errorf("error = %q, want it to contain field path %q", msg, "pipeline[0].agent")
	}
}

// TestFR3_UndefinedDependsOn verifies that a depends_on entry referencing
// a step name that does not exist produces an error.
func TestFR3_UndefinedDependsOn(t *testing.T) {
	cfg := validConfig()
	cfg.Pipeline = []StepDef{
		{Name: "step1", Agent: "worker", DependsOn: []string{"phantom"}},
	}
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("Validate returned nil for undefined depends_on reference")
	}
	msg := err.Error()
	if !strings.Contains(msg, "pipeline[0].depends_on") {
		t.Errorf("error = %q, want it to contain field path %q", msg, "pipeline[0].depends_on")
	}
	if !strings.Contains(msg, `"phantom"`) {
		t.Errorf("error = %q, want it to contain the step name %q", msg, "phantom")
	}
}

// TestFR3_InvalidWorkspace verifies that an invalid agent workspace value
// produces a descriptive error.
func TestFR3_InvalidWorkspace(t *testing.T) {
	cfg := validConfig()
	cfg.Agents = map[string]AgentDef{
		"worker": {
			Prompt:    PromptDef{System: "s.md", Task: "do"},
			Workspace: "readwrite",
		},
	}
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("Validate returned nil for invalid workspace")
	}
	msg := err.Error()
	if !strings.Contains(msg, "agents.worker.workspace") {
		t.Errorf("error = %q, want it to contain field path %q", msg, "agents.worker.workspace")
	}
	if !strings.Contains(msg, "must be rw or ro") {
		t.Errorf("error = %q, want it to contain %q", msg, "must be rw or ro")
	}
}

// TestNFR1_MultipleErrorsCollected verifies that Validate collects all
// errors and returns them together via errors.Join, not just the first.
func TestNFR1_MultipleErrorsCollected(t *testing.T) {
	cfg := Config{} // everything missing
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("Validate returned nil for empty config")
	}
	msg := err.Error()
	// Should contain at least these distinct field-path errors.
	wantSubstrings := []string{
		"project.name: required",
		"project.repository: required",
		"credentials.backend",
		"docker.base_image: required",
		"agents: at least one agent must be defined",
		"pipeline: at least one step required",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(msg, want) {
			t.Errorf("error = %q, want it to contain %q", msg, want)
		}
	}
}

// TestFR3_ValidDependsOnChain verifies that depends_on referencing earlier
// steps in the pipeline is accepted.
func TestFR3_ValidDependsOnChain(t *testing.T) {
	cfg := validConfig()
	cfg.Agents = map[string]AgentDef{
		"a": {Prompt: PromptDef{System: "s.md", Task: "do"}, Workspace: "rw"},
		"b": {Prompt: PromptDef{System: "s.md", Task: "do"}, Workspace: "ro"},
	}
	cfg.Pipeline = []StepDef{
		{Name: "first", Agent: "a"},
		{Name: "second", Agent: "b", DependsOn: []string{"first"}},
	}
	err := Validate(&cfg)
	if err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
}

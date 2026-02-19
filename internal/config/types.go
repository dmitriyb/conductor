package config

// Config is the top-level structure mapping to orchestrator.yaml.
type Config struct {
	Project     Project              `yaml:"project"`
	Credentials Credentials          `yaml:"credentials"`
	Docker      Docker               `yaml:"docker"`
	Agents      map[string]AgentDef  `yaml:"agents"`
	Pipeline    []StepDef            `yaml:"pipeline"`
}

// Project identifies the target repository.
type Project struct {
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
}

// Credentials configures the secret backend and its entries.
type Credentials struct {
	Backend string               `yaml:"backend"` // rbw | env | file
	Secrets map[string]SecretRef `yaml:"secrets"`
}

// SecretRef maps a backend-specific key to a container environment variable.
type SecretRef struct {
	Name string `yaml:"name"` // backend-specific key
	Env  string `yaml:"env"`  // env var name to expose in container
}

// Docker holds container build configuration.
type Docker struct {
	BaseImage  string            `yaml:"base_image"`
	Dockerfile string            `yaml:"dockerfile"`
	BuildArgs  map[string]string `yaml:"build_args"`
}

// AgentDef defines a single agent's prompt, workspace mode, and capabilities.
type AgentDef struct {
	Prompt       PromptDef      `yaml:"prompt"`
	Workspace    string         `yaml:"workspace"` // rw | ro
	OutputSchema map[string]any `yaml:"output_schema"`
	Tools        []string       `yaml:"tools"`
}

// PromptDef holds the system and task prompt templates for an agent.
type PromptDef struct {
	System string `yaml:"system"` // file path or inline text
	Task   string `yaml:"task"`   // Go text/template
}

// StepDef defines a single pipeline step.
type StepDef struct {
	Name      string   `yaml:"name"`
	Agent     string   `yaml:"agent"`
	DependsOn []string `yaml:"depends_on"`
	Condition string   `yaml:"condition"` // CEL expression, empty = always
}

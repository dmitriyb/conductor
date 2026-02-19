package config

import (
	"errors"
	"fmt"
)

// Validate checks all Config fields for completeness and consistency.
// It collects all errors and returns them via errors.Join.
func Validate(cfg *Config) error {
	var errs []error
	check := func(cond bool, path, msg string) {
		if !cond {
			errs = append(errs, fmt.Errorf("%s: %s", path, msg))
		}
	}

	check(cfg.Project.Name != "", "project.name", "required")
	check(cfg.Project.Repository != "", "project.repository", "required")

	validBackends := map[string]bool{"rbw": true, "env": true, "file": true}
	check(validBackends[cfg.Credentials.Backend], "credentials.backend",
		fmt.Sprintf("must be one of: rbw, env, file (got %q)", cfg.Credentials.Backend))
	check(cfg.Docker.BaseImage != "", "docker.base_image", "required")
	check(len(cfg.Agents) > 0, "agents", "at least one agent must be defined")

	for name, agent := range cfg.Agents {
		p := "agents." + name
		check(agent.Prompt.System != "", p+".prompt.system", "required")
		check(agent.Prompt.Task != "", p+".prompt.task", "required")
		check(agent.Workspace == "rw" || agent.Workspace == "ro",
			p+".workspace", fmt.Sprintf("must be rw or ro (got %q)", agent.Workspace))
	}

	check(len(cfg.Pipeline) > 0, "pipeline", "at least one step required")
	stepNames := map[string]bool{}
	for i, step := range cfg.Pipeline {
		p := fmt.Sprintf("pipeline[%d]", i)
		check(step.Name != "", p+".name", "required")
		if step.Agent != "" {
			_, ok := cfg.Agents[step.Agent]
			check(ok, p+".agent", fmt.Sprintf("references undefined agent %q", step.Agent))
		}
		for _, dep := range step.DependsOn {
			check(stepNames[dep], p+".depends_on", fmt.Sprintf("unknown step %q", dep))
		}
		stepNames[step.Name] = true
	}
	return errors.Join(errs...)
}

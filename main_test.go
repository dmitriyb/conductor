package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// validYAML is a minimal orchestrator.yaml that passes Validate.
const validYAML = `
project:
  name: test-project
  repository: https://github.com/test/repo.git

credentials:
  backend: env

docker:
  base_image: debian:bookworm-slim

agents:
  worker:
    prompt:
      system: system.md
      task: "do the thing"
    workspace: rw

pipeline:
  - { name: build, agent: worker }
`

// writeConfig writes YAML content to a temp file and returns its path.
func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "orchestrator.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

// TestFR4_ValidateSubcommandSuccess verifies that `conductor validate --config path`
// exits 0 and prints a success message for a valid config.
func TestFR4_ValidateSubcommandSuccess(t *testing.T) {
	cfgPath := writeConfig(t, validYAML)
	var stdout, stderr bytes.Buffer

	code := run([]string{"--config", cfgPath, "validate"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("want exit 0, got %d; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "configuration is valid") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), "configuration is valid")
	}
}

// TestFR4_ValidateDefaultConfig verifies that `conductor validate` defaults
// to ./orchestrator.yaml when --config is not specified.
func TestFR4_ValidateDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "orchestrator.yaml")
	if err := os.WriteFile(cfgPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Change to the temp directory so the default path resolves.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	var stdout, stderr bytes.Buffer
	code := run([]string{"validate"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("want exit 0, got %d; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "configuration is valid") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), "configuration is valid")
	}
}

// TestFR4_ValidateInvalidConfig verifies that `conductor validate` exits
// non-zero when the config has validation errors.
func TestFR4_ValidateInvalidConfig(t *testing.T) {
	invalidYAML := `
project:
  name: ""
  repository: ""
credentials:
  backend: env
docker:
  base_image: debian:bookworm-slim
agents:
  worker:
    prompt:
      system: s.md
      task: do
    workspace: rw
pipeline:
  - { name: build, agent: worker }
`
	cfgPath := writeConfig(t, invalidYAML)
	var stdout, stderr bytes.Buffer

	code := run([]string{"--config", cfgPath, "validate"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("want non-zero exit for invalid config, got 0")
	}
}

// TestFR4_ValidateMissingConfig verifies that `conductor validate` exits
// non-zero when the config file does not exist.
func TestFR4_ValidateMissingConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--config", "/nonexistent/path.yaml", "validate"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("want non-zero exit for missing config, got 0")
	}
}

// TestFR4_RunSubcommandRecognized verifies that the `run` subcommand is
// recognized and loads config before reporting unimplemented.
func TestFR4_RunSubcommandRecognized(t *testing.T) {
	cfgPath := writeConfig(t, validYAML)
	var stdout, stderr bytes.Buffer

	_ = run([]string{"--config", cfgPath, "run"}, &stdout, &stderr)

	// run is a stub that exits non-zero, but should not report "unknown subcommand"
	if strings.Contains(stderr.String(), "unknown subcommand") {
		t.Fatalf("run should be recognized, stderr: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "not yet implemented") {
		t.Fatalf("stderr = %q, want it to contain %q", stderr.String(), "not yet implemented")
	}
}

// TestFR4_BuildSubcommandRecognized verifies that the `build` subcommand is
// recognized and loads config before reporting unimplemented.
func TestFR4_BuildSubcommandRecognized(t *testing.T) {
	cfgPath := writeConfig(t, validYAML)
	var stdout, stderr bytes.Buffer

	_ = run([]string{"--config", cfgPath, "build"}, &stdout, &stderr)
	if strings.Contains(stderr.String(), "unknown subcommand") {
		t.Fatalf("build should be recognized, stderr: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "not yet implemented") {
		t.Fatalf("stderr = %q, want it to contain %q", stderr.String(), "not yet implemented")
	}
}

// TestFR4_UnknownSubcommand verifies that an unknown subcommand prints
// usage and exits non-zero.
func TestFR4_UnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"frobnicate"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("want non-zero exit for unknown subcommand, got 0")
	}
	if !strings.Contains(stderr.String(), "unknown subcommand") {
		t.Fatalf("stderr = %q, want it to contain %q", stderr.String(), "unknown subcommand")
	}
	if !strings.Contains(stderr.String(), "frobnicate") {
		t.Fatalf("stderr = %q, want it to contain the bad subcommand name", stderr.String())
	}
}

// TestFR4_NoSubcommand verifies that running with no subcommand prints
// usage and exits non-zero.
func TestFR4_NoSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("want non-zero exit for missing subcommand, got 0")
	}
	if !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("stderr = %q, want it to contain usage info", stderr.String())
	}
}

// TestFR4_LogLevelFlag verifies that the --log-level flag is accepted
// and does not cause an error.
func TestFR4_LogLevelFlag(t *testing.T) {
	cfgPath := writeConfig(t, validYAML)
	for _, level := range []string{"debug", "info", "warn", "error"} {
		t.Run(level, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{"--log-level", level, "--config", cfgPath, "validate"}, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("want exit 0 with --log-level %s, got %d; stderr: %s", level, code, stderr.String())
			}
		})
	}
}

// TestFR6_NoGlobalState verifies that *Config is loaded via the --config
// flag and passed explicitly — not via global state. Multiple invocations
// with different configs produce independent results.
func TestFR6_NoGlobalState(t *testing.T) {
	// First call with valid config
	validPath := writeConfig(t, validYAML)
	var stdout1, stderr1 bytes.Buffer
	code1 := run([]string{"--config", validPath, "validate"}, &stdout1, &stderr1)

	// Second call with a missing config
	var stdout2, stderr2 bytes.Buffer
	code2 := run([]string{"--config", "/nonexistent.yaml", "validate"}, &stdout2, &stderr2)

	if code1 != 0 {
		t.Fatalf("first run: want exit 0, got %d", code1)
	}
	if code2 == 0 {
		t.Fatal("second run: want non-zero exit, got 0")
	}
}

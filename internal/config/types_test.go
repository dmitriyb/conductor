package config

import (
	"reflect"
	"testing"
)

// TestFR1_StructYAMLTags verifies that all config structs have correct yaml
// tags matching the expected YAML keys from the orchestrator.yaml schema.
func TestFR1_StructYAMLTags(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		wantTags map[string]string // field name → yaml tag value
	}{
		{
			"Config",
			reflect.TypeOf(Config{}),
			map[string]string{
				"Project":     "project",
				"Credentials": "credentials",
				"Docker":      "docker",
				"Agents":      "agents",
				"Pipeline":    "pipeline",
			},
		},
		{
			"Project",
			reflect.TypeOf(Project{}),
			map[string]string{
				"Name":       "name",
				"Repository": "repository",
			},
		},
		{
			"Credentials",
			reflect.TypeOf(Credentials{}),
			map[string]string{
				"Backend": "backend",
				"Secrets": "secrets",
			},
		},
		{
			"SecretRef",
			reflect.TypeOf(SecretRef{}),
			map[string]string{
				"Name": "name",
				"Env":  "env",
			},
		},
		{
			"Docker",
			reflect.TypeOf(Docker{}),
			map[string]string{
				"BaseImage":  "base_image",
				"Dockerfile": "dockerfile",
				"BuildArgs":  "build_args",
			},
		},
		{
			"AgentDef",
			reflect.TypeOf(AgentDef{}),
			map[string]string{
				"Prompt":       "prompt",
				"Workspace":    "workspace",
				"OutputSchema": "output_schema",
				"Tools":        "tools",
			},
		},
		{
			"PromptDef",
			reflect.TypeOf(PromptDef{}),
			map[string]string{
				"System": "system",
				"Task":   "task",
			},
		},
		{
			"StepDef",
			reflect.TypeOf(StepDef{}),
			map[string]string{
				"Name":      "name",
				"Agent":     "agent",
				"DependsOn": "depends_on",
				"Condition": "condition",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for fieldName, wantTag := range tt.wantTags {
				field, ok := tt.typ.FieldByName(fieldName)
				if !ok {
					t.Errorf("field %s not found on %s", fieldName, tt.name)
					continue
				}
				gotTag := field.Tag.Get("yaml")
				if gotTag != wantTag {
					t.Errorf("%s.%s: yaml tag = %q, want %q", tt.name, fieldName, gotTag, wantTag)
				}
			}
		})
	}
}

// TestFR1_StructFieldTypes verifies that collection fields use the correct
// Go types (maps for named collections, slices for ordered lists).
func TestFR1_StructFieldTypes(t *testing.T) {
	tests := []struct {
		name      string
		typ       reflect.Type
		field     string
		wantKind  reflect.Kind
		wantElem  string // element type name for maps/slices
	}{
		{"Config.Agents", reflect.TypeOf(Config{}), "Agents", reflect.Map, "AgentDef"},
		{"Config.Pipeline", reflect.TypeOf(Config{}), "Pipeline", reflect.Slice, "StepDef"},
		{"Credentials.Secrets", reflect.TypeOf(Credentials{}), "Secrets", reflect.Map, "SecretRef"},
		{"Docker.BuildArgs", reflect.TypeOf(Docker{}), "BuildArgs", reflect.Map, "string"},
		{"AgentDef.Tools", reflect.TypeOf(AgentDef{}), "Tools", reflect.Slice, "string"},
		{"StepDef.DependsOn", reflect.TypeOf(StepDef{}), "DependsOn", reflect.Slice, "string"},
		{"AgentDef.OutputSchema", reflect.TypeOf(AgentDef{}), "OutputSchema", reflect.Map, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := tt.typ.FieldByName(tt.field)
			if !ok {
				t.Fatalf("field %s not found", tt.field)
			}
			if field.Type.Kind() != tt.wantKind {
				t.Errorf("kind = %v, want %v", field.Type.Kind(), tt.wantKind)
			}
			var elemType reflect.Type
			switch tt.wantKind {
			case reflect.Map:
				elemType = field.Type.Elem()
			case reflect.Slice:
				elemType = field.Type.Elem()
			}
			if elemType != nil && elemType.Name() != tt.wantElem {
				t.Errorf("element type = %q, want %q", elemType.Name(), tt.wantElem)
			}
		})
	}
}

// TestFR1_StructFieldCompleteness verifies that each struct has exactly the
// expected number of fields, catching accidental additions or omissions.
func TestFR1_StructFieldCompleteness(t *testing.T) {
	tests := []struct {
		name      string
		typ       reflect.Type
		wantCount int
	}{
		{"Config", reflect.TypeOf(Config{}), 5},
		{"Project", reflect.TypeOf(Project{}), 2},
		{"Credentials", reflect.TypeOf(Credentials{}), 2},
		{"SecretRef", reflect.TypeOf(SecretRef{}), 2},
		{"Docker", reflect.TypeOf(Docker{}), 3},
		{"AgentDef", reflect.TypeOf(AgentDef{}), 4},
		{"PromptDef", reflect.TypeOf(PromptDef{}), 2},
		{"StepDef", reflect.TypeOf(StepDef{}), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.typ.NumField()
			if got != tt.wantCount {
				t.Errorf("%s has %d fields, want %d", tt.name, got, tt.wantCount)
			}
		})
	}
}

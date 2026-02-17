# Config Module — Architecture

## 1. Component Diagram

```
┌──────────────────────────────────────────────────────┐
│                      CLI Layer                       │
│  main.go: parse flags → dispatch run/validate/build  │
└──────────────┬───────────────────────────┬───────────┘
               │ --config path             │ --log-level
               ▼                           ▼
┌──────────────────────┐    ┌─────────────────────────┐
│      config.Load()   │    │     logging.Init()       │
│  file → yaml.v3 →    │    │  slog.New(handler)       │
│  *Config             │    │  returns *slog.Logger     │
└──────────┬───────────┘    └─────────────────────────┘
           │
           ▼
┌──────────────────────┐
│   config.Validate()  │
│  *Config → []error   │
│  field-path errors   │
└──────────┬───────────┘
           │
           ▼
   validated *Config
   passed to infra,
   agent, pipeline
```

## 2. Package Layout

```
conductor/
├── main.go              CLI entry point, flag parsing
├── internal/
│   └── config/
│       ├── types.go     Config struct tree (§3)
│       ├── load.go      YAML loading (FR2)
│       ├── validate.go  Validation rules (FR3)
│       └── logging.go   slog initialization (FR5)
```

All config types live in `internal/config`. The package exports `Load`,
`Validate`, and `InitLogging`. No sub-packages.

## 3. Data Model

The struct tree mirrors the YAML structure exactly. Top-level sections map to
named struct fields with `yaml` tags.

```
Config
├── Project
│   ├── Name        string
│   └── Repository  string
├── Credentials
│   ├── Backend     string          (rbw | env | file)
│   └── Secrets     map[string]SecretRef
├── Docker
│   ├── BaseImage   string
│   ├── Dockerfile  string          (optional override)
│   └── BuildArgs   map[string]string
├── Agents          map[string]AgentDef
│   └── AgentDef
│       ├── Prompt       PromptDef
│       │   ├── System   string     (path or inline)
│       │   └── Task     string     (Go template)
│       ├── Workspace    string     (rw | ro)
│       ├── OutputSchema map[string]any (JSON schema)
│       └── Tools        []string
└── Pipeline        []StepDef
    └── StepDef
        ├── Name      string
        ├── Agent     string        (key into Agents map)
        ├── DependsOn []string
        └── Condition string        (CEL expression)
```

## 4. Data Flow

```
orchestrator.yaml
       │
       │  os.ReadFile
       ▼
  []byte (raw YAML)
       │
       │  yaml.Unmarshal
       ▼
  *Config (typed but unvalidated)
       │
       │  Validate()
       ▼
  *Config (validated) ──► returned to caller
       │                   or []error on failure
```

## 5. Design Decisions

**D1 — yaml.v3 over encoding/json**
YAML is the only supported format. Using yaml.v3 directly avoids a JSON
intermediate layer. The struct tags are `yaml:"field_name"`.

**D2 — Validation as a separate pass**
`Load` and `Validate` are separate functions. This lets `validate` CLI print
partial results even when validation fails, and allows tests to construct
`Config` values directly without going through YAML.

**D3 — No global logger**
`InitLogging` returns a `*slog.Logger`. Each module creates a child logger
with `logger.With("component", "name")`. No `slog.SetDefault`.

**D4 — Flat package, no sub-packages**
At ~400 LOC the config module does not warrant sub-packages. All types,
loading, and validation live in `internal/config`.

**D5 — No config hot-reload**
The config is loaded once at startup. No file watching, no SIGHUP reload.
This keeps the module simple and avoids concurrency concerns.

**D6 — Multi-error aggregation**
Validation collects all errors rather than failing on the first. This gives
the user a complete list of problems in a single run of `conductor validate`.
Uses `errors.Join` from the standard library.

# Agent Module — Architecture

## 1. Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                       agent package                         │
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │  Template     │    │  Runner      │    │  Parser      │   │
│  │  Renderer     │    │              │    │              │   │
│  │              │    │  prepares     │    │  scans for   │   │
│  │  renders     │    │  mounts,     │    │  marker,     │   │
│  │  system +    │    │  env, init   │    │  extracts    │   │
│  │  task prompts│    │  script;     │    │  JSON,       │   │
│  │  via         │    │  calls       │    │  validates   │   │
│  │  text/template    │  docker run  │    │  against     │   │
│  │              │    │              │    │  schema      │   │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘   │
│         │                   │                   │           │
│         └───────┬───────────┘                   │           │
│                 │                               │           │
│                 ▼                               │           │
│         ┌──────────────┐                        │           │
│         │  RunAgent()  │◄───────────────────────┘           │
│         │              │                                    │
│         │  orchestrates│                                    │
│         │  render →    │                                    │
│         │  run →       │                                    │
│         │  parse       │                                    │
│         └──────┬───────┘                                    │
│                │                                            │
│                ▼                                            │
│         StepResult                                          │
└─────────────────────────────────────────────────────────────┘
```

## 2. Package Layout

```
internal/
└── agent/
    ├── runner.go      RunAgent entry point, container lifecycle
    ├── template.go    Prompt rendering via text/template
    ├── parser.go      Output marker parsing + JSON validation
    └── result.go      StepResult type definition
```

## 3. Container Mount Layout

```
Host                              Container
─────────────────────────         ────────────────────────
<clone>/repo/               ──►   /workspace        (rw or ro)
<tmpdir>/system-prompt.txt  ──►   /tmp/orchestrator-prompts/
<tmpdir>/task-prompt.txt          system-prompt.txt  (ro)
                                  task-prompt.txt    (ro)
~/.claude/skills/           ──►   /home/agent/.claude/skills/ (ro)
<ssh-agent-sock>            ──►   /tmp/ssh-agent.sock
```

Workspace mode (rw/ro) is determined by `AgentDef.Workspace`.

## 4. Agent Execution Flow

```
RunAgent(ctx, agentName, agentDef, templateData, runCfg)
    │
    ├── 1. Render prompts
    │     template.go: render system + task templates
    │     write to tmpdir as prompt files
    │
    ├── 2. Generate init script
    │     runner.go: bash script that configures git,
    │     gh auth, SSH signing, then exec claude
    │
    ├── 3. Build docker args
    │     runner.go: --cap-drop, -u, --env-file,
    │     -v mounts, image name, init script
    │
    ├── 4. Execute container
    │     exec.CommandContext(ctx, "docker", "run", ...)
    │     capture stdout+stderr
    │
    ├── 5. Write log file
    │     runner.go: save full output to log dir
    │
    ├── 6. Parse output
    │     parser.go: find ###PIPELINE_OUTPUT### marker
    │     extract JSON, validate against schema
    │
    └── 7. Return StepResult
          result.go: status, output map, log path
```

## 5. Template Data Model

Templates receive a `TemplateData` struct. Fields are populated from the
pipeline context — issue number, repo URL, PR numbers from prior steps, and
the full output map of each completed step.

```
TemplateData
├── IssueNumber  string
├── RepoURL      string
├── RepoOwner    string
├── RepoName     string
├── PRNumber     string
└── Steps        map[string]StepResult
    └── StepResult
        ├── Status   string
        └── Output   map[string]any
```

A task template accesses prior outputs as `{{.Steps.implement.Output.pr_number}}`.

## 6. Output Protocol

The agent protocol defines a marker-based output contract:

```
<arbitrary agent output>
...
###PIPELINE_OUTPUT###{"status":"success","pr_number":42,"branch":"issue-55"}
```

The parser scans all lines for the marker prefix. If multiple matches exist,
the last one wins (agents may print intermediate markers during retries).
The JSON payload after the marker is extracted, parsed, and validated.

## 7. Design Decisions

**D1 — text/template over more powerful engines**
Go's `text/template` is sufficient for prompt variable substitution. It
avoids external dependencies and is well-understood. Complex logic belongs
in CEL conditions (pipeline module), not in prompt templates.

**D2 — Init script as bash string**
The init script is passed as a single string argument to the container's
`ENTRYPOINT ["/bin/bash", "-c"]`. This avoids mounting an extra script file
and matches the existing orchestrator pattern.

**D3 — Schema validation is lightweight**
Output schema validation checks key presence and basic type matching
(`string`, `int`), not full JSON Schema. This is sufficient for the
pipeline's needs and avoids a JSON Schema library dependency.

**D4 — One container per agent invocation**
Each `RunAgent` call creates and destroys exactly one container. There is no
container pooling or reuse. This provides clean isolation between steps.

**D5 — Log file per agent run**
Each agent run gets its own timestamped log file. This allows post-mortem
debugging without sifting through a single monolithic log.

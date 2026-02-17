# Agent Module — Implementation

## 1. StepResult Type

```go
// internal/agent/result.go

type StepResult struct {
    Name    string         // step name
    Agent   string         // agent name
    Status  string         // "success", "failure", or "skipped"
    Output  map[string]any // parsed JSON payload
    LogPath string         // path to full output log
    Error   string         // error description if failure
}
```

## 2. Template Rendering

```go
// internal/agent/template.go

type TemplateData struct {
    IssueNumber string
    RepoURL     string
    RepoOwner   string
    RepoName    string
    PRNumber    string
    Steps       map[string]StepResult
}

func RenderPrompts(def config.AgentDef, repoPath string,
    data TemplateData) (system, task string, err error) {
    // System: load from file (relative to repo) or use inline if contains newlines
    if strings.Contains(def.Prompt.System, "\n") {
        system = def.Prompt.System
    } else {
        raw, err := os.ReadFile(filepath.Join(repoPath, def.Prompt.System))
        if err != nil { return "", "", err }
        system = string(raw)
    }
    system += outputContract(def.OutputSchema) // append marker instructions

    // Task: render as Go template
    tmpl, err := template.New("task").Parse(def.Prompt.Task)
    if err != nil { return "", "", fmt.Errorf("parse task template: %w", err) }
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil { return "", "", err }
    return system, buf.String(), nil
}
```

`outputContract` appends a section listing the `###PIPELINE_OUTPUT###` marker
format and required JSON fields from the output schema.

## 3. Init Script

Static bash script, variable substitution via container env vars:

```bash
git config --global user.name "$AGENT_GIT_NAME"
git config --global user.email "$AGENT_GIT_EMAIL"
echo "${AGENT_GH_TOKEN}" | gh auth login --with-token 2>/dev/null
gh auth setup-git 2>/dev/null
if [ -S "${SSH_AUTH_SOCK:-}" ]; then
    signing_pubkey=$(ssh-add -L 2>/dev/null | head -1)
    if [ -n "$signing_pubkey" ]; then
        git config --global gpg.format ssh
        git config --global user.signingkey "key::${signing_pubkey}"
        git config --global commit.gpgsign true
    fi
fi
cd /workspace
exec claude -p --dangerously-skip-permissions --output-format text \
    --system-prompt "$(cat /tmp/orchestrator-prompts/system-prompt.txt)" \
    "$(cat /tmp/orchestrator-prompts/task-prompt.txt)"
```

Passed as the single argument to `ENTRYPOINT ["/bin/bash", "-c"]`.

## 4. Container Runner

```go
// internal/agent/runner.go

type RunConfig struct {
    Image, EnvFilePath, RepoPath string
    SkillsDir, SSHSock, LogDir   string
    GitName, GitEmail            string
}

func RunAgent(ctx context.Context, stepName string, def config.AgentDef,
    data TemplateData, cfg RunConfig, logger *slog.Logger) (*StepResult, error) {
    system, task, err := RenderPrompts(def, cfg.RepoPath, data)
    if err != nil { return nil, err }

    // Write prompts to tmpdir
    promptDir, _ := os.MkdirTemp("", "conductor-prompts-")
    defer os.RemoveAll(promptDir)
    os.WriteFile(filepath.Join(promptDir, "system-prompt.txt"), []byte(system), 0644)
    os.WriteFile(filepath.Join(promptDir, "task-prompt.txt"), []byte(task), 0644)

    // Docker args: --rm, --cap-drop=ALL, --security-opt=no-new-privileges,
    // -u 1000:1000, --env-file, -v workspace, -v prompts:ro, [-v skills:ro],
    // [-v ssh-sock], image, init_script
    args := buildDockerArgs(def, cfg, promptDir)
    output, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()

    // Log to file
    ts := time.Now().Format("20060102-150405")
    logPath := filepath.Join(cfg.LogDir, fmt.Sprintf("conductor-%s-%s.log", stepName, ts))
    os.WriteFile(logPath, output, 0644)

    if err != nil {
        return &StepResult{Name: stepName, Status: "failure",
            Error: err.Error(), LogPath: logPath}, err
    }

    result, err := ParseOutput(stepName, string(output), def.OutputSchema)
    if err != nil { return nil, err }
    result.LogPath = logPath
    return result, nil
}
```

## 5. Output Parser

```go
// internal/agent/parser.go

const pipelineMarker = "###PIPELINE_OUTPUT###"

func ParseOutput(name, output string, schema map[string]any) (*StepResult, error) {
    var payload string
    for _, line := range strings.Split(output, "\n") {
        if idx := strings.Index(line, pipelineMarker); idx >= 0 {
            payload = line[idx+len(pipelineMarker):]
        }
    }
    if payload == "" {
        return nil, fmt.Errorf("no %s marker found in output", pipelineMarker)
    }
    var parsed map[string]any
    if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
        return nil, fmt.Errorf("invalid JSON after marker: %w", err)
    }
    if err := validateSchema(parsed, schema); err != nil {
        return nil, fmt.Errorf("output schema mismatch: %w", err)
    }
    status, _ := parsed["status"].(string)
    r := &StepResult{Name: name, Status: status, Output: parsed}
    if status == "failure" { r.Error, _ = parsed["error"].(string) }
    return r, nil
}

func validateSchema(data, schema map[string]any) error {
    var errs []error
    for key, typ := range schema {
        val, ok := data[key]
        if !ok { errs = append(errs, fmt.Errorf("missing field %q", key)); continue }
        switch typ {
        case "string":
            if _, ok := val.(string); !ok { errs = append(errs, fmt.Errorf("%q: want string", key)) }
        case "int":
            if _, ok := val.(float64); !ok { errs = append(errs, fmt.Errorf("%q: want int", key)) }
        }
    }
    return errors.Join(errs...)
}
```

## 6. Template Data Flow Example

```
Step "review" uses agent "reviewer". TemplateData at render time:

  IssueNumber: "55"
  RepoOwner:   "dmitriyb"
  RepoName:    "differentia"
  PRNumber:    "42"                           ← from steps.implement.Output.pr_number
  Steps:       {"implement": {Status: "success", Output: {pr_number: 42, ...}}}

Template: "Review PR #{{.PRNumber}} in {{.RepoOwner}}/{{.RepoName}}."
Rendered: "Review PR #42 in dmitriyb/differentia."
```

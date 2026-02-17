# Pipeline Module — Implementation

## 1. DAG Types and Construction

```go
// internal/pipeline/dag.go

type Graph struct {
    Nodes map[string]*Node
    Order []string // topological order
}

type Node struct {
    Name       string
    Step       config.StepDef
    Deps       []*Node
    Dependents []*Node
    InDegree   int
}

func BuildGraph(steps []config.StepDef) (*Graph, error) {
    g := &Graph{Nodes: make(map[string]*Node, len(steps))}
    for _, s := range steps {
        g.Nodes[s.Name] = &Node{Name: s.Name, Step: s}
    }
    for _, s := range steps {
        node := g.Nodes[s.Name]
        for _, dep := range s.DependsOn {
            d, ok := g.Nodes[dep]
            if !ok { return nil, fmt.Errorf("step %q: unknown dep %q", s.Name, dep) }
            node.Deps = append(node.Deps, d)
            d.Dependents = append(d.Dependents, node)
            node.InDegree++
        }
    }
    order, err := topoSort(g)
    if err != nil { return nil, err }
    g.Order = order
    return g, nil
}
```

## 2. Topological Sort + Cycle Detection

Kahn's algorithm. Alphabetical tie-breaking for determinism.

```go
func topoSort(g *Graph) ([]string, error) {
    inDeg := make(map[string]int, len(g.Nodes))
    for n, node := range g.Nodes { inDeg[n] = node.InDegree }

    var queue []string
    for n, d := range inDeg { if d == 0 { queue = append(queue, n) } }
    sort.Strings(queue)

    var order []string
    for len(queue) > 0 {
        name := queue[0]; queue = queue[1:]
        order = append(order, name)
        for _, dep := range g.Nodes[name].Dependents {
            inDeg[dep.Name]--
            if inDeg[dep.Name] == 0 { queue = append(queue, dep.Name); sort.Strings(queue) }
        }
    }
    if len(order) != len(g.Nodes) {
        return nil, fmt.Errorf("pipeline has a cycle: %s", strings.Join(findCycle(g), " → "))
    }
    return order, nil
}
```

`findCycle` uses DFS with an on-stack set to find and report the exact cycle path.

## 3. CEL Evaluator

```go
// internal/pipeline/cel.go

func EvalCondition(expr string, results map[string]agent.StepResult) (bool, error) {
    if expr == "" { return true, nil }

    env, _ := cel.NewEnv(cel.Variable("steps", cel.MapType(cel.StringType, cel.DynType)))
    ast, iss := env.Parse(expr)
    if iss.Err() != nil { return false, fmt.Errorf("cel parse: %w", iss.Err()) }
    checked, iss := env.Check(ast)
    if iss.Err() != nil { return false, fmt.Errorf("cel check: %w", iss.Err()) }

    stepsMap := make(map[string]any, len(results))
    for name, r := range results {
        stepsMap[name] = map[string]any{"status": r.Status, "output": r.Output}
    }

    prg, _ := env.Program(checked)
    out, _, err := prg.Eval(map[string]any{"steps": stepsMap})
    if err != nil { return false, fmt.Errorf("cel eval: %w", err) }

    val, ok := out.Value().(bool)
    if !ok { return false, fmt.Errorf("condition must be bool, got %T", out.Value()) }
    return val, nil
}
```

## 4. Parallel Executor

```go
// internal/pipeline/executor.go

func Execute(ctx context.Context, cfg *config.Config,
    runCfg agent.RunConfig, logger *slog.Logger) (*PipelineResult, error) {
    start := time.Now()
    graph, err := BuildGraph(cfg.Pipeline)
    if err != nil { return nil, err }

    var mu sync.Mutex
    results := make(map[string]agent.StepResult)
    inDeg := make(map[string]int, len(graph.Nodes))
    for n, node := range graph.Nodes { inDeg[n] = node.InDegree }

    ready := make(chan *Node, len(graph.Nodes))
    var wg sync.WaitGroup

    // Seed roots
    for _, n := range graph.Order { if inDeg[n] == 0 { ready <- graph.Nodes[n] } }

    // Coordinator
    remaining := len(graph.Nodes)
    go func() {
        for remaining > 0 {
            select {
            case node := <-ready:
                wg.Add(1)
                go func(n *Node) {
                    defer wg.Done()
                    executeStep(ctx, n, cfg, runCfg, logger, &mu, results, inDeg, ready)
                }(node)
            case <-ctx.Done(): return
            }
            remaining--
        }
    }()
    wg.Wait()

    // Aggregate in config order
    pr := &PipelineResult{Duration: time.Since(start)}
    allOK := true
    for _, step := range cfg.Pipeline {
        if r, ok := results[step.Name]; ok {
            pr.Steps = append(pr.Steps, r)
            if r.Status == "failure" { allOK = false }
        }
    }
    pr.Status = map[bool]string{true: "success", false: "failure"}[allOK]
    return pr, nil
}
```

## 5. Step Execution

```go
func executeStep(ctx context.Context, node *Node, cfg *config.Config,
    runCfg agent.RunConfig, logger *slog.Logger,
    mu *sync.Mutex, results map[string]agent.StepResult,
    inDeg map[string]int, ready chan<- *Node) {

    // Snapshot results for CEL
    mu.Lock()
    snap := maps.Clone(results)
    mu.Unlock()

    // Evaluate condition
    if node.Step.Condition != "" {
        pass, err := EvalCondition(node.Step.Condition, snap)
        if err != nil || !pass {
            status := "skipped"
            mu.Lock(); results[node.Name] = agent.StepResult{Name: node.Name, Status: status}; mu.Unlock()
            signalDependents(node, mu, inDeg, ready)
            return
        }
    }

    // Run agent
    data := buildTemplateData(cfg, snap)
    def := cfg.Agents[node.Step.Agent]
    result, err := agent.RunAgent(ctx, node.Name, def, data, runCfg, logger)
    if err != nil {
        if result == nil { result = &agent.StepResult{Name: node.Name, Status: "failure", Error: err.Error()} }
        mu.Lock(); results[node.Name] = *result; mu.Unlock()
        propagateFailure(node, mu, results, inDeg, ready)
        return
    }
    mu.Lock(); results[node.Name] = *result; mu.Unlock()
    signalDependents(node, mu, inDeg, ready)
}
```

## 6. Failure Propagation

BFS from failed node through `Dependents`. Each unreached dependent is
marked `skipped` with reason `"dependency X failed"`. Independent branches
are unaffected — they continue via the normal `ready` channel flow.

## 7. PipelineResult

```go
type PipelineResult struct {
    Steps    []agent.StepResult
    Status   string
    Duration time.Duration
}

func (r *PipelineResult) Print(w io.Writer) {
    fmt.Fprintln(w, "\n============================================")
    fmt.Fprintln(w, "  Pipeline Summary")
    fmt.Fprintln(w, "============================================")
    for _, s := range r.Steps { fmt.Fprintf(w, "  %-20s %s\n", s.Name, s.Status) }
    fmt.Fprintf(w, "\n  Status:   %s\n  Duration: %s\n", r.Status, r.Duration.Round(time.Second))
    fmt.Fprintln(w, "============================================")
}
```

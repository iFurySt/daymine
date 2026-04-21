package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ifuryst/daymine/packages/workspace"
)

type Provider interface {
	Name() string
	Run(ctx context.Context, req Request) (Result, error)
}

type Request struct {
	Query     string
	Workspace string
	Timeout   time.Duration
}

type Result struct {
	ExitCode  int
	Output    string
	Artifacts []string
}

type Controller struct {
	Store     *workspace.Store
	Providers map[string]Provider
}

func NewController(store *workspace.Store, providers ...Provider) *Controller {
	controller := &Controller{Store: store, Providers: map[string]Provider{}}
	for _, provider := range providers {
		controller.Providers[provider.Name()] = provider
	}
	return controller
}

func (c *Controller) Run(ctx context.Context, providerName, query string) (workspace.RunRecord, error) {
	if providerName == "" {
		providerName = "local-command"
	}
	provider, ok := c.Providers[providerName]
	if !ok {
		return workspace.RunRecord{}, fmt.Errorf("provider %q not registered", providerName)
	}
	if strings.TrimSpace(query) == "" {
		return workspace.RunRecord{}, fmt.Errorf("query is required")
	}

	started := time.Now()
	record := workspace.RunRecord{
		ID:        fmt.Sprintf("run-%s", started.UTC().Format("20060102T150405.000000000")),
		Provider:  provider.Name(),
		Query:     query,
		Status:    "running",
		StartedAt: started,
	}
	before, err := c.Store.Snapshot()
	if err != nil {
		return record, err
	}

	result, runErr := provider.Run(ctx, Request{Query: query, Workspace: c.Store.Root, Timeout: 2 * time.Minute})
	record.CompletedAt = time.Now()
	record.ExitCode = result.ExitCode
	record.Output = trimOutput(result.Output, 4000)
	record.Artifacts = result.Artifacts
	if runErr != nil {
		record.Status = "failed"
		record.Error = runErr.Error()
	} else {
		record.Status = "completed"
	}

	if len(record.Artifacts) == 0 {
		if files, err := c.Store.NewFiles(before); err == nil {
			record.Artifacts = files
		}
	}
	if err := c.Store.AppendRun(record); err != nil {
		return record, err
	}
	return record, runErr
}

type LocalCommandProvider struct{}

func (LocalCommandProvider) Name() string {
	return "local-command"
}

func (LocalCommandProvider) Run(ctx context.Context, req Request) (Result, error) {
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "sh", "-lc", req.Query)
	cmd.Dir = req.Workspace
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	result := Result{Output: out.String()}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if runCtx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		return result, runCtx.Err()
	}
	return result, err
}

type CodexCLIProvider struct{}

func (CodexCLIProvider) Name() string {
	return "codex-cli"
}

func (CodexCLIProvider) Run(ctx context.Context, req Request) (Result, error) {
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "codex", "exec", req.Query)
	cmd.Dir = req.Workspace
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	result := Result{Output: out.String()}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if runCtx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		return result, runCtx.Err()
	}
	return result, err
}

func trimOutput(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "\n[output truncated]"
}

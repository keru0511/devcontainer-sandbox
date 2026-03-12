package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/quest-one/quest-one/internal/adapters/events"
	"github.com/quest-one/quest-one/internal/adapters/priority"
	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/mcp"
)

// newTestServer creates a minimal App with nil repositories (MCP initialize/tools/list don't hit DB)
func newTestServer(t *testing.T) (*mcp.Server, *bytes.Buffer) {
	t.Helper()
	app := &application.App{
		Events:   events.NewLogPublisher(slog.Default()),
		Priority: priority.New(),
		Log:      slog.Default(),
	}
	out := &bytes.Buffer{}
	s := mcp.NewWithIO(app, slog.Default(), strings.NewReader(""), out)
	return s, out
}

func TestMCPInitialize(t *testing.T) {
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]any{},
	}
	reqJSON, _ := json.Marshal(req)

	out := &bytes.Buffer{}
	app := &application.App{
		Events:   events.NewLogPublisher(slog.Default()),
		Priority: priority.New(),
		Log:      slog.Default(),
	}
	s := mcp.NewWithIO(app, slog.Default(), bytes.NewReader(append(reqJSON, '\n')), out)
	_ = s.Run(context.Background())

	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("unmarshal response: %v (raw: %s)", err, out.Bytes())
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocolVersion: %v", result["protocolVersion"])
	}
}

func TestMCPToolsList(t *testing.T) {
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	}
	reqJSON, _ := json.Marshal(req)

	out := &bytes.Buffer{}
	app := &application.App{
		Events:   events.NewLogPublisher(slog.Default()),
		Priority: priority.New(),
		Log:      slog.Default(),
	}
	s := mcp.NewWithIO(app, slog.Default(), bytes.NewReader(append(reqJSON, '\n')), out)
	_ = s.Run(context.Background())

	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("unmarshal: %v (raw: %q)", err, out.Bytes())
	}
	result := resp["result"].(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) == 0 {
		t.Error("expected at least one tool")
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "nonexistent/method",
	}
	reqJSON, _ := json.Marshal(req)

	out := &bytes.Buffer{}
	app := &application.App{
		Events:   events.NewLogPublisher(slog.Default()),
		Priority: priority.New(),
		Log:      slog.Default(),
	}
	s := mcp.NewWithIO(app, slog.Default(), bytes.NewReader(append(reqJSON, '\n')), out)
	_ = s.Run(context.Background())

	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["error"] == nil {
		t.Errorf("expected error for unknown method, got: %v", resp)
	}
}

// Ensure newTestServer compiles (used above).
var _ = newTestServer

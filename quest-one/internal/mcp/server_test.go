package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/quest-one/quest-one/internal/adapters/events"
	"github.com/quest-one/quest-one/internal/adapters/priority"
	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/mcp"
)

func runMCPRequest(t *testing.T, req map[string]any) map[string]any {
	t.Helper()
	reqJSON, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	out := &bytes.Buffer{}
	app := &application.App{
		Events:   events.NewLogPublisher(slog.Default()),
		Priority: priority.New(),
		Log:      slog.Default(),
	}
	s := mcp.NewWithIO(app, slog.Default(), bytes.NewReader(append(reqJSON, '\n')), out)
	if err := s.Run(context.Background()); err != nil {
		t.Fatalf("server run: %v", err)
	}
	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("unmarshal response: %v (raw: %s)", err, out.Bytes())
	}
	return resp
}

func TestMCPInitialize(t *testing.T) {
	resp := runMCPRequest(t, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]any{},
	})
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocolVersion: %v", result["protocolVersion"])
	}
}

func TestMCPToolsList(t *testing.T) {
	resp := runMCPRequest(t, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	result := resp["result"].(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) == 0 {
		t.Error("expected at least one tool")
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	resp := runMCPRequest(t, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "nonexistent/method",
	})
	if resp["error"] == nil {
		t.Errorf("expected error for unknown method, got: %v", resp)
	}
}

// Package mcp implements a Model Context Protocol (MCP) server over stdio.
// It speaks JSON-RPC 2.0 and exposes quest-one operations as MCP tools.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/domain"
)

// jsonRPCRequest is a JSON-RPC 2.0 request.
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonRPCResponse is a JSON-RPC 2.0 response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server is the MCP stdio server.
type Server struct {
	app *application.App
	log *slog.Logger
	in  io.Reader
	out io.Writer
}

// New creates an MCP server reading from stdin and writing to stdout.
func New(app *application.App, log *slog.Logger) *Server {
	return &Server{app: app, log: log, in: os.Stdin, out: os.Stdout}
}

// NewWithIO creates an MCP server with custom reader/writer, useful for testing.
func NewWithIO(app *application.App, log *slog.Logger, in io.Reader, out io.Writer) *Server {
	return &Server{app: app, log: log, in: in, out: out}
}

// Run starts the stdio JSON-RPC loop. Blocks until ctx is cancelled or EOF.
func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	enc := json.NewEncoder(s.out)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(errorResponse(nil, -32700, "parse error"))
			continue
		}

		result, rpcErr := s.dispatch(ctx, req.Method, req.Params)
		if rpcErr != nil {
			_ = enc.Encode(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   rpcErr,
			})
			continue
		}

		resultJSON, _ := json.Marshal(result)
		_ = enc.Encode(jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  resultJSON,
		})
	}

	return scanner.Err()
}

func (s *Server) dispatch(ctx context.Context, method string, params json.RawMessage) (any, *jsonRPCError) {
	switch method {
	case "initialize":
		return s.handleInitialize()
	case "tools/list":
		return s.handleToolsList()
	case "tools/call":
		return s.handleToolsCall(ctx, params)
	default:
		return nil, &jsonRPCError{Code: -32601, Message: fmt.Sprintf("method not found: %s", method)}
	}
}

func (s *Server) handleInitialize() (any, *jsonRPCError) {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo": map[string]any{
			"name":    "quest-one",
			"version": "0.1.0",
		},
	}, nil
}

func (s *Server) handleToolsList() (any, *jsonRPCError) {
	tools := []map[string]any{
		{
			"name":        "next_task",
			"description": "Get the highest-priority active task",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			"name":        "add_task",
			"description": "Add a new task",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":       map[string]string{"type": "string", "description": "Task title"},
					"description": map[string]string{"type": "string", "description": "Optional description"},
				},
				"required": []string{"title"},
			},
		},
		{
			"name":        "list_tasks",
			"description": "List active tasks ordered by priority",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit": map[string]string{"type": "integer", "description": "Max results (default 20)"},
				},
			},
		},
		{
			"name":        "complete_task",
			"description": "Mark a task as completed",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{"id": map[string]string{"type": "string"}},
				"required":   []string{"id"},
			},
		},
		{
			"name":        "candidates",
			"description": "Get the top N priority tasks",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"n": map[string]string{"type": "integer", "description": "Number of candidates"},
				},
			},
		},
		{
			"name":        "search_tasks",
			"description": "Full-text search over tasks",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{"query": map[string]string{"type": "string"}},
				"required":   []string{"query"},
			},
		},
	}
	return map[string]any{"tools": tools}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (any, *jsonRPCError) {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &jsonRPCError{Code: -32602, Message: "invalid params"}
	}

	var result any
	var err error

	switch p.Name {
	case "next_task":
		result, err = s.app.NextTask(ctx)
	case "add_task":
		title, _ := p.Arguments["title"].(string)
		desc, _ := p.Arguments["description"].(string)
		result, err = s.app.AddTask(ctx, application.AddTaskInput{
			Title:       title,
			Description: desc,
		})
	case "list_tasks":
		n := 20
		if v, ok := p.Arguments["limit"].(float64); ok {
			n = int(v)
		}
		result, err = s.app.ListTasks(ctx, application.ListTasksInput{Limit: n})
	case "complete_task":
		id, _ := p.Arguments["id"].(string)
		result, err = s.app.CompleteTask(ctx, domain.TaskID(id))
	case "candidates":
		n := 5
		if v, ok := p.Arguments["n"].(float64); ok {
			n = int(v)
		}
		result, err = s.app.Candidates(ctx, n)
	case "search_tasks":
		q, _ := p.Arguments["query"].(string)
		result, err = s.app.SearchTasks(ctx, q, 20)
	default:
		return nil, &jsonRPCError{Code: -32601, Message: "unknown tool: " + p.Name}
	}

	if err != nil {
		return nil, &jsonRPCError{Code: -32000, Message: err.Error()}
	}

	b, _ := json.Marshal(result)
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(b)},
		},
	}, nil
}

func errorResponse(id any, code int, msg string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: msg},
	}
}

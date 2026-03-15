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

// JSON-RPC method names.
const (
	methodInitialize = "initialize"
	methodToolsList  = "tools/list"
	methodToolsCall  = "tools/call"
)

// MCP tool names.
const (
	toolNextTask     = "next_task"
	toolAddTask      = "add_task"
	toolListTasks    = "list_tasks"
	toolCompleteTask = "complete_task"
	toolCandidates   = "candidates"
	toolSearchTasks  = "search_tasks"
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
	app       *application.App
	log       *slog.Logger
	in        io.Reader
	out       io.Writer
	initResp  json.RawMessage // pre-marshalled initialize response
	toolsResp json.RawMessage // pre-marshalled tools/list response
}

// New creates an MCP server reading from stdin and writing to stdout.
func New(app *application.App, log *slog.Logger) *Server {
	return NewWithIO(app, log, os.Stdin, os.Stdout)
}

// NewWithIO creates an MCP server with custom reader/writer, useful for testing.
func NewWithIO(app *application.App, log *slog.Logger, in io.Reader, out io.Writer) *Server {
	s := &Server{app: app, log: log, in: in, out: out}
	s.initResp, _ = json.Marshal(map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": "quest-one", "version": "0.1.0"},
	})
	s.toolsResp, _ = json.Marshal(buildToolsList())
	return s
}

// scannerBufSize is the maximum line size the MCP scanner will accept.
// 1 MiB is sufficient for large tool responses (e.g. full task lists).
const scannerBufSize = 1 << 20

// scannerInitBufSize is the initial scanner buffer; grows on demand up to scannerBufSize.
const scannerInitBufSize = 4096

// Run starts the stdio JSON-RPC loop. Blocks until ctx is cancelled or EOF.
func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	scanner.Buffer(make([]byte, scannerInitBufSize), scannerBufSize)
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
			if err := enc.Encode(errorResponse(nil, -32700, "parse error")); err != nil {
				return fmt.Errorf("mcp: write: %w", err)
			}
			continue
		}

		result, rpcErr := s.dispatch(ctx, req.Method, req.Params)
		if rpcErr != nil {
			if err := enc.Encode(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   rpcErr,
			}); err != nil {
				return fmt.Errorf("mcp: write: %w", err)
			}
			continue
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			if err := enc.Encode(errorResponse(req.ID, -32603, "internal error: marshal result")); err != nil {
				return fmt.Errorf("mcp: write: %w", err)
			}
			continue
		}
		if err = enc.Encode(jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  resultJSON,
		}); err != nil {
			return fmt.Errorf("mcp: write: %w", err)
		}
	}

	return scanner.Err()
}

func (s *Server) dispatch(ctx context.Context, method string, params json.RawMessage) (any, *jsonRPCError) {
	switch method {
	case methodInitialize:
		return s.handleInitialize()
	case methodToolsList:
		return s.handleToolsList()
	case methodToolsCall:
		return s.handleToolsCall(ctx, params)
	default:
		return nil, &jsonRPCError{Code: -32601, Message: fmt.Sprintf("method not found: %s", method)}
	}
}

func (s *Server) handleInitialize() (any, *jsonRPCError) {
	return json.RawMessage(s.initResp), nil
}

func (s *Server) handleToolsList() (any, *jsonRPCError) {
	return json.RawMessage(s.toolsResp), nil
}

func buildToolsList() map[string]any {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        toolNextTask,
				"description": "Get the highest-priority active task",
				"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
			},
			{
				"name":        toolAddTask,
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
				"name":        toolListTasks,
				"description": "List active tasks ordered by priority",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"limit": map[string]string{"type": "integer", "description": "Max results (default 20)"},
					},
				},
			},
			{
				"name":        toolCompleteTask,
				"description": "Mark a task as completed",
				"inputSchema": map[string]any{
					"type":       "object",
					"properties": map[string]any{"id": map[string]string{"type": "string"}},
					"required":   []string{"id"},
				},
			},
			{
				"name":        toolCandidates,
				"description": "Get the top N priority tasks",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"n": map[string]string{"type": "integer", "description": "Number of candidates"},
					},
				},
			},
			{
				"name":        toolSearchTasks,
				"description": "Full-text search over tasks",
				"inputSchema": map[string]any{
					"type":       "object",
					"properties": map[string]any{"query": map[string]string{"type": "string"}},
					"required":   []string{"query"},
				},
			},
		},
	}
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
	case toolNextTask:
		result, err = s.app.NextTask(ctx)
	case toolAddTask:
		title, ok := p.Arguments["title"].(string)
		if !ok || title == "" {
			return nil, &jsonRPCError{Code: -32602, Message: "add_task: missing required argument: title"}
		}
		desc, _ := p.Arguments["description"].(string)
		result, err = s.app.AddTask(ctx, application.AddTaskInput{
			Title:       title,
			Description: desc,
		})
	case toolListTasks:
		n := 20
		if v, ok := p.Arguments["limit"].(float64); ok {
			n = int(v)
		}
		result, err = s.app.ListTasks(ctx, application.ListTasksInput{Limit: n})
	case toolCompleteTask:
		id, ok := p.Arguments["id"].(string)
		if !ok || id == "" {
			return nil, &jsonRPCError{Code: -32602, Message: "complete_task: missing required argument: id"}
		}
		result, err = s.app.CompleteTask(ctx, domain.TaskID(id))
	case toolCandidates:
		n := 5
		if v, ok := p.Arguments["n"].(float64); ok {
			n = int(v)
		}
		result, err = s.app.Candidates(ctx, n)
	case toolSearchTasks:
		q, ok := p.Arguments["query"].(string)
		if !ok || q == "" {
			return nil, &jsonRPCError{Code: -32602, Message: "search_tasks: missing required argument: query"}
		}
		result, err = s.app.SearchTasks(ctx, q, 20)
	default:
		return nil, &jsonRPCError{Code: -32601, Message: "unknown tool: " + p.Name}
	}

	if err != nil {
		return nil, &jsonRPCError{Code: -32000, Message: err.Error()}
	}

	b, err := json.Marshal(result)
	if err != nil {
		return nil, &jsonRPCError{Code: -32603, Message: "internal error: marshal tool result"}
	}
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

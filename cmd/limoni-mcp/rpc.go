package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// The Model Context Protocol over stdio is JSON-RPC 2.0, one message per line.
// Only the part a tool server needs is implemented here — initialize, ping,
// tools/list, tools/call and cancellation — which is small enough that pulling
// an SDK into Limoni's dependency graph would cost more than it saves.

// supportedVersions lists the protocol revisions this server speaks, newest
// first. Everything it uses — tools with text content and annotations — reads
// the same in all of them.
var supportedVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26", "2024-11-05"}

const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
)

// maxMessageBytes bounds one incoming line. Tool arguments are a selector or a
// few words of text; a megabyte is far beyond any honest request.
const maxMessageBytes = 1 << 20

type incoming struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type outgoing struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return e.Message }

// toolResult is the result of tools/call. A tool that fails reports the
// failure here with IsError set, not as a JSON-RPC error: the model is meant
// to read it and correct course, for instance by fixing an ambiguous selector.
type toolResult struct {
	Content []textContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// mcpServer dispatches MCP requests to a tool set.
type mcpServer struct {
	name, version string
	instructions  string
	tools         []tool

	writeMu sync.Mutex
	out     *json.Encoder

	mu       sync.Mutex
	inFlight map[string]context.CancelFunc
}

func newMCPServer(name, version, instructions string, tools []tool) *mcpServer {
	return &mcpServer{
		name:         name,
		version:      version,
		instructions: instructions,
		tools:        tools,
		inFlight:     make(map[string]context.CancelFunc),
	}
}

// serve reads requests from r until it is exhausted or ctx ends, answering on
// w. Requests run concurrently, so a long wait_for does not hold up a ping or
// the cancellation that ends it.
func (s *mcpServer) serve(ctx context.Context, r io.Reader, w io.Writer) error {
	// Deferred in this order so that, once input ends, calls still in flight
	// are cancelled first and only then waited for.
	var wg sync.WaitGroup
	defer wg.Wait()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.out = json.NewEncoder(w)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxMessageBytes)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if line[0] == '[' {
			// JSON-RPC batches were removed from MCP in 2025-06-18.
			s.write(outgoing{ID: json.RawMessage("null"), Error: &rpcError{codeInvalidRequest, "batched requests are not supported"}})
			continue
		}
		var msg incoming
		if err := json.Unmarshal(line, &msg); err != nil {
			s.write(outgoing{ID: json.RawMessage("null"), Error: &rpcError{codeParseError, "parse error: " + err.Error()}})
			continue
		}
		if msg.Method == "" {
			continue // A response to a request this server never sends.
		}
		if len(msg.ID) == 0 {
			s.notification(msg)
			continue
		}

		id := string(msg.ID)
		reqCtx, reqCancel := context.WithCancel(ctx)
		s.mu.Lock()
		s.inFlight[id] = reqCancel
		s.mu.Unlock()

		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := s.request(reqCtx, msg)

			s.mu.Lock()
			_, stillWanted := s.inFlight[id]
			delete(s.inFlight, id)
			s.mu.Unlock()
			reqCancel()
			// A cancelled request gets no response; the client has already
			// stopped waiting for it.
			if !stillWanted {
				return
			}

			resp := outgoing{ID: msg.ID, Result: result}
			if err != nil {
				resp.Result = nil
				resp.Error = err
			}
			s.write(resp)
		}()
	}
	return scanner.Err()
}

func (s *mcpServer) write(msg outgoing) {
	msg.JSONRPC = "2.0"
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	// json.Encoder writes one line per value and escapes newlines inside
	// strings, which is exactly the framing stdio requires.
	_ = s.out.Encode(msg)
}

func (s *mcpServer) notification(msg incoming) {
	if msg.Method != "notifications/cancelled" {
		return // notifications/initialized and anything newer need no action.
	}
	var params struct {
		RequestID json.RawMessage `json:"requestId"`
	}
	if json.Unmarshal(msg.Params, &params) != nil || len(params.RequestID) == 0 {
		return
	}
	s.mu.Lock()
	cancel, ok := s.inFlight[string(params.RequestID)]
	delete(s.inFlight, string(params.RequestID))
	s.mu.Unlock()
	if ok {
		cancel()
	}
}

func (s *mcpServer) request(ctx context.Context, msg incoming) (any, *rpcError) {
	switch msg.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(msg.Params, &params)
		return map[string]any{
			"protocolVersion": negotiateVersion(params.ProtocolVersion),
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": s.name, "version": s.version},
			"instructions":    s.instructions,
		}, nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		list := make([]map[string]any, 0, len(s.tools))
		for _, t := range s.tools {
			list = append(list, t.descriptor())
		}
		return map[string]any{"tools": list}, nil

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return nil, &rpcError{codeInvalidParams, "invalid tools/call params: " + err.Error()}
		}
		for _, t := range s.tools {
			if t.name != params.Name {
				continue
			}
			args := params.Arguments
			if len(args) == 0 || string(args) == "null" {
				args = json.RawMessage("{}")
			}
			text, err := t.run(ctx, args)
			if err != nil {
				return toolResult{Content: []textContent{{Type: "text", Text: err.Error()}}, IsError: true}, nil
			}
			return toolResult{Content: []textContent{{Type: "text", Text: text}}}, nil
		}
		return nil, &rpcError{codeInvalidParams, fmt.Sprintf("unknown tool %q", params.Name)}

	default:
		return nil, &rpcError{codeMethodNotFound, fmt.Sprintf("method %q not found", msg.Method)}
	}
}

// negotiateVersion answers with the client's revision when this server speaks
// it, and otherwise with the newest one it does, leaving the client to decide
// whether it can continue.
func negotiateVersion(requested string) string {
	for _, v := range supportedVersions {
		if v == requested {
			return v
		}
	}
	return supportedVersions[0]
}

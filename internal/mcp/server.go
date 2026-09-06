package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Server handles MCP JSON-RPC 2.0 requests over an io.Reader and io.Writer.
type Server struct {
	Version string
}

// NewServer creates a new MCP server instance.
func NewServer(version string) *Server {
	return &Server{Version: version}
}

// Serve reads JSON-RPC requests from r and writes responses to w until EOF.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	// Allow large requests if necessary
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			resp := JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			}
			_ = sendResponse(w, resp)
			continue
		}

		// Handle notifications (no ID)
		if req.ID == nil {
			continue
		}

		resp := s.handleRequest(&req)
		if err := sendResponse(w, resp); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req *JSONRPCRequest) JSONRPCResponse {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "arch-vet",
				"version": s.Version,
			},
		}

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": ListTools(),
		}

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
			}
			return resp
		}

		var result *CallToolResult
		var err error

		switch params.Name {
		case "verify_architecture":
			result, err = HandleVerifyArchitecture(params.Arguments)
		case "check_diff":
			result, err = HandleCheckDiff(params.Arguments)
		case "get_architectural_map":
			result, err = HandleGetArchitecturalMap(params.Arguments)
		case "explain_rule":
			result, err = HandleExplainRule(params.Arguments)
		default:
			resp.Error = &JSONRPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method '%s' not found", params.Name),
			}
			return resp
		}

		if err != nil {
			resp.Error = &JSONRPCError{
				Code:    -32000,
				Message: err.Error(),
			}
			return resp
		}
		resp.Result = result

	default:
		resp.Error = &JSONRPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method '%s' not found", req.Method),
		}
	}

	return resp
}

func sendResponse(w io.Writer, resp JSONRPCResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

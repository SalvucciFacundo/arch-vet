package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPServer_InitializeAndToolsList(t *testing.T) {
	server := NewServer("0.1.0-test")

	// 1. Send initialize request
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	// 2. Send tools/list request
	toolsReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"
	// 3. Send tools/call for explain_rule
	callReq := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"explain_rule","arguments":{"rule_id":"ARCH001"}}}` + "\n"

	input := initReq + toolsReq + callReq
	var output bytes.Buffer

	err := server.Serve(strings.NewReader(input), &output)
	if err != nil {
		t.Fatalf("server.Serve returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 responses, got %d: %s", len(lines), output.String())
	}

	// Verify initialize response
	var initResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[0]), &initResp); err != nil {
		t.Fatalf("failed to unmarshal init response: %v", err)
	}
	if initResp.Error != nil {
		t.Fatalf("init returned error: %v", initResp.Error)
	}

	// Verify tools/list response
	var toolsResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[1]), &toolsResp); err != nil {
		t.Fatalf("failed to unmarshal tools response: %v", err)
	}
	resMap, ok := toolsResp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map in tools response")
	}
	tools, ok := resMap["tools"].([]interface{})
	if !ok || len(tools) != 3 {
		t.Fatalf("expected 3 tools in list, got %v", tools)
	}

	// Verify tools/call response
	var callResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[2]), &callResp); err != nil {
		t.Fatalf("failed to unmarshal call response: %v", err)
	}
	callResMap, ok := callResp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map in call response")
	}
	contentList, ok := callResMap["content"].([]interface{})
	if !ok || len(contentList) == 0 {
		t.Fatalf("expected content in call response")
	}
	item := contentList[0].(map[string]interface{})
	text := item["text"].(string)
	if !strings.Contains(text, "ARCH001: forbidden-import") {
		t.Errorf("expected text to explain ARCH001, got: %s", text)
	}
}

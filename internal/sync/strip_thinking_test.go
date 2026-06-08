package sync

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestStripThinkingFromSessionJSONStripsAssistantThinkingBlocks(t *testing.T) {
	body := []byte(`{
		"id":"s1",
		"messages":[
			{"role":"assistant","content":[
				{"type":"thinking","thinking":"secret"},
				{"type":"text","text":"visible"},
				{"type":"redacted_thinking","data":"hidden"}
			]},
			{"role":"user","content":[{"type":"thinking","text":"literal user content"}]}
		]
	}`)

	got, err := StripThinkingFromSessionJSON(body)
	if err != nil {
		t.Fatalf("StripThinkingFromSessionJSON: %v", err)
	}

	var decoded struct {
		Messages []struct {
			Role    string           `json:"role"`
			Content []map[string]any `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("unmarshal stripped body: %v", err)
	}
	assistantTypes := blockTypes(decoded.Messages[0].Content)
	if !reflect.DeepEqual(assistantTypes, []string{"text"}) {
		t.Fatalf("assistant types = %v, want [text]", assistantTypes)
	}
	userTypes := blockTypes(decoded.Messages[1].Content)
	if !reflect.DeepEqual(userTypes, []string{"thinking"}) {
		t.Fatalf("user types = %v, want [thinking]", userTypes)
	}
}

func TestStripThinkingFromSessionJSONLeavesStringContentUnchanged(t *testing.T) {
	body := []byte(`{"messages":[{"role":"assistant","content":"thinking is text here"}]}`)

	got, err := StripThinkingFromSessionJSON(body)
	if err != nil {
		t.Fatalf("StripThinkingFromSessionJSON: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("body changed:\ngot  %s\nwant %s", got, body)
	}
}

func TestStripThinkingFromSessionJSONInvalidJSONReturnsOriginalAndError(t *testing.T) {
	body := []byte(`{"messages":[`)

	got, err := StripThinkingFromSessionJSON(body)
	if err == nil {
		t.Fatal("StripThinkingFromSessionJSON should return parse error")
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("body changed on parse error:\ngot  %s\nwant %s", got, body)
	}
}

func blockTypes(blocks []map[string]any) []string {
	types := make([]string, 0, len(blocks))
	for _, block := range blocks {
		t, _ := block["type"].(string)
		types = append(types, t)
	}
	return types
}

package sync

import "encoding/json"

// StripThinkingFromSessionJSON removes assistant thinking/redacted_thinking
// content blocks from a marshaled session JSON body.
func StripThinkingFromSessionJSON(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}
	var top map[string]any
	if err := json.Unmarshal(body, &top); err != nil {
		return body, err
	}
	messages, ok := top["messages"].([]any)
	if !ok {
		return body, nil
	}

	mutated := false
	for _, rawMessage := range messages {
		message, ok := rawMessage.(map[string]any)
		if !ok {
			continue
		}
		role, _ := message["role"].(string)
		if role != "assistant" {
			continue
		}
		content, ok := message["content"].([]any)
		if !ok {
			continue
		}
		filtered := make([]any, 0, len(content))
		dropped := false
		for _, rawBlock := range content {
			block, ok := rawBlock.(map[string]any)
			if !ok {
				filtered = append(filtered, rawBlock)
				continue
			}
			blockType, _ := block["type"].(string)
			if blockType == "thinking" || blockType == "redacted_thinking" {
				dropped = true
				continue
			}
			filtered = append(filtered, rawBlock)
		}
		if dropped {
			message["content"] = filtered
			mutated = true
		}
	}
	if !mutated {
		return body, nil
	}
	return json.Marshal(top)
}

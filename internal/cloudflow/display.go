package cloudflow

func CloudStatusLine(agentID, status, message string) string {
	msg := message
	if msg == "" {
		msg = cloudStatusFallback(status)
	}
	if agentID != "" && agentID != "orchestrator" && agentID != "streaming" {
		return "[" + agentID + "] " + msg
	}
	return msg
}

func cloudStatusFallback(status string) string {
	switch status {
	case "started":
		return "Agent working..."
	case "completed":
		return "Agent completed"
	case "thinking":
		return "Thinking..."
	case "tool":
		return "Calling tool..."
	case "processing":
		return "Processing data..."
	default:
		return "Working..."
	}
}

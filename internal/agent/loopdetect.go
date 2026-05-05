package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// LoopAction tells the agent loop how to respond to a detection signal.
type LoopAction int

const (
	LoopContinue  LoopAction = iota // proceed normally
	LoopNudge                        // inject "try different approach" message
	LoopForceStop                    // force final response without tools
)

// ToolCallRecord tracks a single tool invocation for loop detection.
type ToolCallRecord struct {
	Name      string
	ArgsHash  string // hex-encoded hash of raw args
	TopicHash string // hex-encoded hash of normalized args (web tools)
	ResultSig string // domain signature from results (web tools)
	IsError   bool
	ErrorSig  string // first 100 chars of error for grouping
	IsSleep   bool   // bash command contains sleep
}

// toolFamilies maps tool names to their logical family for grouping.
var toolFamilies = map[string]string{
	"web_search":    "web",
	"web_fetch":     "web",
	"browser":       "web",
	"accessibility": "gui",
	"screenshot":    "gui",
	"computer":      "gui",
	"applescript":   "gui",
	"grep":          "search",
	"glob":          "search",
}

// LoopDetector uses a sliding window of recent tool calls to detect stuck loops.
type LoopDetector struct {
	history     []ToolCallRecord
	historySize int

	consecDupThreshold   int
	exactDupThreshold    int
	sameToolErrThreshold int
	noProgressThreshold  int

	repeatableTools map[string]bool

	lastNonGUISuccess bool
	lastNonGUITool    string
	modeSwitchNudged  bool

	recentRecovery bool
	recoveredTool  string
}

// visualTools are tools used purely for visual verification.
var visualTools = map[string]bool{
	"screenshot": true, "computer": true, "accessibility": true,
}

// repeatableGUITools are tools expected to be called many times in GUI workflows.
var repeatableGUITools = map[string]bool{
	"screenshot": true, "computer": true, "accessibility": true, "browser": true,
}

// NewLoopDetector creates a detector with production defaults.
func NewLoopDetector() *LoopDetector {
	return &LoopDetector{
		history:              make([]ToolCallRecord, 0, 20),
		historySize:          20,
		consecDupThreshold:   2,
		exactDupThreshold:    3,
		sameToolErrThreshold: 4,
		noProgressThreshold:  8,
		repeatableTools:      repeatableGUITools,
	}
}

// Record adds a tool call to the sliding window.
func (ld *LoopDetector) Record(name, argsJSON string, isError bool, errMsg string, resultSig string) {
	topicHash := ""
	if toolFamilies[name] != "" {
		normalized := normalizeWebQuery(argsJSON)
		if normalized != "" {
			topicHash = hashArgs(normalized)
		}
	}
	rec := ToolCallRecord{
		Name:      name,
		ArgsHash:  hashArgs(argsJSON),
		TopicHash: topicHash,
		ResultSig: resultSig,
		IsError:   isError,
		ErrorSig:  truncateErrSig(errMsg, 100),
		IsSleep:   name == "bash" && isSleepCommand(argsJSON),
	}
	ld.history = append(ld.history, rec)
	if len(ld.history) > ld.historySize {
		ld.history = ld.history[len(ld.history)-ld.historySize:]
	}

	if !visualTools[name] {
		if isError {
			ld.lastNonGUISuccess = false
			ld.lastNonGUITool = ""
		} else {
			ld.lastNonGUISuccess = true
			ld.lastNonGUITool = name
			ld.modeSwitchNudged = false
		}
	}

	if !isError && !visualTools[name] {
		hasEarlierError := false
		for _, rec := range ld.history[:len(ld.history)-1] {
			if rec.Name == name && rec.IsError && rec.ArgsHash != ld.history[len(ld.history)-1].ArgsHash {
				hasEarlierError = true
				break
			}
		}
		if hasEarlierError {
			ld.recentRecovery = true
			ld.recoveredTool = name
		} else if name != ld.recoveredTool {
			ld.recentRecovery = false
			ld.recoveredTool = ""
		}
	}
}

// Check evaluates all detectors for the named tool.
func (ld *LoopDetector) Check(name string) (LoopAction, string) {
	if len(ld.history) < 2 {
		return LoopContinue, ""
	}

	// Mode switch: visual tool after successful non-visual tool
	if visualTools[name] && ld.lastNonGUISuccess && !ld.modeSwitchNudged && repeatableGUITools[ld.lastNonGUITool] {
		ld.modeSwitchNudged = true
		return LoopNudge, fmt.Sprintf(
			"Your previous tool call (%s) returned a success result. Visual verification is likely unnecessary.", ld.lastNonGUITool)
	}

	// Success after error: agent recovered but continues verifying
	if visualTools[name] && ld.recentRecovery {
		ld.recentRecovery = false
		return LoopNudge, fmt.Sprintf(
			"You recovered from the earlier %s error. The successful result is your confirmation — proceed to your final answer.", ld.recoveredTool)
	}

	var latestHash string
	for i := len(ld.history) - 1; i >= 0; i-- {
		if ld.history[i].Name == name {
			latestHash = ld.history[i].ArgsHash
			break
		}
	}

	// Consecutive exact duplicate
	consecCount := 0
	for i := len(ld.history) - 1; i >= 0; i-- {
		if ld.history[i].Name != name || ld.history[i].ArgsHash != latestHash {
			break
		}
		consecCount++
	}
	if consecCount >= ld.consecDupThreshold+1 {
		return LoopForceStop, fmt.Sprintf(
			"You have called %s with identical arguments %d times in a row. Stop retrying and provide your answer now.", name, consecCount)
	}
	if consecCount >= ld.consecDupThreshold {
		return LoopNudge, fmt.Sprintf(
			"You've called %s %d times consecutively with identical arguments. The results won't change.", name, consecCount)
	}

	// Window-based exact duplicate
	dupCount := 0
	if latestHash != "" {
		for _, rec := range ld.history {
			if rec.Name == name && rec.ArgsHash == latestHash {
				dupCount++
			}
		}
	}
	if dupCount >= ld.exactDupThreshold*2 {
		return LoopForceStop, fmt.Sprintf(
			"You have called %s with identical arguments %d times. Stop retrying and provide your answer now.", name, dupCount)
	}
	if dupCount >= ld.exactDupThreshold {
		return LoopNudge, fmt.Sprintf(
			"You've called %s %d times with identical arguments. Try a fundamentally different approach.", name, dupCount)
	}

	// Same tool error detector
	errCount := 0
	var lastErr string
	for _, rec := range ld.history {
		if rec.Name == name && rec.IsError {
			errCount++
			lastErr = rec.ErrorSig
		}
	}
	if errCount >= ld.sameToolErrThreshold*2 {
		return LoopForceStop, fmt.Sprintf(
			"Tool %s has failed %d times. Stop using it and provide your answer now.", name, errCount)
	}
	if errCount >= ld.sameToolErrThreshold {
		return LoopNudge, fmt.Sprintf(
			"Tool %s has failed %d times with: %s. Use a different approach.", name, errCount, lastErr)
	}

	// Family no-progress
	family := toolFamilies[name]
	if family != "" {
		latestTopic := ""
		latestResult := ""
		for i := len(ld.history) - 1; i >= 0; i-- {
			if toolFamilies[ld.history[i].Name] == family {
				if latestTopic == "" && ld.history[i].TopicHash != "" {
					latestTopic = ld.history[i].TopicHash
				}
				if latestResult == "" && ld.history[i].ResultSig != "" {
					latestResult = ld.history[i].ResultSig
				}
				if latestTopic != "" && latestResult != "" {
					break
				}
			}
		}

		familyCount := 0
		sameTopicCount := 0
		sameResultCount := 0
		for _, rec := range ld.history {
			if toolFamilies[rec.Name] != family {
				continue
			}
			familyCount++
			if latestTopic != "" && rec.TopicHash == latestTopic {
				sameTopicCount++
			}
			if latestResult != "" && rec.ResultSig == latestResult {
				sameResultCount++
			}
		}

		progressCount := sameTopicCount
		if sameResultCount > progressCount {
			progressCount = sameResultCount
		}

		if progressCount >= 7 {
			return LoopForceStop, fmt.Sprintf(
				"You have made %d calls with %d on the same topic. Return your collected results now.", familyCount, progressCount)
		}
		if progressCount >= 5 {
			return LoopNudge, fmt.Sprintf(
				"You've searched the same topic %d times. Summarize and present to the user.", progressCount)
		}
		if progressCount >= 3 {
			return LoopNudge, fmt.Sprintf(
				"You've searched the same topic %d times with similar results. Try a fundamentally different query.", progressCount)
		}

		if progressCount == 0 && !ld.repeatableTools[name] {
			sameToolInFamily := 0
			for _, rec := range ld.history {
				if rec.Name == name {
					sameToolInFamily++
				}
			}
			if sameToolInFamily >= 7 {
				return LoopForceStop, fmt.Sprintf(
					"You have called %s %d times without meaningful progress. Provide your answer now.", name, sameToolInFamily)
			}
			if sameToolInFamily >= 5 {
				return LoopNudge, fmt.Sprintf(
					"You've called %s %d times. Consider whether you're making progress or stuck in a loop.", name, sameToolInFamily)
			}
		}
	}

	// Search escalation
	if family == "search" {
		consecSearch := 0
		for i := len(ld.history) - 1; i >= 0; i-- {
			if toolFamilies[ld.history[i].Name] == "search" {
				consecSearch++
			} else {
				break
			}
		}
		if consecSearch >= 5 {
			return LoopForceStop, fmt.Sprintf(
				"You have made %d consecutive search calls without acting on results. Stop searching and use what you have.", consecSearch)
		}
		if consecSearch >= 3 {
			return LoopNudge, fmt.Sprintf(
				"You've made %d search calls without finding useful results. Reconsider your approach.", consecSearch)
		}
	}

	// No progress
	if !ld.repeatableTools[name] {
		count := 0
		for _, rec := range ld.history {
			if rec.Name == name {
				count++
			}
		}
		if count >= ld.noProgressThreshold*2 {
			return LoopForceStop, fmt.Sprintf(
				"You have called %s %d times without meaningful progress. Provide your answer now.", name, count)
		}
		if count >= ld.noProgressThreshold {
			return LoopNudge, fmt.Sprintf(
				"You've called %s %d times. Summarize what you've learned and try a different approach.", name, count)
		}
	}

	// Sleep detector
	sleepCount := 0
	for _, rec := range ld.history {
		if rec.IsSleep {
			sleepCount++
		}
	}
	if sleepCount >= 4 {
		return LoopForceStop, fmt.Sprintf(
			"You have used `sleep` in bash commands %d times. Stop polling and provide your answer now.", sleepCount)
	}
	if sleepCount >= 2 {
		return LoopNudge, fmt.Sprintf(
			"You've used `sleep` in bash commands %d times. Do not poll or wait in loops.", sleepCount)
	}

	return LoopContinue, ""
}

func hashArgs(args string) string {
	h := sha256.Sum256([]byte(args))
	return hex.EncodeToString(h[:8])
}

var sleepPattern = regexp.MustCompile(`\bsleep\s+\d`)

func isSleepCommand(argsJSON string) bool {
	var args struct {
		Command string `json:"command"`
	}
	if json.Unmarshal([]byte(argsJSON), &args) != nil {
		return false
	}
	return sleepPattern.MatchString(args.Command)
}

func truncateErrSig(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen])
}

// normalizeWebQuery extracts a search query from JSON args and returns a canonical form.
func normalizeWebQuery(argsJSON string) string {
	var raw map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &raw); err != nil {
		return ""
	}

	query := ""
	for _, key := range []string{"query", "q", "queries", "url", "urls"} {
		if v, ok := raw[key]; ok {
			switch val := v.(type) {
			case string:
				query = val
			case []any:
				if len(val) > 0 {
					if s, ok := val[0].(string); ok {
						query = s
					}
				}
			}
			if query != "" {
				break
			}
		}
	}

	if query == "" {
		// Also check for "pattern" key (grep/glob)
		if v, ok := raw["pattern"]; ok {
			if s, ok := v.(string); ok {
				query = s
			}
		}
	}

	if query == "" {
		return ""
	}

	// Normalize: lowercase, strip URLs, strip dates, strip filler words
	query = strings.ToLower(query)

	// Strip URLs
	urlPat := regexp.MustCompile(`https?://[^\s]+`)
	query = urlPat.ReplaceAllString(query, "")

	// Strip dates
	datePat := regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)
	query = datePat.ReplaceAllString(query, "")

	// Tokenize and filter
	tokens := strings.Fields(query)
	fillerWords := map[string]bool{
		"today": true, "yesterday": true, "latest": true, "recent": true,
		"top": true, "major": true, "breaking": true, "headlines": true,
		"news": true, "current": true, "update": true, "updates": true,
	}
	var filtered []string
	for _, t := range tokens {
		if !fillerWords[t] && len(t) > 1 {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	return strings.Join(filtered, " ")
}

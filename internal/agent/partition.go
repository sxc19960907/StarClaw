package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const maxToolConcurrency = 10

// approvedToolCall represents a tool call that has been approved for execution.
type approvedToolCall struct {
	index   int
	tool    Tool
	argsStr string
}

// toolExecResult holds the result of a single tool execution.
type toolExecResult struct {
	result  ToolResult
	elapsed time.Duration
	err     error
}

// isReadOnly checks if a tool call is read-only by testing the ReadOnlyChecker interface.
func isReadOnlyTool(ac approvedToolCall) bool {
	checker, ok := ac.tool.(ReadOnlyChecker)
	if !ok {
		return false
	}
	return checker.IsReadOnlyCall(ac.argsStr)
}

// PartitionToolCalls groups approved tool calls into execution batches.
// Consecutive read-only calls are grouped into a single concurrent batch.
// Non-read-only calls each get their own sequential batch of size 1.
func PartitionToolCalls(approved []approvedToolCall) [][]approvedToolCall {
	if len(approved) == 0 {
		return nil
	}
	var batches [][]approvedToolCall
	var currentBatch []approvedToolCall
	currentIsReadOnly := false

	for i, ac := range approved {
		ro := isReadOnlyTool(ac)
		if i == 0 {
			currentBatch = []approvedToolCall{ac}
			currentIsReadOnly = ro
			continue
		}
		if ro && currentIsReadOnly {
			currentBatch = append(currentBatch, ac)
		} else {
			batches = append(batches, currentBatch)
			currentBatch = []approvedToolCall{ac}
			currentIsReadOnly = ro
		}
	}
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}
	return batches
}

// ExecuteBatches runs partitioned tool call batches sequentially.
// Read-only batches run concurrently with a channel semaphore capped at maxToolConcurrency.
// Write batches run one at a time.
func ExecuteBatches(ctx context.Context, batches [][]approvedToolCall, execResults []toolExecResult, readTracker *ReadTracker, handler EventHandler) {
	for _, batch := range batches {
		if len(batch) == 1 {
			ac := batch[0]
			func() {
				defer func() {
					if r := recover(); r != nil {
						execResults[ac.index] = toolExecResult{
							result: ToolResult{Content: fmt.Sprintf("tool panicked: %v", r), IsError: true},
						}
					}
				}()
				if handler != nil {
					handler.OnToolCall(ac.tool.Info().Name, ac.argsStr)
				}
				startTime := time.Now()
				result, runErr := ac.tool.Run(ctx, ac.argsStr)
				execResults[ac.index] = toolExecResult{result: result, elapsed: time.Since(startTime), err: runErr}
			}()
		} else {
			sem := make(chan struct{}, maxToolConcurrency)
			var wg sync.WaitGroup
			wg.Add(len(batch))
			for _, ac := range batch {
				sem <- struct{}{}
				if handler != nil {
					handler.OnToolCall(ac.tool.Info().Name, ac.argsStr)
				}
				go func(ac approvedToolCall) {
					defer wg.Done()
					defer func() { <-sem }()
					defer func() {
						if r := recover(); r != nil {
							execResults[ac.index] = toolExecResult{
								result: ToolResult{Content: fmt.Sprintf("tool panicked: %v", r), IsError: true},
							}
						}
					}()
					startTime := time.Now()
					result, runErr := ac.tool.Run(ctx, ac.argsStr)
					execResults[ac.index] = toolExecResult{result: result, elapsed: time.Since(startTime), err: runErr}
				}(ac)
			}
			wg.Wait()
		}

		// Inter-batch side effect: track file_read results for ReadTracker
		if readTracker != nil {
			for _, ac := range batch {
				info := ac.tool.Info()
				if info.Name == "file_read" {
					er := execResults[ac.index]
					if !er.result.IsError && er.err == nil {
						if p := extractPathArg(ac.argsStr); p != "" {
							readTracker.MarkRead(p)
						}
					}
				}
			}
		}
	}
}

// extractPathArg extracts the "path" field from JSON arguments string.
func extractPathArg(argsJSON string) string {
	var args struct {
		Path string `json:"path"`
	}
	_ = jsonUnmarshal([]byte(argsJSON), &args)
	return args.Path
}

// jsonUnmarshal is a wrapper for json.Unmarshal.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

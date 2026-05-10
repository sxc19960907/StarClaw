package agent

import (
	"context"
	"fmt"
	"sync"
)

const defaultConcurrencyLimit = 5

// ToolCall represents a callable tool invocation for parallel execution.
type ToolCall struct {
	// Name is the tool name for identification.
	Name string
	// Args is the JSON-encoded arguments string.
	Args string
	// Exec executes the tool call with the given context and args.
	Exec func(ctx context.Context, args string) (ToolResult, error)
}

// PartitionConcurrency manages parallel execution of independent tool calls
// with bounded concurrency. Tool calls are executed in separate goroutines
// and results are collected in input order.
type PartitionConcurrency struct {
	maxConcurrency int
}

// NewPartitionConcurrency creates a PartitionConcurrency with the default
// concurrency limit of 5.
func NewPartitionConcurrency() *PartitionConcurrency {
	return &PartitionConcurrency{maxConcurrency: defaultConcurrencyLimit}
}

// ExecuteParallel runs the given tool calls in parallel with bounded concurrency.
// Each tool call runs in its own goroutine. Results are returned in the same
// order as the input tools slice. If handler is non-nil, OnToolCall is called
// for each tool immediately before execution begins.
func (pc *PartitionConcurrency) ExecuteParallel(tools []ToolCall, handler EventHandler) []ToolResult {
	if len(tools) == 0 {
		return nil
	}

	results := make([]ToolResult, len(tools))
	sem := make(chan struct{}, pc.maxConcurrency)
	var wg sync.WaitGroup
	wg.Add(len(tools))

	for i := range tools {
		sem <- struct{}{} // acquire semaphore slot
		tc := tools[i]
		if handler != nil {
			handler.OnToolCall(tc.Name, tc.Args)
		}

		go func(idx int, call ToolCall) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore slot
			defer func() {
				if r := recover(); r != nil {
					results[idx] = ToolResult{
						Content: fmt.Sprintf("tool panicked: %v", r),
						IsError: true,
					}
				}
			}()

			result, err := call.Exec(context.Background(), call.Args)
			if err != nil {
				results[idx] = ToolResult{
					Content: err.Error(),
					IsError: true,
				}
				return
			}
			results[idx] = result
		}(i, tc)
	}

	wg.Wait()
	return results
}

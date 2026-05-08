package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/session"
)

// State represents the TUI state
type State int

const (
	StateIdle State = iota
	StateThinking
	StateAwaitingApproval
	StateStreaming
)

// Message represents a chat message
type Message struct {
	Role    string
	Content string
	ToolCall *ToolCallInfo
}

// ToolCallInfo represents tool call information
type ToolCallInfo struct {
	Name   string
	Args   string
	Result string
	Error  bool
	Approved bool
}

// Model is the TUI model
type Model struct {
	// Core components
	loop    *agent.AgentLoop
	ctx     context.Context

	// UI Components
	textarea textarea.Model
	messages []Message
	viewport int

	// State
	state        State
	pendingTool  *ToolCallInfo
	width        int
	height       int

	// Startup header animation
	headerFrame    int
	headerDone     bool
	headerSessions []session.SessionSummary
	headerTipIdx   int
	headerCWD      string

	// Tool result tracking
	lastToolResults []toolResultEntry
	toolExpandLevel int
	processingStartTime time.Time

	// Styling
	userStyle      lipgloss.Style
	assistantStyle lipgloss.Style
	systemStyle    lipgloss.Style
	toolStyle      lipgloss.Style
	errorStyle     lipgloss.Style
	inputStyle     lipgloss.Style
}

// NewModel creates a new TUI model
func NewModel(loop *agent.AgentLoop) *Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message... (Ctrl+Enter to send, Ctrl+Q to quit, Ctrl+L to clear)"
	ta.Focus()

	return &Model{
		loop:     loop,
		ctx:      context.Background(),
		textarea: ta,
		state:    StateIdle,
		messages: make([]Message, 0),
		// Header animation disabled by default (enabled in Init)
		headerDone: true,
		userStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true),
		assistantStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")),
		systemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
		toolStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true),
		inputStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	// Start header animation
	m.headerFrame = 0
	m.headerDone = false
	m.headerSessions = nil
	m.headerTipIdx = pickTipIdx()
	m.headerCWD, _ = os.Getwd()
	return tea.Batch(
		textarea.Blink,
		headerFrameTick(),
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle header animation early return
	if !m.headerDone {
		switch msg := msg.(type) {
		case headerTickMsg:
			m.headerFrame++
			if m.headerFrame >= headerTotalFrames {
				m.headerDone = true
			}
			return m, headerFrameTick()
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.textarea.SetWidth(msg.Width - 4)
			return m, nil
		case tea.KeyMsg:
			// Skip animation on any key except Ctrl+C/Q
			if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlQ {
				return m, tea.Quit
			}
			m.headerDone = true
			return m, nil
		default:
			return m, nil
		}
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlQ:
			return m, tea.Quit

		case tea.KeyCtrlL:
			// Clear screen
			m.messages = make([]Message, 0)
			return m, nil

		case tea.KeyCtrlY:
			// Auto-approve all
			if m.state == StateAwaitingApproval && m.pendingTool != nil {
				m.pendingTool.Approved = true
				m.state = StateThinking
				// Continue processing
				return m, m.processToolResponse(true)
			}

		case tea.KeyEnter:
			// Check for Ctrl+Enter
			if msg.Alt {
				// Send message
				content := strings.TrimSpace(m.textarea.Value())
				if content != "" && m.state == StateIdle {
					m.textarea.SetValue("")
					return m, tea.Batch(
						m.sendMessage(content),
						textarea.Blink,
					)
				}
			}
		}

		// Handle approval keys
		if m.state == StateAwaitingApproval {
			switch msg.String() {
			case "y", "Y":
				if m.pendingTool != nil {
					m.pendingTool.Approved = true
					m.state = StateThinking
					return m, m.processToolResponse(true)
				}
			case "n", "N":
				if m.pendingTool != nil {
					m.pendingTool.Approved = false
					m.state = StateIdle
					return m, m.processToolResponse(false)
				}
			}
		}

		// Ctrl+O: expand tool results (shows detailed args/result)
		if msg.String() == "ctrl+o" && len(m.lastToolResults) > 0 && m.toolExpandLevel == 0 {
			for _, r := range m.lastToolResults {
				expanded := formatExpandedToolResult(r.name, r.args, r.isError, r.content, r.elapsed)
				m.messages = append(m.messages, Message{
					Role:    "system",
					Content: expanded,
				})
			}
			m.toolExpandLevel = 1
			return m, nil
		}

		// Update textarea
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

	case agentMessage:
		// Agent response message
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: string(msg),
		})
		m.state = StateIdle
		return m, nil

	case streamingMsg:
		// Streaming text update - append to last assistant message or create new one
		m.state = StateStreaming
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "assistant" {
			// Append to existing assistant message
			m.messages[len(m.messages)-1].Content += string(msg)
		} else {
			// Create new assistant message
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: string(msg),
			})
		}
		return m, nil

	case toolCallMsg:
		// Tool call started
		m.state = StateAwaitingApproval
		m.pendingTool = &ToolCallInfo{
			Name: msg.name,
			Args: msg.args,
		}
		m.messages = append(m.messages, Message{
			Role:     "system",
			ToolCall: m.pendingTool,
		})
		return m, nil

	case toolResultMsg:
		// Tool result received
		if m.pendingTool != nil {
			m.pendingTool.Result = msg.result
			m.pendingTool.Error = msg.isError
		}
		return m, nil

	case usageMsg:
		// Usage info
		m.messages = append(m.messages, Message{
			Role:    "system",
			Content: fmt.Sprintf("📊 Usage: %d input, %d output tokens", msg.input, msg.output),
		})
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *Model) View() string {
	var b strings.Builder

	// Show startup header animation during initialization
	if !m.headerDone {
		width := m.width
		if width <= 0 {
			width = 80
		}
		b.WriteString(renderStartupHeader(m.headerFrame, width, "dev", "small", "", m.headerCWD, m.headerSessions, m.headerTipIdx))
		return b.String()
	}

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render("StarClaw AI Agent")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Messages
	for _, msg := range m.messages {
		b.WriteString(m.renderMessage(msg))
		b.WriteString("\n")
	}

	// Approval dialog
	if m.state == StateAwaitingApproval && m.pendingTool != nil {
		b.WriteString(m.renderApprovalDialog())
		b.WriteString("\n")
	}

	// Status line
	if m.state == StateThinking {
		b.WriteString(m.systemStyle.Render("🤔 Thinking..."))
		b.WriteString("\n")
	} else if m.state == StateStreaming {
		b.WriteString(m.systemStyle.Render("✨ Receiving..."))
		b.WriteString("\n")
	}

	// Input area
	b.WriteString("\n")
	inputView := m.inputStyle.Width(m.width - 4).Render(m.textarea.View())
	b.WriteString(inputView)

	// Help
	help := m.systemStyle.Render("Ctrl+Enter: Send | Ctrl+Q: Quit | Ctrl+L: Clear | Ctrl+Y: Auto-approve")
	b.WriteString("\n")
	b.WriteString(help)

	return b.String()
}

// renderMessage renders a single message
func (m *Model) renderMessage(msg Message) string {
	switch msg.Role {
	case "user":
		return m.userStyle.Render("You: ") + msg.Content
	case "assistant":
		rendered := renderMarkdown(msg.Content, m.width)
		return m.assistantStyle.Render("Assistant: ") + rendered
	case "system":
		if msg.ToolCall != nil {
			return m.renderToolCall(msg.ToolCall)
		}
		return m.systemStyle.Render(msg.Content)
	default:
		return msg.Content
	}
}

// renderToolCall renders a tool call
func (m *Model) renderToolCall(tool *ToolCallInfo) string {
	if tool.Result != "" {
		return formatCompactToolResult(tool.Name, tool.Args, tool.Error, tool.Result, 0)
	}
	// Pending tool call
	return m.toolStyle.Render(fmt.Sprintf("⏵ %s(%s)", tool.Name, toolKeyArg(tool.Name, tool.Args)))
}

// renderApprovalDialog renders the approval dialog
func (m *Model) renderApprovalDialog() string {
	var b strings.Builder

	b.WriteString("\n")
	keyArg := toolKeyArg(m.pendingTool.Name, m.pendingTool.Args)
	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("3")).
		Padding(1).
		Render(
			m.toolStyle.Render("⚠️  Tool Approval Required\n\n") +
			fmt.Sprintf("Tool: %s\n", m.pendingTool.Name) +
			fmt.Sprintf("Args: %s\n\n", keyArg) +
			"Approve? [Y/n] (or Ctrl+Y to auto-approve all)",
		))

	return b.String()
}

// sendMessage sends a message to the agent
func (m *Model) sendMessage(content string) tea.Cmd {
	return func() tea.Msg {
		// Add user message
		m.messages = append(m.messages, Message{
			Role:    "user",
			Content: content,
		})

		m.state = StateThinking

		// Create event handler for this conversation
		handler := &TUIEventHandler{
			model: m,
		}
		m.loop.SetEventHandler(handler)

		// Run the agent loop
		resp, err := m.loop.Run(m.ctx, content)
		if err != nil {
			return agentMessage(fmt.Sprintf("Error: %v", err))
		}

		return agentMessage(resp.Content)
	}
}

// processToolResponse handles tool approval response
func (m *Model) processToolResponse(approved bool) tea.Cmd {
	return func() tea.Msg {
		// This would continue the agent loop
		// For now, just return to idle
		if !approved {
			return agentMessage("Tool call cancelled by user.")
		}
		return nil
	}
}

// TUIEventHandler handles events for TUI
type TUIEventHandler struct {
	model *Model
}

func (h *TUIEventHandler) OnToolCall(name string, args string) {
	// Send tool call message to update UI
	// This would need a proper command to update state
	h.model.pendingTool = &ToolCallInfo{
		Name: name,
		Args: args,
	}
	h.model.state = StateAwaitingApproval
}

func (h *TUIEventHandler) OnToolResult(name string, result agent.ToolResult) {
	if h.model.pendingTool != nil {
		h.model.pendingTool.Result = result.Content
		h.model.pendingTool.Error = result.IsError

		// Track tool result for expandable display
		entry := toolResultEntry{
			name:    name,
			args:    h.model.pendingTool.Args,
			content: result.Content,
			isError: result.IsError,
			elapsed: 0,
		}
		h.model.lastToolResults = append(h.model.lastToolResults, entry)
		if len(h.model.lastToolResults) > 20 {
			h.model.lastToolResults = h.model.lastToolResults[1:]
		}
		h.model.toolExpandLevel = 0
	}
}

func (h *TUIEventHandler) OnText(text string) {
	// Text updates happen via agentMessage
}

func (h *TUIEventHandler) OnUsage(usage client.Usage) {
	// Usage updates
}

func (h *TUIEventHandler) OnStreamDelta(delta string) {
	// Streamed text handling for TUI (to be implemented)
}

// Message types for tea.Cmd
type agentMessage string
type streamingMsg string
type toolCallMsg struct {
	name string
	args string
}
type toolResultMsg struct {
	result  string
	isError bool
}
type usageMsg struct {
	input  int
	output int
}

// headerTickMsg advances the startup header animation by one frame.
type headerTickMsg struct{}

// toolResultEntry stores a single tool result for display.
type toolResultEntry struct {
	name    string
	args    string
	content string
	isError bool
	elapsed time.Duration
}

// truncate truncates a string
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Run starts the TUI
func Run(loop *agent.AgentLoop) error {
	p := tea.NewProgram(NewModel(loop), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

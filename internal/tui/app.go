package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
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
	return textarea.Blink
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m.assistantStyle.Render("Assistant: ") + msg.Content
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
	var b strings.Builder

	b.WriteString(m.toolStyle.Render(fmt.Sprintf("🔧 Tool: %s", tool.Name)))
	if tool.Args != "" {
		b.WriteString(m.systemStyle.Render(fmt.Sprintf("  Args: %s", truncate(tool.Args, 80))))
	}

	if tool.Result != "" {
		if tool.Error {
			b.WriteString("\n")
			b.WriteString(m.errorStyle.Render(fmt.Sprintf("   ❌ Error: %s", truncate(tool.Result, 100))))
		} else {
			b.WriteString("\n")
			b.WriteString(m.systemStyle.Render("   ✅ Done"))
		}
	}

	return b.String()
}

// renderApprovalDialog renders the approval dialog
func (m *Model) renderApprovalDialog() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("3")).
		Padding(1).
		Render(
			m.toolStyle.Render("⚠️  Tool Approval Required\n\n") +
			fmt.Sprintf("Tool: %s\n", m.pendingTool.Name) +
			fmt.Sprintf("Args: %s\n\n", truncate(m.pendingTool.Args, 100)) +
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
	}
}

func (h *TUIEventHandler) OnText(text string) {
	// Text updates happen via agentMessage
}

func (h *TUIEventHandler) OnUsage(usage client.Usage) {
	// Usage updates
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

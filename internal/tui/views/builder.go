package views

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/adammpkins/req/internal/parser"
	"github.com/adammpkins/req/internal/planner"
	"github.com/adammpkins/req/internal/runtime"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(1, 2).
			Width(80)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Padding(1, 2).
			Width(80)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(1, 2).
			Width(80)

	// JSON syntax highlighting styles
	jsonKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	jsonStringStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	jsonNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	jsonBoolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213"))

	jsonNullStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	jsonPunctStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	outputStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)

// View represents a TUI view interface.
type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
}

// BuilderView is an interactive command builder.
type BuilderView struct {
	form         *huh.Form
	executed     bool
	response     string
	responseBody string
	formattedBody string
	err          error
	verb         string
	url          string
	execute      bool
	width        int
	height       int
	viewport     viewport.Model
}

// NewBuilderView creates a new builder view.
func NewBuilderView() View {
	vp := viewport.New(80, 20) // default width and height
	b := &BuilderView{
		width:    80, // default width
		height:   20, // default height
		viewport: vp,
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Verb").
				Description("Select the action to perform").
				Options(
					huh.NewOption("read - Read a resource (GET)", "read"),
					huh.NewOption("save - Save a resource to file (GET)", "save"),
					huh.NewOption("send - Send data (POST)", "send"),
				).
				Value(&b.verb).
				Key("verb"),

			huh.NewInput().
				Title("URL").
				Description("Enter the target URL").
				Placeholder("https://api.example.com/users").
				Value(&b.url).
				Key("url"),

			huh.NewConfirm().
				Title("Execute immediately?").
				Description("Execute the command when form is complete").
				Value(&b.execute).
				Key("execute"),
		),
	)

	b.form = form
	return b
}

// Init initializes the view.
func (b *BuilderView) Init() tea.Cmd {
	return b.form.Init()
}

// Update handles messages.
func (b *BuilderView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle window size messages first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
		if b.width == 0 {
			b.width = 80 // default width
		}
		if b.height == 0 {
			b.height = 20 // default height
		}
		// Update viewport size
		b.updateViewportSize()
	}

	// Handle keyboard input
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		
		// If we have output to scroll, handle scrolling keys first
		if b.formattedBody != "" {
			// Check if it's a scrolling key
			switch key {
			case "up", "k", "pgup":
				b.viewport.LineUp(1)
				return b, nil
			case "down", "j", "pgdown":
				b.viewport.LineDown(1)
				return b, nil
			case "home":
				b.viewport.GotoTop()
				return b, nil
			case "end":
				b.viewport.GotoBottom()
				return b, nil
			case "ctrl+u":
				b.viewport.LineUp(b.viewport.Height / 2)
				return b, nil
			case "ctrl+d":
				b.viewport.LineDown(b.viewport.Height / 2)
				return b, nil
			case "esc":
				return b, tea.Quit
			}
		} else {
			// No output, just handle quit
			switch key {
			case "esc":
				return b, tea.Quit
			}
		}
	}

	// Update form (only if not a scrolling key when we have output)
	form, cmd := b.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		b.form = f
		cmds = append(cmds, cmd)
	}

	// Handle messages from command execution
	switch msg := msg.(type) {
	case ErrorMsg:
		b.err = msg.Err
		b.executed = false
		b.responseBody = ""
		b.formattedBody = ""
		b.viewport.SetContent("")
	case SuccessMsg:
		b.response = msg.Message
		b.responseBody = msg.ResponseBody
		b.err = nil
		// Format the response body
		b.updateFormattedBody()
	}
	
	// Handle viewport updates for other messages (like mouse wheel, etc.)
	if b.formattedBody != "" {
		vp, cmd := b.viewport.Update(msg)
		b.viewport = vp
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Check if form is complete and should execute
	if b.form.State == huh.StateCompleted {
		if !b.executed && b.verb != "" && b.url != "" {
			// Values are already bound to b.verb, b.url, and b.execute via Value() in form creation
			// The bound variables are updated automatically when form fields change
			if b.execute {
				b.executed = true
				cmds = append(cmds, b.executeCommand())
			} else {
				// Form completed but execute was false - show message
				b.response = "Command built but not executed. Press 'esc' to exit."
			}
		}
	}

	return b, tea.Batch(cmds...)
}

// executeCommand executes the built command.
func (b *BuilderView) executeCommand() tea.Cmd {
	return func() tea.Msg {
		// Build command string
		cmdStr := fmt.Sprintf("%s %s", b.verb, b.url)
		
		// Parse command
		cmd, err := parser.Parse(cmdStr)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Plan execution
		plan, err := planner.Plan(cmd)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Execute
		executor, err := runtime.NewExecutor(plan)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Capture response body for TUI display
		responseBody, err := executor.ExecuteWithResponse(plan)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return SuccessMsg{
			Message:      "Command executed successfully",
			ResponseBody: responseBody,
		}
	}
}

// View renders the view.
func (b *BuilderView) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("req - Interactive Command Builder"))
	s.WriteString("\n\n")

	if b.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", b.err)))
		s.WriteString("\n\n")
	}

	if b.response != "" {
		s.WriteString(successStyle.Render(b.response))
		s.WriteString("\n\n")
	}

	// Display response body with formatting (using viewport for scrolling)
	if b.formattedBody != "" {
		// Calculate available width and height for viewport
		contentWidth := b.width - 6 // Account for border and padding
		if contentWidth < 20 {
			contentWidth = 20 // Minimum width
		}
		
		// Calculate available height (account for header, success message, form, command line, instructions)
		// Rough estimate: title ~3, success ~2, form ~varies, command ~2, instructions ~1 = ~8-10 lines
		// Reserve some space for the form and other UI elements
		availableHeight := b.height - 15 // Reserve space for other UI elements
		if availableHeight < 5 {
			availableHeight = 5 // Minimum height
		}
		
		// Update viewport dimensions if needed
		b.updateViewportSize()
		
		// Render viewport with border
		// The viewport handles its own height, so we just need to wrap it with the border style
		viewportContent := b.viewport.View()
		// Use the viewport's actual dimensions for the border
		s.WriteString(outputStyle.Width(contentWidth + 4).Render(viewportContent))
		s.WriteString("\n\n")
	}

	s.WriteString(b.form.View())
	
	// Show current values when form is completed
	if b.form.State == huh.StateCompleted {
		s.WriteString("\n\n")
		if b.verb != "" && b.url != "" {
			cmdText := fmt.Sprintf("Command: %s %s", b.verb, b.url)
			// Wrap the command text to fit terminal width
			width := b.width
			if width == 0 {
				width = 80 // default width
			}
			wrapped := wrapText(cmdText, width)
			s.WriteString(commandStyle.Render(wrapped))
			s.WriteString("\n")
			if b.response != "" {
				s.WriteString("\n")
			}
		}
	}
	
	s.WriteString("\n")
	if b.formattedBody != "" {
		s.WriteString("Press 'esc' to quit, 'ctrl+c' to exit, ↑/↓ to scroll, pgup/pgdn for page scroll\n")
	} else {
		s.WriteString("Press 'esc' to quit, 'ctrl+c' to exit\n")
	}

	return s.String()
}

// ErrorMsg represents an error message.
type ErrorMsg struct {
	Err error
}

// SuccessMsg represents a success message.
type SuccessMsg struct {
	Message      string
	ResponseBody string
}

// wrapText wraps text to the specified width, breaking at word boundaries.
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}
	
	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""
	
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word
		
		if len(testLine) > width {
			if currentLine != "" {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			} else {
				// Word is longer than width, just add it
				result.WriteString(word)
				result.WriteString("\n")
				currentLine = ""
			}
		} else {
			currentLine = testLine
		}
	}
	
	if currentLine != "" {
		result.WriteString(currentLine)
	}
	
	return result.String()
}

// updateFormattedBody formats the response body and updates the viewport content.
func (b *BuilderView) updateFormattedBody() {
	if b.responseBody == "" {
		b.formattedBody = ""
		b.viewport.SetContent("")
		return
	}
	
	// Calculate available width for content
	contentWidth := b.width - 6 // Account for border and padding
	if contentWidth < 20 {
		contentWidth = 20 // Minimum width
	}
	
	// Format the response
	formatted := formatResponse(b.responseBody, contentWidth)
	b.formattedBody = formatted
	
	// Update viewport content
	b.viewport.SetContent(formatted)
	b.viewport.GotoTop() // Start at the top
}

// updateViewportSize updates the viewport dimensions based on available space.
func (b *BuilderView) updateViewportSize() {
	if b.responseBody == "" {
		return
	}
	
	// Calculate available width and height
	contentWidth := b.width - 6 // Account for border and padding
	if contentWidth < 20 {
		contentWidth = 20 // Minimum width
	}
	
	availableHeight := b.height - 15 // Reserve space for other UI elements
	if availableHeight < 5 {
		availableHeight = 5 // Minimum height
	}
	
	// Update viewport dimensions
	b.viewport.Width = contentWidth
	b.viewport.Height = availableHeight
	
	// If content is already set, ensure it's properly sized
	if b.formattedBody != "" {
		b.viewport.SetContent(b.formattedBody)
	}
}

// formatResponse formats the response body with syntax highlighting for JSON.
func formatResponse(body string, width int) string {
	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
		// It's valid JSON, format it with syntax highlighting
		return formatJSON(body, width)
	}

	// Not JSON, return as-is with word wrapping
	return wrapText(body, width)
}

// formatJSON formats JSON with syntax highlighting using lipgloss.
func formatJSON(jsonStr string, width int) string {
	// First, pretty-print the JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return jsonStr // Return original if parsing fails
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return jsonStr // Return original if formatting fails
	}

	// Now apply syntax highlighting
	lines := strings.Split(string(prettyJSON), "\n")
	var formattedLines []string

	for _, line := range lines {
		formattedLine := highlightJSONLine(line)
		formattedLines = append(formattedLines, formattedLine)
	}

	return strings.Join(formattedLines, "\n")
}

// highlightJSONLine applies syntax highlighting to a single line of JSON.
func highlightJSONLine(line string) string {
	// This is a simple JSON highlighter that handles common cases
	// For a more robust solution, consider using a proper JSON tokenizer
	
	result := ""
	i := 0
	
	for i < len(line) {
		char := line[i]
		
		// Skip whitespace
		if char == ' ' || char == '\t' {
			result += string(char)
			i++
			continue
		}
		
		// Handle string literals
		if char == '"' {
			end := i + 1
			escaped := false
			for end < len(line) {
				if line[end] == '\\' && !escaped {
					escaped = true
					end++
				} else if line[end] == '"' && !escaped {
					end++
					// Check if this is a key (followed by :)
					isKey := end < len(line) && line[end] == ':'
					str := line[i:end]
					if isKey {
						result += jsonKeyStyle.Render(str)
					} else {
						result += jsonStringStyle.Render(str)
					}
					i = end
					break
				} else {
					escaped = false
					end++
				}
			}
			if end >= len(line) {
				// Unterminated string, just add it
				result += jsonStringStyle.Render(line[i:])
				break
			}
			continue
		}
		
		// Handle numbers
		if (char >= '0' && char <= '9') || char == '-' {
			start := i
			for i < len(line) && ((line[i] >= '0' && line[i] <= '9') || 
				line[i] == '.' || line[i] == 'e' || line[i] == 'E' || 
				line[i] == '+' || line[i] == '-' || line[i] == 'i' || 
				line[i] == 'n' || line[i] == 'f') {
				i++
			}
			result += jsonNumberStyle.Render(line[start:i])
			continue
		}
		
		// Handle boolean and null
		if strings.HasPrefix(line[i:], "true") {
			result += jsonBoolStyle.Render("true")
			i += 4
			continue
		}
		if strings.HasPrefix(line[i:], "false") {
			result += jsonBoolStyle.Render("false")
			i += 5
			continue
		}
		if strings.HasPrefix(line[i:], "null") {
			result += jsonNullStyle.Render("null")
			i += 4
			continue
		}
		
		// Handle punctuation
		if char == '{' || char == '}' || char == '[' || char == ']' || 
		   char == ',' || char == ':' {
			result += jsonPunctStyle.Render(string(char))
			i++
			continue
		}
		
		// Default: just add the character
		result += string(char)
		i++
	}
	
	return result
}

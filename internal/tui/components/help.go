// Package components provides reusable TUI components.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Binding represents a keyboard binding for help display.
type Binding struct {
	// Key is the key sequence (e.g., "ctrl+c", "q", "Enter").
	Key string

	// Description describes what action the key performs.
	Description string

	// Style is the style applied to the key display.
	KeyStyle lipgloss.Style

	// DescStyle is the style applied to the description.
	DescStyle lipgloss.Style
}

// Help represents a help overlay component that displays key bindings.
type Help struct {
	// Title is the help overlay title.
	Title string

	// Bindings contains all the key bindings to display.
	Bindings []Binding

	// Styles
	TitleStyle     lipgloss.Style
	KeyStyle       lipgloss.Style
	DescStyle      lipgloss.Style
	Separator      string
	SeparatorStyle lipgloss.Style

	// Layout
	Columns int // Number of columns for bindings (0 for single column)

	// MaxWidth limits the total width of the help display.
	MaxWidth int
}

// NewHelp creates a new help component with default settings.
func NewHelp() *Help {
	defaultKeyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true)

	return &Help{
		Title:          "Key Bindings",
		Bindings:       nil,
		TitleStyle:     lipgloss.NewStyle().Bold(true),
		KeyStyle:       defaultKeyStyle,
		DescStyle:      lipgloss.NewStyle(),
		Separator:      " : ",
		SeparatorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		Columns:        2,
		MaxWidth:       0,
	}
}

// WithTitle sets the help title.
func (h *Help) WithTitle(title string) *Help {
	h.Title = title
	return h
}

// WithBindings sets the key bindings.
func (h *Help) WithBindings(bindings []Binding) *Help {
	h.Bindings = bindings
	return h
}

// AddBinding appends a key binding.
func (h *Help) AddBinding(binding Binding) *Help {
	h.Bindings = append(h.Bindings, binding)
	return h
}

// AddSimpleBinding adds a simple key binding (key + description only).
func (h *Help) AddSimpleBinding(key, desc string) *Help {
	return h.AddBinding(Binding{
		Key:         key,
		Description: desc,
	})
}

// WithTitleStyle sets the title style.
func (h *Help) WithTitleStyle(style lipgloss.Style) *Help {
	h.TitleStyle = style
	return h
}

// WithKeyStyle sets the default key style.
func (h *Help) WithKeyStyle(style lipgloss.Style) *Help {
	h.KeyStyle = style
	return h
}

// WithDescStyle sets the default description style.
func (h *Help) WithDescStyle(style lipgloss.Style) *Help {
	h.DescStyle = style
	return h
}

// WithSeparator sets the separator between key and description.
func (h *Help) WithSeparator(sep string) *Help {
	h.Separator = sep
	return h
}

// WithSeparatorStyle sets the separator style.
func (h *Help) WithSeparatorStyle(style lipgloss.Style) *Help {
	h.SeparatorStyle = style
	return h
}

// WithColumns sets the number of columns for bindings.
func (h *Help) WithColumns(cols int) *Help {
	h.Columns = cols
	return h
}

// WithMaxWidth sets the maximum width of the help display.
func (h *Help) WithMaxWidth(width int) *Help {
	h.MaxWidth = width
	return h
}

// renderBinding renders a single binding as a string.
func (h *Help) renderBinding(b Binding) string {
	// Use binding's style, or default to help's style
	// Since we can't detect "unset" style, always use binding's styles
	// Users should set them if they want custom styling
	keyStyle := b.KeyStyle
	descStyle := b.DescStyle

	// If binding styles weren't set (they're effectively empty), use help's defaults
	// We'll just use the binding's style as-is - if it's empty, it renders unstyled
	key := keyStyle.Render(b.Key)
	sep := h.SeparatorStyle.Render(h.Separator)
	desc := descStyle.Render(b.Description)

	return key + sep + desc
}

// calculateColumnWidths determines optimal widths for multi-column layout.
func (h *Help) calculateColumnWidths() []int {
	if h.Columns <= 1 || len(h.Bindings) == 0 {
		return nil
	}

	rowsPerCol := (len(h.Bindings) + h.Columns - 1) / h.Columns
	widths := make([]int, h.Columns)

	for col := 0; col < h.Columns; col++ {
		maxWidth := 0
		for row := 0; row < rowsPerCol; row++ {
			idx := row + col*rowsPerCol
			if idx >= len(h.Bindings) {
				break
			}
			width := lipgloss.Width(h.renderBinding(h.Bindings[idx]))
			if width > maxWidth {
				maxWidth = width
			}
		}
		widths[col] = maxWidth
	}

	return widths
}

// Render renders the help overlay.
func (h *Help) Render() string {
	var builder strings.Builder

	// Render title
	if h.Title != "" {
		builder.WriteString(h.TitleStyle.Render(h.Title))
		builder.WriteString("\n\n")
	}

	if len(h.Bindings) == 0 {
		return builder.String()
	}

	// Single column layout
	if h.Columns <= 1 {
		for _, binding := range h.Bindings {
			builder.WriteString(h.renderBinding(binding))
			builder.WriteString("\n")
		}
		return builder.String()
	}

	// Multi-column layout
	colWidths := h.calculateColumnWidths()
	rowsPerCol := (len(h.Bindings) + h.Columns - 1) / h.Columns

	for row := 0; row < rowsPerCol; row++ {
		for col := 0; col < h.Columns; col++ {
			idx := row + col*rowsPerCol
			if idx >= len(h.Bindings) {
				break
			}

			bindingText := h.renderBinding(h.Bindings[idx])

			// Add padding for alignment
			if col < len(colWidths) && col < h.Columns-1 {
				currentWidth := lipgloss.Width(bindingText)
				padding := colWidths[col] - currentWidth + 4 // 4 spaces between columns
				if padding > 0 {
					bindingText += strings.Repeat(" ", padding)
				}
			}

			builder.WriteString(bindingText)
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// String returns the rendered help.
func (h *Help) String() string {
	return h.Render()
}

// Width returns the width of the rendered help.
func (h *Help) Width() int {
	if h.Columns <= 1 {
		maxWidth := 0
		for _, b := range h.Bindings {
			width := lipgloss.Width(h.renderBinding(b))
			if width > maxWidth {
				maxWidth = width
			}
		}
		return maxWidth
	}

	colWidths := h.calculateColumnWidths()
	if colWidths == nil {
		return 0
	}

	totalWidth := 0
	for _, w := range colWidths {
		totalWidth += w
	}
	totalWidth += (h.Columns - 1) * 4 // Column gap

	return totalWidth
}

// Height returns the number of lines in the rendered help.
func (h *Help) Height() int {
	height := len(h.Bindings)
	if h.Title != "" {
		height += 2 // Title + blank line
	}

	if h.Columns > 1 && height > 0 {
		rowsPerCol := (len(h.Bindings) + h.Columns - 1) / h.Columns
		if h.Title != "" {
			height = rowsPerCol + 2
		} else {
			height = rowsPerCol
		}
	}

	return height
}

// BindingCount returns the number of bindings.
func (h *Help) BindingCount() int {
	return len(h.Bindings)
}

// Clear removes all bindings.
func (h *Help) Clear() *Help {
	h.Bindings = nil
	return h
}

// HelpConfig provides configuration for creating a help component.
type HelpConfig struct {
	Title          string
	Bindings       []Binding
	TitleStyle     lipgloss.Style
	KeyStyle       lipgloss.Style
	DescStyle      lipgloss.Style
	Separator      string
	SeparatorStyle lipgloss.Style
	Columns        int
	MaxWidth       int
}

// NewHelpFromConfig creates a help component from a configuration.
func NewHelpFromConfig(cfg HelpConfig) *Help {
	h := NewHelp()
	h.WithTitle(cfg.Title)
	h.WithBindings(cfg.Bindings)
	h.WithTitleStyle(cfg.TitleStyle)
	h.WithKeyStyle(cfg.KeyStyle)
	h.WithDescStyle(cfg.DescStyle)
	h.WithSeparator(cfg.Separator)
	h.WithSeparatorStyle(cfg.SeparatorStyle)
	h.WithColumns(cfg.Columns)
	h.WithMaxWidth(cfg.MaxWidth)
	return h
}

// KeyGroup represents a group of related key bindings.
type KeyGroup struct {
	Title    string
	Bindings []Binding
}

// RenderHelpWithGroups renders help with grouped bindings.
func RenderHelpWithGroups(groups []KeyGroup, titleStyle, keyStyle, descStyle lipgloss.Style) string {
	var builder strings.Builder

	for i, group := range groups {
		if group.Title != "" {
			builder.WriteString(titleStyle.Render(group.Title))
			builder.WriteString("\n")
		}

		for _, binding := range group.Bindings {
			key := keyStyle.Render(binding.Key)
			sep := " : "
			desc := descStyle.Render(binding.Description)
			builder.WriteString(key + sep + desc + "\n")
		}

		if i < len(groups)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

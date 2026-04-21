// Package palette provides command builders for the palette.
package palette

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/tui/macro"
	"github.com/shinsekai/finterm/internal/tui/news"
	"github.com/shinsekai/finterm/internal/tui/quote"
	"github.com/shinsekai/finterm/internal/tui/trend"
)

// CommandBuilder builds palette commands.
type CommandBuilder struct {
	commands []Command
}

// NewBuilder creates a new command builder.
func NewBuilder() *CommandBuilder {
	return &CommandBuilder{
		commands: []Command{},
	}
}

// Add adds a command to the builder.
func (b *CommandBuilder) Add(cmd Command) *CommandBuilder {
	b.commands = append(b.commands, cmd)
	return b
}

// AddTab adds a tab navigation command.
func (b *CommandBuilder) AddTab(id, label, description, shortcut string, tab int) *CommandBuilder {
	b.commands = append(b.commands, Command{
		ID:          fmt.Sprintf("go.%s", id),
		Label:       label,
		Description: description,
		Shortcut:    shortcut,
		Action: func() tea.Cmd {
			return func() tea.Msg { return SwitchTabMsg{Tab: tab} }
		},
	})
	return b
}

// AddRefresh adds a refresh command.
func (b *CommandBuilder) AddRefresh() *CommandBuilder {
	b.commands = append(b.commands, Command{
		ID:          "refresh",
		Label:       "Refresh",
		Description: "Refresh current view",
		Shortcut:    "r",
		Action: func() tea.Cmd {
			return func() tea.Msg { return RefreshCurrentTabMsg{} }
		},
	})
	return b
}

// AddTheme adds theme switch commands.
func (b *CommandBuilder) AddTheme() *CommandBuilder {
	themes := []string{"default", "minimal", "colorblind"}
	for _, theme := range themes {
		t := theme // capture loop variable
		b.commands = append(b.commands, Command{
			ID:          fmt.Sprintf("theme.%s", theme),
			Label:       fmt.Sprintf("Theme: %s", theme),
			Description: fmt.Sprintf("Switch to %s theme", theme),
			Action: func() tea.Cmd {
				return func() tea.Msg { return ChangeThemeMsg{Theme: t} }
			},
		})
	}
	return b
}

// AddQuit adds a quit command.
func (b *CommandBuilder) AddQuit() *CommandBuilder {
	b.commands = append(b.commands, Command{
		ID:          "quit",
		Label:       "Quit",
		Description: "Exit finterm",
		Shortcut:    "q / Ctrl+C",
		Action: func() tea.Cmd {
			return tea.Quit
		},
	})
	return b
}

// AddHelp adds a help command.
func (b *CommandBuilder) AddHelp() *CommandBuilder {
	b.commands = append(b.commands, Command{
		ID:          "help",
		Label:       "Help",
		Description: "Show key bindings",
		Shortcut:    "?",
		Action: func() tea.Cmd {
			return func() tea.Msg { return ShowHelpMsg{} }
		},
	})
	return b
}

// AddTickerCommands adds commands for each ticker in the watchlist.
func (b *CommandBuilder) AddTickerCommands(watchlist *config.WatchlistConfig) *CommandBuilder {
	for _, ticker := range watchlist.Equities {
		t := ticker // capture loop variable
		b.commands = append(b.commands, Command{
			ID:          fmt.Sprintf("quote.%s", t),
			Label:       fmt.Sprintf("Quote: %s", t),
			Description: fmt.Sprintf("Open Quote view with %s", t),
			Action: func() tea.Cmd {
				return func() tea.Msg { return OpenQuoteWithTickerMsg{Symbol: t} }
			},
		})
	}

	for _, ticker := range watchlist.Crypto {
		t := ticker // capture loop variable
		b.commands = append(b.commands, Command{
			ID:          fmt.Sprintf("quote.%s", t),
			Label:       fmt.Sprintf("Quote: %s", t),
			Description: fmt.Sprintf("Open Quote view with %s", t),
			Action: func() tea.Cmd {
				return func() tea.Msg { return OpenQuoteWithTickerMsg{Symbol: t} }
			},
		})
	}

	return b
}

// Build returns the built command list.
func (b *CommandBuilder) Build() []Command {
	return b.commands
}

// BuildDefaultCommands builds the default command set.
func BuildDefaultCommands(watchlist *config.WatchlistConfig) []Command {
	return NewBuilder().
		AddTab("trend", "Go to Trend", "Switch to Trend view", "1", 0).
		AddTab("quote", "Go to Quote", "Switch to Quote view", "2", 1).
		AddTab("macro", "Go to Macro", "Switch to Macro view", "3", 2).
		AddTab("news", "Go to News", "Switch to News view", "4", 3).
		AddRefresh().
		AddTheme().
		AddQuit().
		AddHelp().
		AddTickerCommands(watchlist).
		Build()
}

// SwitchTabMsg is a message to switch to a specific tab.
type SwitchTabMsg struct {
	Tab int
}

// RefreshCurrentTabMsg is a message to refresh the current tab.
type RefreshCurrentTabMsg struct{}

// ChangeThemeMsg is a message to change the theme.
type ChangeThemeMsg struct {
	Theme string
}

// ShowHelpMsg is a message to show the help overlay.
type ShowHelpMsg struct{}

// OpenQuoteWithTickerMsg is a message to open the Quote view with a specific ticker.
type OpenQuoteWithTickerMsg struct {
	Symbol string
}

// ToRefreshMsg converts RefreshCurrentTabMsg to the appropriate view-specific refresh message.
func ToRefreshMsg(activeTab int) tea.Msg {
	switch activeTab {
	case 0:
		return trend.RefreshMsg{}
	case 1:
		return quote.RefreshMsg{}
	case 2:
		return macro.RefreshMsg{}
	case 3:
		return news.RefreshMsg{}
	default:
		return nil
	}
}

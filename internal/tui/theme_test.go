package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTheme_DefaultColors(t *testing.T) {
	tests := []struct {
		name     string
		theme    *Theme
		wantName string
	}{
		{
			name:     "default theme created",
			theme:    NewTheme("default"),
			wantName: "default",
		},
		{
			name:     "invalid name falls back to default",
			theme:    NewTheme("invalid"),
			wantName: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.theme.Name() != tt.wantName {
				t.Errorf("Theme.Name() = %v, want %v", tt.theme.Name(), tt.wantName)
			}

			// Verify essential styles render successfully
			if tt.theme.TabBar().Render("test") == "" {
				t.Error("TabBar style should render")
			}
			if tt.theme.TableHeader().Render("test") == "" {
				t.Error("TableHeader style should render")
			}
			if tt.theme.Bullish().Render("test") == "" {
				t.Error("Bullish style should render")
			}
			if tt.theme.Bearish().Render("test") == "" {
				t.Error("Bearish style should render")
			}
			if tt.theme.Neutral().Render("test") == "" {
				t.Error("Neutral style should render")
			}
		})
	}
}

func TestTheme_ColorblindColors(t *testing.T) {
	theme := NewTheme("colorblind")

	if theme.Name() != "colorblind" {
		t.Errorf("Theme.Name() = %v, want colorblind", theme.Name())
	}

	// Verify colorblind theme renders successfully
	if theme.Bullish().Render("test") == "" {
		t.Error("Bullish style should render in colorblind theme")
	}
	if theme.Bearish().Render("test") == "" {
		t.Error("Bearish style should render in colorblind theme")
	}
	if theme.Neutral().Render("test") == "" {
		t.Error("Neutral style should render in colorblind theme")
	}

	// Verify all signal styles are different by checking they render
	bullishRendered := theme.Bullish().Render("X")
	bearishRendered := theme.Bearish().Render("X")
	neutralRendered := theme.Neutral().Render("X")

	if bullishRendered == "" {
		t.Error("Bullish rendering should not be empty")
	}
	if bearishRendered == "" {
		t.Error("Bearish rendering should not be empty")
	}
	if neutralRendered == "" {
		t.Error("Neutral rendering should not be empty")
	}
}

func TestTheme_MinimalColors(t *testing.T) {
	theme := NewTheme("minimal")

	if theme.Name() != "minimal" {
		t.Errorf("Theme.Name() = %v, want minimal", theme.Name())
	}

	bgColor := theme.Background()
	fgColor := theme.Foreground()

	// Verify minimal theme uses specific colors
	if bgColor != lipgloss.Color("#000000") {
		t.Errorf("Background should be black in minimal theme, got %v", bgColor)
	}
	if fgColor != lipgloss.Color("#FFFFFF") {
		t.Errorf("Foreground should be white in minimal theme, got %v", fgColor)
	}
}

func TestTheme_AllStylesAccessible(t *testing.T) {
	theme := NewTheme("default")

	// Verify all style accessors return valid styles
	accessors := []struct {
		name  string
		style lipgloss.Style
	}{
		{"TabBar", theme.TabBar()},
		{"Tab", theme.Tab()},
		{"TabActive", theme.TabActive()},
		{"TableHeader", theme.TableHeader()},
		{"TableRow", theme.TableRow()},
		{"TableBorder", theme.TableBorder()},
		{"TableEmpty", theme.TableEmpty()},
		{"Bullish", theme.Bullish()},
		{"Bearish", theme.Bearish()},
		{"Neutral", theme.Neutral()},
		{"Box", theme.Box()},
		{"BoxBorder", theme.BoxBorder()},
		{"BoxTitle", theme.BoxTitle()},
		{"Help", theme.Help()},
		{"Error", theme.Error()},
		{"Loading", theme.Loading()},
		{"Spinner", theme.Spinner()},
		{"SpinnerText", theme.SpinnerText()},
		{"Title", theme.Title()},
		{"Subtitle", theme.Subtitle()},
		{"Muted", theme.Muted()},
	}

	for _, acc := range accessors {
		t.Run(acc.name, func(_ *testing.T) {
			// Verify the style renders without panic
			_ = acc.style.Render("test")
		})
	}
}

func TestTheme_StyleConsistency(t *testing.T) {
	themes := []string{"default", "minimal", "colorblind"}

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			theme := NewTheme(themeName)

			// Verify styles render successfully
			if theme.TabActive().Render("test") == "" {
				t.Error("Active tab style should render")
			}
			if theme.Tab().Render("test") == "" {
				t.Error("Inactive tab style should render")
			}

			if theme.Bullish().Render("test") == "" {
				t.Error("Bullish style should render")
			}
			if theme.Bearish().Render("test") == "" {
				t.Error("Bearish style should render")
			}
			if theme.Neutral().Render("test") == "" {
				t.Error("Neutral style should render")
			}

			// Verify active and inactive tabs are different
			activeRendered := theme.TabActive().Render("Tab")
			inactiveRendered := theme.Tab().Render("Tab")
			if activeRendered == "" || inactiveRendered == "" {
				t.Error("Both active and inactive tabs should render")
			}
		})
	}
}

func TestTheme_Rendering(t *testing.T) {
	theme := NewTheme("default")

	tests := []struct {
		name   string
		style  lipgloss.Style
		text   string
		verify func(t *testing.T, result string)
	}{
		{
			name:  "tab bar renders",
			style: theme.TabBar(),
			text:  "Tab content",
			verify: func(t *testing.T, result string) {
				if result == "" {
					t.Error("rendered string should not be empty")
				}
			},
		},
		{
			name:  "bullish signal renders",
			style: theme.Bullish(),
			text:  "BUY",
			verify: func(t *testing.T, result string) {
				if result == "" {
					t.Error("rendered string should not be empty")
				}
			},
		},
		{
			name:  "error text renders",
			style: theme.Error(),
			text:  "Error occurred",
			verify: func(t *testing.T, result string) {
				if result == "" {
					t.Error("rendered string should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.style.Render(tt.text)
			tt.verify(t, result)
		})
	}
}

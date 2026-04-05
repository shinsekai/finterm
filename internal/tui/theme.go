// Package tui provides the terminal user interface for finterm.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palettes for different themes.
type palette struct {
	// Primary colors
	primary, primaryDark, secondary lipgloss.Color

	// Signal colors
	bullish, bearish, neutral lipgloss.Color

	// Semantic colors
	foreground, background, muted, border, header lipgloss.Color

	// Status colors
	error, warning, success, info lipgloss.Color

	// Connection state colors
	connOnline, connRateLimited, connOffline lipgloss.Color

	// Table colors
	tableHeader, tableBody, tableBorder, tableRowAlt lipgloss.Color
}

// defaultPalette provides the standard colorful theme.
var defaultPalette = palette{
	primary:         lipgloss.Color("#7D56F4"),
	primaryDark:     lipgloss.Color("#5C3FD6"),
	secondary:       lipgloss.Color("#FA7921"),
	bullish:         lipgloss.Color("#50FA7B"),
	bearish:         lipgloss.Color("#FF5555"),
	neutral:         lipgloss.Color("#F1FA8C"),
	foreground:      lipgloss.Color("#F8F8F2"),
	background:      lipgloss.Color("#282A36"),
	muted:           lipgloss.Color("#6272A4"),
	border:          lipgloss.Color("#44475A"),
	header:          lipgloss.Color("#BD93F9"),
	error:           lipgloss.Color("#FF5555"),
	warning:         lipgloss.Color("#F1FA8C"),
	success:         lipgloss.Color("#50FA7B"),
	info:            lipgloss.Color("#8BE9FD"),
	connOnline:      lipgloss.Color("#50FA7B"),
	connRateLimited: lipgloss.Color("#F1FA8C"),
	connOffline:     lipgloss.Color("#FF5555"),
	tableHeader:     lipgloss.Color("#BD93F9"),
	tableBody:       lipgloss.Color("#F8F8F2"),
	tableBorder:     lipgloss.Color("#44475A"),
	tableRowAlt:     lipgloss.Color("#2D2F3D"),
}

// minimalPalette provides a minimal black-and-white theme.
var minimalPalette = palette{
	primary:         lipgloss.Color("#FFFFFF"),
	primaryDark:     lipgloss.Color("#CCCCCC"),
	secondary:       lipgloss.Color("#AAAAAA"),
	bullish:         lipgloss.Color("#FFFFFF"),
	bearish:         lipgloss.Color("#AAAAAA"),
	neutral:         lipgloss.Color("#666666"),
	foreground:      lipgloss.Color("#FFFFFF"),
	background:      lipgloss.Color("#000000"),
	muted:           lipgloss.Color("#888888"),
	border:          lipgloss.Color("#333333"),
	header:          lipgloss.Color("#FFFFFF"),
	error:           lipgloss.Color("#FFFFFF"),
	warning:         lipgloss.Color("#AAAAAA"),
	success:         lipgloss.Color("#FFFFFF"),
	info:            lipgloss.Color("#AAAAAA"),
	connOnline:      lipgloss.Color("#FFFFFF"),
	connRateLimited: lipgloss.Color("#AAAAAA"),
	connOffline:     lipgloss.Color("#888888"),
	tableHeader:     lipgloss.Color("#FFFFFF"),
	tableBody:       lipgloss.Color("#FFFFFF"),
	tableBorder:     lipgloss.Color("#333333"),
	tableRowAlt:     lipgloss.Color("#111111"),
}

// colorblindPalette provides a theme optimized for colorblind users.
var colorblindPalette = palette{
	primary:         lipgloss.Color("#4D90FE"),
	primaryDark:     lipgloss.Color("#3367D6"),
	secondary:       lipgloss.Color("#F09300"),
	bullish:         lipgloss.Color("#00C853"),
	bearish:         lipgloss.Color("#D50000"),
	neutral:         lipgloss.Color("#FFB300"),
	foreground:      lipgloss.Color("#202124"),
	background:      lipgloss.Color("#FFFFFF"),
	muted:           lipgloss.Color("#5F6368"),
	border:          lipgloss.Color("#DADCE0"),
	header:          lipgloss.Color("#4D90FE"),
	error:           lipgloss.Color("#D50000"),
	warning:         lipgloss.Color("#FFB300"),
	success:         lipgloss.Color("#00C853"),
	info:            lipgloss.Color("#4D90FE"),
	connOnline:      lipgloss.Color("#00C853"),
	connRateLimited: lipgloss.Color("#FFB300"),
	connOffline:     lipgloss.Color("#D50000"),
	tableHeader:     lipgloss.Color("#4D90FE"),
	tableBody:       lipgloss.Color("#202124"),
	tableBorder:     lipgloss.Color("#DADCE0"),
	tableRowAlt:     lipgloss.Color("#F0F0F0"),
}

// Style definitions for UI elements.
type styles struct {
	// Tab bar styles
	tabBar, tab, tabActive, tabIcon lipgloss.Style

	// Table styles
	tableHeader, tableRow, tableRowAlt, tableBorder, tableEmpty lipgloss.Style

	// Signal styles
	bullish, bearish, neutral                lipgloss.Style
	bullishBadge, bearishBadge, neutralBadge lipgloss.Style

	// Card styles
	card, cardHeader, cardHeaderAccent lipgloss.Style

	// Box and border styles
	box, boxBorder, boxTitle lipgloss.Style

	// Text styles
	help, error, loading, warning lipgloss.Style

	// Status bar styles
	statusBar, statusBarLeft, statusBarRight lipgloss.Style

	// Section styles
	sectionHeader, divider lipgloss.Style

	// Accent and input styles
	accent                 lipgloss.Style
	inputField, inputLabel lipgloss.Style
	metaLabel, metaValue   lipgloss.Style

	// Spinner styles
	spinner, spinnerText lipgloss.Style

	// Connection state styles
	statusOnline, statusRateLimited, statusOffline lipgloss.Style

	// General styles
	title, subtitle, muted lipgloss.Style
}

// Theme encapsulates the visual appearance of the TUI.
type Theme struct {
	palette palette
	styles  styles
	name    string
}

// NewTheme creates a new theme based on the given name.
// Valid names: "default", "minimal", "colorblind".
func NewTheme(name string) *Theme {
	var p palette
	themeName := name

	switch name {
	case "minimal":
		p = minimalPalette
	case "colorblind":
		p = colorblindPalette
	default:
		p = defaultPalette
		themeName = "default"
	}

	return newThemeFromPalette(themeName, p)
}

func newThemeFromPalette(name string, p palette) *Theme {
	t := &Theme{
		palette: p,
		name:    name,
	}

	t.buildStyles()
	return t
}

func (t *Theme) buildStyles() {
	base := lipgloss.NewStyle().
		Foreground(t.palette.foreground).
		Background(t.palette.background)

	// Tab bar styles
	t.styles.tabBar = base.
		Background(t.palette.background).
		Border(lipgloss.Border{Top: "─", Bottom: "─"}).
		BorderForeground(t.palette.border).
		Padding(0, 1)

	t.styles.tab = base.
		Foreground(t.palette.muted).
		Background(t.palette.background).
		Padding(0, 2)

	t.styles.tabActive = base.
		Foreground(t.palette.foreground).
		Background(t.palette.primary).
		Bold(true).
		Padding(0, 2)

	t.styles.tabIcon = base.
		Foreground(t.palette.primary).
		Bold(true)

	// Table styles
	t.styles.tableHeader = lipgloss.NewStyle().
		Foreground(t.palette.tableHeader).
		Bold(true).
		BorderBottom(true).
		BorderForeground(t.palette.tableBorder).
		Padding(0, 1)

	t.styles.tableRow = lipgloss.NewStyle().
		Foreground(t.palette.tableBody).
		Background(t.palette.background).
		Padding(0, 1)

	t.styles.tableRowAlt = lipgloss.NewStyle().
		Foreground(t.palette.tableBody).
		Background(t.palette.tableRowAlt).
		Padding(0, 1)

	t.styles.tableBorder = lipgloss.NewStyle().
		Foreground(t.palette.tableBorder).
		BorderStyle(lipgloss.RoundedBorder())

	t.styles.tableEmpty = lipgloss.NewStyle().
		Foreground(t.palette.muted).
		Italic(true)

	// Signal styles
	t.styles.bullish = lipgloss.NewStyle().
		Foreground(t.palette.bullish).
		Bold(true)

	t.styles.bearish = lipgloss.NewStyle().
		Foreground(t.palette.bearish).
		Bold(true)

	t.styles.neutral = lipgloss.NewStyle().
		Foreground(t.palette.neutral).
		Bold(true)

	// Signal badge styles
	t.styles.bullishBadge = lipgloss.NewStyle().
		Foreground(t.palette.background).
		Background(t.palette.bullish).
		Bold(true).
		Padding(0, 1)

	t.styles.bearishBadge = lipgloss.NewStyle().
		Foreground(t.palette.foreground).
		Background(t.palette.bearish).
		Bold(true).
		Padding(0, 1)

	t.styles.neutralBadge = lipgloss.NewStyle().
		Foreground(t.palette.background).
		Background(t.palette.neutral).
		Bold(true).
		Padding(0, 1)

	// Card styles
	t.styles.card = base.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.palette.border).
		Padding(1, 2)

	t.styles.cardHeader = lipgloss.NewStyle().
		Foreground(t.palette.foreground).
		Bold(true).
		MarginBottom(1)

	t.styles.cardHeaderAccent = lipgloss.NewStyle().
		Foreground(t.palette.primary).
		Bold(true)

	// Box styles (switch to RoundedBorder)
	t.styles.box = base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.palette.border).
		Padding(1)

	t.styles.boxBorder = lipgloss.NewStyle().
		Foreground(t.palette.border).
		BorderStyle(lipgloss.RoundedBorder())

	t.styles.boxTitle = lipgloss.NewStyle().
		Foreground(t.palette.header).
		Bold(true)

	// Text styles
	t.styles.help = lipgloss.NewStyle().
		Foreground(t.palette.muted).
		Italic(true)

	t.styles.error = lipgloss.NewStyle().
		Foreground(t.palette.error).
		Bold(true)

	t.styles.loading = lipgloss.NewStyle().
		Foreground(t.palette.info)

	t.styles.warning = lipgloss.NewStyle().
		Foreground(t.palette.warning).
		Bold(true)

	// Status bar styles
	t.styles.statusBar = base.
		Background(t.palette.muted).
		Foreground(t.palette.foreground).
		Padding(0, 1)

	t.styles.statusBarLeft = lipgloss.NewStyle().
		Foreground(t.palette.foreground)

	t.styles.statusBarRight = lipgloss.NewStyle().
		Foreground(t.palette.foreground)

	// Section styles
	t.styles.sectionHeader = lipgloss.NewStyle().
		Foreground(t.palette.header).
		Bold(true).
		Border(lipgloss.Border{Bottom: "─"}).
		BorderForeground(t.palette.border).
		Padding(0, 1)

	t.styles.divider = lipgloss.NewStyle().
		Foreground(t.palette.border).
		SetString("─")

	// Accent and input styles
	t.styles.accent = lipgloss.NewStyle().
		Foreground(t.palette.primary).
		Bold(true)

	t.styles.inputField = lipgloss.NewStyle().
		Foreground(t.palette.foreground).
		Background(t.palette.background).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.palette.primary).
		Padding(0, 1)

	t.styles.inputLabel = lipgloss.NewStyle().
		Foreground(t.palette.muted).
		MarginRight(1)

	t.styles.metaLabel = lipgloss.NewStyle().
		Foreground(t.palette.muted).
		Width(16)

	t.styles.metaValue = lipgloss.NewStyle().
		Foreground(t.palette.foreground).
		Bold(true)

	// Spinner styles
	t.styles.spinner = lipgloss.NewStyle().
		Foreground(t.palette.primary)

	t.styles.spinnerText = lipgloss.NewStyle().
		Foreground(t.palette.muted)

	// Connection state styles
	t.styles.statusOnline = lipgloss.NewStyle().
		Foreground(t.palette.connOnline).
		Bold(true)

	t.styles.statusRateLimited = lipgloss.NewStyle().
		Foreground(t.palette.connRateLimited).
		Bold(true)

	t.styles.statusOffline = lipgloss.NewStyle().
		Foreground(t.palette.connOffline).
		Bold(true)

	// General styles
	t.styles.title = lipgloss.NewStyle().
		Foreground(t.palette.header).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)

	t.styles.subtitle = lipgloss.NewStyle().
		Foreground(t.palette.foreground).
		Bold(true)

	t.styles.muted = lipgloss.NewStyle().
		Foreground(t.palette.muted)
}

// Name returns the theme name.
func (t *Theme) Name() string {
	return t.name
}

// TabBar returns the tab bar style.
func (t *Theme) TabBar() lipgloss.Style {
	return t.styles.tabBar
}

// Tab returns the tab style for inactive tabs.
func (t *Theme) Tab() lipgloss.Style {
	return t.styles.tab
}

// TabActive returns the tab style for the active tab.
func (t *Theme) TabActive() lipgloss.Style {
	return t.styles.tabActive
}

// TabIcon returns the tab icon style.
func (t *Theme) TabIcon() lipgloss.Style {
	return t.styles.tabIcon
}

// TableHeader returns the table header style.
func (t *Theme) TableHeader() lipgloss.Style {
	return t.styles.tableHeader
}

// TableRow returns the table row style.
func (t *Theme) TableRow() lipgloss.Style {
	return t.styles.tableRow
}

// TableRowAlt returns the alternating table row style.
func (t *Theme) TableRowAlt() lipgloss.Style {
	return t.styles.tableRowAlt
}

// TableBorder returns the table border style.
func (t *Theme) TableBorder() lipgloss.Style {
	return t.styles.tableBorder
}

// TableEmpty returns the empty table message style.
func (t *Theme) TableEmpty() lipgloss.Style {
	return t.styles.tableEmpty
}

// Bullish returns the bullish signal style.
func (t *Theme) Bullish() lipgloss.Style {
	return t.styles.bullish
}

// Bearish returns the bearish signal style.
func (t *Theme) Bearish() lipgloss.Style {
	return t.styles.bearish
}

// Neutral returns the neutral signal style.
func (t *Theme) Neutral() lipgloss.Style {
	return t.styles.neutral
}

// BullishBadge returns the bullish badge style with background fill.
func (t *Theme) BullishBadge() lipgloss.Style {
	return t.styles.bullishBadge
}

// BearishBadge returns the bearish badge style with background fill.
func (t *Theme) BearishBadge() lipgloss.Style {
	return t.styles.bearishBadge
}

// NeutralBadge returns the neutral badge style with background fill.
func (t *Theme) NeutralBadge() lipgloss.Style {
	return t.styles.neutralBadge
}

// Box returns the box style.
func (t *Theme) Box() lipgloss.Style {
	return t.styles.box
}

// BoxBorder returns the box border style.
func (t *Theme) BoxBorder() lipgloss.Style {
	return t.styles.boxBorder
}

// BoxTitle returns the box title style.
func (t *Theme) BoxTitle() lipgloss.Style {
	return t.styles.boxTitle
}

// Card returns the card style with rounded border.
func (t *Theme) Card() lipgloss.Style {
	return t.styles.card
}

// CardHeader returns the card header style.
func (t *Theme) CardHeader() lipgloss.Style {
	return t.styles.cardHeader
}

// CardHeaderAccent returns the card header accent style.
func (t *Theme) CardHeaderAccent() lipgloss.Style {
	return t.styles.cardHeaderAccent
}

// Help returns the help text style.
func (t *Theme) Help() lipgloss.Style {
	return t.styles.help
}

// Error returns the error text style.
func (t *Theme) Error() lipgloss.Style {
	return t.styles.error
}

// Loading returns the loading text style.
func (t *Theme) Loading() lipgloss.Style {
	return t.styles.loading
}

// Warning returns the warning text style.
func (t *Theme) Warning() lipgloss.Style {
	return t.styles.warning
}

// Spinner returns the spinner frame style.
func (t *Theme) Spinner() lipgloss.Style {
	return t.styles.spinner
}

// SpinnerText returns the spinner text style.
func (t *Theme) SpinnerText() lipgloss.Style {
	return t.styles.spinnerText
}

// Title returns the title style.
func (t *Theme) Title() lipgloss.Style {
	return t.styles.title
}

// Subtitle returns the subtitle style.
func (t *Theme) Subtitle() lipgloss.Style {
	return t.styles.subtitle
}

// Muted returns the muted text style.
func (t *Theme) Muted() lipgloss.Style {
	return t.styles.muted
}

// StatusOnline returns the online status style.
func (t *Theme) StatusOnline() lipgloss.Style {
	return t.styles.statusOnline
}

// StatusRateLimited returns the rate limited status style.
func (t *Theme) StatusRateLimited() lipgloss.Style {
	return t.styles.statusRateLimited
}

// StatusOffline returns the offline status style.
func (t *Theme) StatusOffline() lipgloss.Style {
	return t.styles.statusOffline
}

// StatusBar returns the status bar style.
func (t *Theme) StatusBar() lipgloss.Style {
	return t.styles.statusBar
}

// StatusBarLeft returns the status bar left section style.
func (t *Theme) StatusBarLeft() lipgloss.Style {
	return t.styles.statusBarLeft
}

// StatusBarRight returns the status bar right section style.
func (t *Theme) StatusBarRight() lipgloss.Style {
	return t.styles.statusBarRight
}

// SectionHeader returns the section header style.
func (t *Theme) SectionHeader() lipgloss.Style {
	return t.styles.sectionHeader
}

// Divider returns the divider style.
func (t *Theme) Divider() lipgloss.Style {
	return t.styles.divider
}

// Accent returns the accent style (primary color, bold).
func (t *Theme) Accent() lipgloss.Style {
	return t.styles.accent
}

// InputField returns the input field style.
func (t *Theme) InputField() lipgloss.Style {
	return t.styles.inputField
}

// InputLabel returns the input label style.
func (t *Theme) InputLabel() lipgloss.Style {
	return t.styles.inputLabel
}

// MetaLabel returns the meta label style for label-value pairs.
func (t *Theme) MetaLabel() lipgloss.Style {
	return t.styles.metaLabel
}

// MetaValue returns the meta value style for label-value pairs.
func (t *Theme) MetaValue() lipgloss.Style {
	return t.styles.metaValue
}

// Primary returns the primary color.
func (t *Theme) Primary() lipgloss.Color {
	return t.palette.primary
}

// Border returns the border color.
func (t *Theme) Border() lipgloss.Color {
	return t.palette.border
}

// Foreground returns the foreground color.
func (t *Theme) Foreground() lipgloss.Color {
	return t.palette.foreground
}

// Background returns the background color.
func (t *Theme) Background() lipgloss.Color {
	return t.palette.background
}

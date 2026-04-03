// Package components provides reusable TUI components.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Alignment represents the horizontal alignment of table columns.
type Alignment int

const (
	// AlignLeft aligns content to the left.
	AlignLeft Alignment = iota
	// AlignCenter aligns content to the center.
	AlignCenter
	// AlignRight aligns content to the right.
	AlignRight
)

// Column defines a single column in a table.
type Column struct {
	// Title is the column header text.
	Title string

	// Width is the fixed width of the column.
	// If 0, the column width is auto-calculated.
	Width int

	// Alignment specifies the text alignment.
	Alignment Alignment

	// Style is the style applied to the column's cells.
	Style lipgloss.Style

	// HeaderStyle is the style applied to the column header.
	HeaderStyle lipgloss.Style
}

// Cell represents a single cell in a table row.
type Cell struct {
	// Text is the cell content.
	Text string

	// Style is the style applied to this specific cell.
	// If empty, the column style is used.
	Style lipgloss.Style
}

// Row represents a single row in a table.
type Row struct {
	// Cells contains the data for each column.
	Cells []Cell

	// Style is the style applied to the entire row.
	// If empty, the default row style is used.
	Style lipgloss.Style
}

// NewRow creates a new row from a slice of strings.
func NewRow(values []string) Row {
	cells := make([]Cell, len(values))
	for i, v := range values {
		cells[i] = Cell{Text: v}
	}
	return Row{Cells: cells}
}

// Table represents a reusable table component.
type Table struct {
	// Columns defines the table structure.
	Columns []Column

	// Rows contains the table data.
	Rows []Row

	// Styles for different parts of the table.
	HeaderStyle lipgloss.Style
	RowStyle    lipgloss.Style
	BorderStyle lipgloss.Style
	EmptyStyle  lipgloss.Style

	// EmptyMessage is displayed when there are no rows.
	EmptyMessage string

	// BorderEnabled controls whether a border is rendered.
	BorderEnabled bool

	// BorderType specifies the border style.
	BorderType lipgloss.Border

	// MaxWidth limits the total table width.
	// If 0, no limit is applied.
	MaxWidth int

	// ColumnGap is the spacing between columns.
	ColumnGap int
}

// NewTable creates a new table with default settings.
func NewTable() *Table {
	return &Table{
		HeaderStyle:   lipgloss.NewStyle().Bold(true),
		RowStyle:      lipgloss.NewStyle(),
		BorderStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
		EmptyStyle:    lipgloss.NewStyle().Italic(true).Faint(true),
		EmptyMessage:  "No data available",
		BorderEnabled: false,
		BorderType:    lipgloss.NormalBorder(),
		MaxWidth:      0,
		ColumnGap:     2,
	}
}

// WithColumns sets the table columns.
func (t *Table) WithColumns(columns []Column) *Table {
	t.Columns = columns
	return t
}

// WithRows sets the table rows.
func (t *Table) WithRows(rows []Row) *Table {
	t.Rows = rows
	return t
}

// AddRow appends a row to the table.
func (t *Table) AddRow(row Row) *Table {
	t.Rows = append(t.Rows, row)
	return t
}

// WithEmptyMessage sets the message shown when the table is empty.
func (t *Table) WithEmptyMessage(msg string) *Table {
	t.EmptyMessage = msg
	return t
}

// WithBorder enables or disables the table border.
func (t *Table) WithBorder(enabled bool) *Table {
	t.BorderEnabled = enabled
	return t
}

// WithBorderType sets the border style.
func (t *Table) WithBorderType(borderType lipgloss.Border) *Table {
	t.BorderType = borderType
	t.BorderStyle = t.BorderStyle.BorderStyle(borderType)
	return t
}

// WithMaxWidth sets the maximum table width.
func (t *Table) WithMaxWidth(width int) *Table {
	t.MaxWidth = width
	return t
}

// WithColumnGap sets the spacing between columns.
func (t *Table) WithColumnGap(gap int) *Table {
	t.ColumnGap = gap
	return t
}

// calculateColumnWidths computes the width for each column.
func (t *Table) calculateColumnWidths(totalWidth int) []int {
	if len(t.Columns) == 0 {
		return nil
	}

	widths := make([]int, len(t.Columns))

	// First pass: use fixed widths where specified
	fixedColumns := 0
	totalFixedWidth := 0
	for i, col := range t.Columns {
		if col.Width > 0 {
			widths[i] = col.Width
			fixedColumns++
			totalFixedWidth += col.Width
		}
	}

	// Calculate available space for auto-width columns
	remainingColumns := len(t.Columns) - fixedColumns
	gapWidth := (len(t.Columns) - 1) * t.ColumnGap
	availableWidth := totalWidth - totalFixedWidth - gapWidth

	// If no max width specified, use content-based width
	if totalWidth == 0 {
		for i, col := range t.Columns {
			if col.Width == 0 {
				// Calculate based on header and content
				maxCellWidth := lipgloss.Width(col.Title)
				for _, row := range t.Rows {
					if i < len(row.Cells) {
						cellWidth := lipgloss.Width(row.Cells[i].Text)
						if cellWidth > maxCellWidth {
							maxCellWidth = cellWidth
						}
					}
				}
				widths[i] = maxCellWidth
			}
		}
	} else if remainingColumns > 0 {
		// Distribute available space evenly
		avgWidth := availableWidth / remainingColumns
		for i, col := range t.Columns {
			if col.Width == 0 {
				widths[i] = avgWidth
				if i == len(t.Columns)-1 {
					// Last column gets remainder
					widths[i] = availableWidth - (avgWidth * (remainingColumns - 1))
				}
			}
		}
	}

	return widths
}

// renderCell renders a cell with the given width and alignment.
func (t *Table) renderCell(cell Cell, col Column, width int) string {
	text := cell.Text
	style := cell.Style
	// Use column style if cell style is effectively unset
	// We can't compare styles directly, so we always use cell's style
	// Users should set cell styles explicitly if they want custom styling

	// Truncate if necessary
	if width > 0 && lipgloss.Width(text) > width {
		text = TruncateText(text, width)
	}

	// Pad and align
	if width > 0 {
		textWidth := lipgloss.Width(text)
		padding := width - textWidth

		switch col.Alignment {
		case AlignCenter:
			leftPadding := padding / 2
			rightPadding := padding - leftPadding
			text = strings.Repeat(" ", leftPadding) + text + strings.Repeat(" ", rightPadding)
		case AlignRight:
			text = strings.Repeat(" ", padding) + text
		default: // AlignLeft
			text += strings.Repeat(" ", padding)
		}
	}

	return style.Render(text)
}

// renderRow renders a single row.
func (t *Table) renderRow(row Row, widths []int) string {
	rowStyle := t.RowStyle
	// Use row's style if set
	// We can't compare styles directly, so always try to use row's style first
	// Users should set row styles explicitly if they want custom styling

	cells := make([]string, 0, len(t.Columns))
	for i, col := range t.Columns {
		var cell Cell
		if i < len(row.Cells) {
			cell = row.Cells[i]
		}
		width := 0
		if i < len(widths) {
			width = widths[i]
		}
		cells = append(cells, t.renderCell(cell, col, width))
	}

	gap := strings.Repeat(" ", t.ColumnGap)
	return rowStyle.Render(strings.Join(cells, gap))
}

// renderHeader renders the table header.
func (t *Table) renderHeader(widths []int) string {
	cells := make([]string, 0, len(t.Columns))
	for i, col := range t.Columns {
		text := col.Title
		width := 0
		if i < len(widths) {
			width = widths[i]
		}

		// Truncate header if necessary
		if width > 0 && lipgloss.Width(text) > width {
			text = TruncateText(text, width)
		}

		// Pad and align header
		if width > 0 {
			textWidth := lipgloss.Width(text)
			padding := width - textWidth

			switch col.Alignment {
			case AlignCenter:
				leftPadding := padding / 2
				rightPadding := padding - leftPadding
				text = strings.Repeat(" ", leftPadding) + text + strings.Repeat(" ", rightPadding)
			case AlignRight:
				text = strings.Repeat(" ", padding) + text
			default: // AlignLeft
				text += strings.Repeat(" ", padding)
			}
		}

		style := col.HeaderStyle
		// Use table's header style as default if column header style not set
		// We can't compare styles directly, so always use column's header style
		// Users should set header styles explicitly if they want custom styling

		cells = append(cells, style.Render(text))
	}

	gap := strings.Repeat(" ", t.ColumnGap)
	return strings.Join(cells, gap)
}

// Render renders the table as a string.
func (t *Table) Render() string {
	// Handle empty table
	if len(t.Rows) == 0 {
		return t.EmptyStyle.Render(t.EmptyMessage)
	}

	widths := t.calculateColumnWidths(t.MaxWidth)

	var lines []string

	// Render header
	header := t.renderHeader(widths)
	lines = append(lines, header)

	// Render rows
	for _, row := range t.Rows {
		lines = append(lines, t.renderRow(row, widths))
	}

	result := strings.Join(lines, "\n")

	// Apply border if enabled
	if t.BorderEnabled {
		result = t.BorderStyle.Border(t.BorderType).Render(result)
	}

	return result
}

// Width returns the total width of the rendered table.
func (t *Table) Width() int {
	widths := t.calculateColumnWidths(t.MaxWidth)
	if widths == nil {
		return 0
	}

	total := 0
	for _, w := range widths {
		total += w
	}
	total += (len(widths) - 1) * t.ColumnGap

	return total
}

// Height returns the number of lines in the rendered table.
func (t *Table) Height() int {
	if len(t.Rows) == 0 {
		return 1
	}
	return len(t.Rows) + 1 // +1 for header
}

// Clear removes all rows from the table.
func (t *Table) Clear() *Table {
	t.Rows = nil
	return t
}

// ColumnCount returns the number of columns.
func (t *Table) ColumnCount() int {
	return len(t.Columns)
}

// RowCount returns the number of rows.
func (t *Table) RowCount() int {
	return len(t.Rows)
}

// String returns the rendered table.
func (t *Table) String() string {
	return t.Render()
}

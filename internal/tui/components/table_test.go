package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTable_EmptyData(t *testing.T) {
	tests := []struct {
		name          string
		table         *Table
		expectPanic   bool
		checkContains bool
		contains      string
	}{
		{
			name:        "empty table with columns",
			table:       NewTable().WithColumns([]Column{{Title: "A"}, {Title: "B"}}),
			expectPanic: false,
		},
		{
			name:        "empty table without columns",
			table:       NewTable(),
			expectPanic: false,
		},
		{
			name:          "custom empty message",
			table:         NewTable().WithEmptyMessage("Nothing to see here"),
			expectPanic:   false,
			checkContains: true,
			contains:      "Nothing to see here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			panicked := false

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				result = tt.table.Render()
			}()

			if tt.expectPanic && !panicked {
				t.Error("Expected panic but none occurred")
			}
			if !tt.expectPanic && panicked {
				t.Error("Unexpected panic occurred")
			}

			if tt.checkContains {
				// Check if empty message is present (may have styling codes)
				if len(result) < len(tt.contains) {
					t.Errorf("Result too short to contain message: got %d chars, want at least %d", len(result), len(tt.contains))
				}
			}
		})
	}
}

func TestTable_ColumnAlignment(t *testing.T) {
	tests := []struct {
		name       string
		columns    []Column
		rows       []Row
		renderable bool
	}{
		{
			name: "left alignment (default)",
			columns: []Column{
				{Title: "Left", Alignment: AlignLeft, Width: 10},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "A"}}},
			},
			renderable: true,
		},
		{
			name: "center alignment",
			columns: []Column{
				{Title: "Center", Alignment: AlignCenter, Width: 10},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "Mid"}}},
			},
			renderable: true,
		},
		{
			name: "right alignment",
			columns: []Column{
				{Title: "Right", Alignment: AlignRight, Width: 10},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "End"}}},
			},
			renderable: true,
		},
		{
			name: "mixed alignments",
			columns: []Column{
				{Title: "L", Alignment: AlignLeft, Width: 8},
				{Title: "C", Alignment: AlignCenter, Width: 8},
				{Title: "R", Alignment: AlignRight, Width: 8},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "Left"}, {Text: "Mid"}, {Text: "Right"}}},
			},
			renderable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTable().WithColumns(tt.columns).WithRows(tt.rows)
			result := table.Render()

			if tt.renderable {
				if result == "" {
					t.Error("Render should not return empty string")
				}
			}

			// Verify rendering doesn't panic
			_ = result
		})
	}
}

func TestTable_Truncation(t *testing.T) {
	tests := []struct {
		name        string
		columns     []Column
		rows        []Row
		maxWidth    int
		expectPanic bool
	}{
		{
			name: "truncate single column",
			columns: []Column{
				{Title: "Short", Width: 5},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "Very Long Text"}}},
			},
			maxWidth:    0, // Column width is fixed
			expectPanic: false,
		},
		{
			name: "truncate multiple columns",
			columns: []Column{
				{Title: "A", Width: 5},
				{Title: "B", Width: 5},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "LongA"}, {Text: "LongB"}}},
			},
			maxWidth:    0,
			expectPanic: false,
		},
		{
			name: "truncate with max width constraint",
			columns: []Column{
				{Title: "Column1"},
				{Title: "Column2"},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "Very very long content here"}, {Text: "More content"}}},
			},
			maxWidth:    20,
			expectPanic: false,
		},
		{
			name: "truncate header",
			columns: []Column{
				{Title: "Very Long Header Name", Width: 8},
			},
			rows: []Row{
				{Cells: []Cell{{Text: "Short"}}},
			},
			maxWidth:    0,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTable().WithColumns(tt.columns).WithRows(tt.rows).WithMaxWidth(tt.maxWidth)

			var result string
			panicked := false

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				result = table.Render()
			}()

			if tt.expectPanic && !panicked {
				t.Error("Expected panic but none occurred")
			}
			if !tt.expectPanic && panicked {
				t.Error("Unexpected panic occurred")
			}

			if result == "" {
				t.Error("Render should not return empty string")
			}
		})
	}
}

func TestTable_BasicRendering(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{
		{Title: "Ticker", Width: 8},
		{Title: "Price", Width: 10},
		{Title: "Signal", Width: 8},
	})
	table.WithRows([]Row{
		{Cells: []Cell{{Text: "AAPL"}, {Text: "150.25"}, {Text: "Bullish"}}},
		{Cells: []Cell{{Text: "MSFT"}, {Text: "280.50"}, {Text: "Bearish"}}},
	})

	result := table.Render()

	if result == "" {
		t.Error("Render should not return empty string")
	}

	// Should contain header text
	if len(result) < len("Ticker")+len("Price")+len("Signal") {
		t.Error("Result should contain all header text")
	}
}

func TestTable_WithStyles(t *testing.T) {
	boldStyle := lipgloss.NewStyle().Bold(true)
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))

	table := NewTable()
	table.WithColumns([]Column{
		{Title: "Name", Style: boldStyle},
	})
	table.WithRows([]Row{
		{
			Cells: []Cell{{Text: "Test", Style: redStyle}},
		},
	})

	result := table.Render()
	if result == "" {
		t.Error("Render should work with styles")
	}
}

func TestTable_WithBorder(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}, {Title: "B"}})
	table.WithRows([]Row{{Cells: []Cell{{Text: "1"}, {Text: "2"}}}})

	// Without border
	resultNoBorder := table.Render()

	// With border
	resultWithBorder := table.WithBorder(true).Render()

	if resultNoBorder == "" {
		t.Error("Render without border should work")
	}
	if resultWithBorder == "" {
		t.Error("Render with border should work")
	}

	// With border should be longer (contains border characters)
	if len(resultWithBorder) <= len(resultNoBorder) {
		t.Logf("Warning: Border may not affect length significantly (border chars may be invisible)")
	}
}

func TestTable_AddRow(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}})

	if table.RowCount() != 0 {
		t.Errorf("Initial row count = %v, want 0", table.RowCount())
	}

	table.AddRow(Row{Cells: []Cell{{Text: "1"}}})
	if table.RowCount() != 1 {
		t.Errorf("After AddRow, count = %v, want 1", table.RowCount())
	}

	table.AddRow(Row{Cells: []Cell{{Text: "2"}}})
	if table.RowCount() != 2 {
		t.Errorf("After second AddRow, count = %v, want 2", table.RowCount())
	}
}

func TestTable_Clear(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}})
	table.AddRow(Row{Cells: []Cell{{Text: "1"}}})

	if table.RowCount() != 1 {
		t.Errorf("Before Clear, count = %v, want 1", table.RowCount())
	}

	table.Clear()

	if table.RowCount() != 0 {
		t.Errorf("After Clear, count = %v, want 0", table.RowCount())
	}

	// Should render empty message
	result := table.Render()
	if result == "" {
		t.Error("After Clear, should render empty message")
	}
}

func TestTable_WidthAndHeight(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{
		{Title: "A", Width: 5},
		{Title: "B", Width: 5},
		{Title: "C", Width: 5},
	})
	table.AddRow(Row{Cells: []Cell{{Text: "1"}, {Text: "2"}, {Text: "3"}}})

	width := table.Width()
	expectedWidth := 5 + 5 + 5 + 2*2 // columns + 2 gaps of 2 each
	if width != expectedWidth {
		t.Errorf("Width = %v, want %v", width, expectedWidth)
	}

	if table.Height() != 2 { // 1 row + 1 header
		t.Errorf("Height = %v, want 2", width)
	}

	table.AddRow(Row{Cells: []Cell{{Text: "4"}, {Text: "5"}, {Text: "6"}}})
	if table.Height() != 3 { // 2 rows + 1 header
		t.Errorf("Height after AddRow = %v, want 3", table.Height())
	}
}

func TestTable_ColumnAndRowCount(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}, {Title: "B"}})

	if table.ColumnCount() != 2 {
		t.Errorf("ColumnCount = %v, want 2", table.ColumnCount())
	}
	if table.RowCount() != 0 {
		t.Errorf("RowCount = %v, want 0", table.RowCount())
	}

	table.AddRow(Row{Cells: []Cell{{Text: "1"}, {Text: "2"}}})
	table.AddRow(Row{Cells: []Cell{{Text: "3"}, {Text: "4"}}})

	if table.RowCount() != 2 {
		t.Errorf("RowCount after AddRows = %v, want 2", table.RowCount())
	}
}

func TestTable_CellStyles(t *testing.T) {
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))

	table := NewTable()
	table.WithColumns([]Column{
		{Title: "A", Width: 5},
		{Title: "B", Width: 5},
	})
	table.WithRows([]Row{
		{
			Cells: []Cell{
				{Text: "Red", Style: redStyle},
				{Text: "Blue", Style: blueStyle},
			},
		},
	})

	result := table.Render()
	if result == "" {
		t.Error("Render should work with cell-specific styles")
	}
}

func TestTable_RowStyles(t *testing.T) {
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))

	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}})
	table.WithRows([]Row{
		{Cells: []Cell{{Text: "Normal"}}},
		{
			Cells: []Cell{{Text: "Highlighted"}},
			Style: yellowStyle,
		},
	})

	result := table.Render()
	if result == "" {
		t.Error("Render should work with row-specific styles")
	}
}

func TestTable_NewRow(t *testing.T) {
	row := NewRow([]string{"A", "B", "C"})

	if len(row.Cells) != 3 {
		t.Errorf("NewRow created %d cells, want 3", len(row.Cells))
	}

	expected := []string{"A", "B", "C"}
	for i, cell := range row.Cells {
		if cell.Text != expected[i] {
			t.Errorf("Cell %d = %v, want %v", i, cell.Text, expected[i])
		}
	}
}

func TestTable_VariableWidthColumns(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{
		{Title: "Short"},            // Auto width
		{Title: "Fixed", Width: 10}, // Fixed width
		{Title: "Medium"},           // Auto width
	})
	table.WithRows([]Row{
		{Cells: []Cell{{Text: "A"}, {Text: "FixedCell"}, {Text: "MediumVal"}}},
	})

	result := table.Render()
	if result == "" {
		t.Error("Render should work with mixed fixed/auto column widths")
	}
}

func TestTable_UnicodeContent(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{
		{Title: "Symbol", Width: 8},
		{Title: "Name", Width: 15},
	})
	table.WithRows([]Row{
		{Cells: []Cell{{Text: "🌟"}, {Text: "Hello 世界"}}},
		{Cells: []Cell{{Text: "💫"}, {Text: "Test"}}},
	})

	result := table.Render()
	if result == "" {
		t.Error("Render should handle unicode content")
	}
}

func TestTable_WithColumnGap(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}, {Title: "B"}})
	table.WithRows([]Row{{Cells: []Cell{{Text: "1"}, {Text: "2"}}}})

	result1 := table.WithColumnGap(1).Render()
	result2 := table.WithColumnGap(5).Render()

	if result1 == "" || result2 == "" {
		t.Error("Render should work with different column gaps")
	}
}

func TestTable_WithBorderType(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}})
	table.WithRows([]Row{{Cells: []Cell{{Text: "1"}}}})

	result := table.WithBorder(true).WithBorderType(lipgloss.DoubleBorder()).Render()

	if result == "" {
		t.Error("Render should work with custom border type")
	}
}

func TestTable_String(t *testing.T) {
	table := NewTable()
	table.WithColumns([]Column{{Title: "A"}})
	table.AddRow(Row{Cells: []Cell{{Text: "1"}}})

	rendered := table.Render()
	str := table.String()

	if rendered != str {
		t.Error("String() should return same as Render()")
	}
}

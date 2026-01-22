package ui

import (
	"fmt"
	"strings"
)

// Table represents a simple table for displaying status
type Table struct {
	Headers []string
	Rows    [][]string
}

// NewTable creates a new table
func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

// Render renders the table
func (t *Table) Render() {
	// Calculate column widths
	widths := make([]int, len(t.Headers))
	for i, header := range t.Headers {
		widths[i] = len(header)
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	fmt.Println(strings.Repeat("─", sum(widths)+len(widths)*3+1))
	fmt.Print("│")
	for i, header := range t.Headers {
		fmt.Printf(" %-*s │", widths[i], header)
	}
	fmt.Println()

	// Print separator
	fmt.Print("├")
	for i, width := range widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")

	// Print rows
	for _, row := range t.Rows {
		fmt.Print("│")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf(" %-*s │", widths[i], cell)
			}
		}
		fmt.Println()
	}

	// Print footer
	fmt.Println(strings.Repeat("─", sum(widths)+len(widths)*3+1))
}

func sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

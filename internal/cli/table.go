package cli

import (
	"fmt"
	"io"
	"strings"
)

// printTable writes a header, a separator sized to the table, and the rows,
// padding every column to a common width.
//
// text/tabwriter is deliberately avoided here: a separator line contains no tab,
// which ends the column block and leaves the header aligned independently of the
// rows.
func printTable(out io.Writer, headers []string, rows [][]string) {
	printTableWithRowColors(out, headers, rows, nil)
}

func printTableWithRowColors(out io.Writer, headers []string, rows [][]string, coloredRows []bool) {
	const gutter = 2
	const yellow = "\x1b[33m"
	const reset = "\x1b[0m"

	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	total := 0
	for i, width := range widths {
		total += width
		if i < len(widths)-1 {
			total += gutter
		}
	}

	_, _ = fmt.Fprintln(out, padRow(headers, widths, gutter))
	_, _ = fmt.Fprintln(out, strings.Repeat("-", total))
	for index, row := range rows {
		line := padRow(row, widths, gutter)
		if index < len(coloredRows) && coloredRows[index] {
			line = yellow + line + reset
		}
		_, _ = fmt.Fprintln(out, line)
	}
}

func padRow(cells []string, widths []int, gutter int) string {
	var b strings.Builder
	for i, cell := range cells {
		if i == len(cells)-1 {
			b.WriteString(cell)
			break
		}
		b.WriteString(cell)
		b.WriteString(strings.Repeat(" ", widths[i]-len(cell)+gutter))
	}
	return b.String()
}

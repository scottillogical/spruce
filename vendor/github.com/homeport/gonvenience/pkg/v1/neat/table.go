// Copyright © 2018 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package neat

import (
	"bytes"
	"strings"

	"github.com/homeport/gonvenience/pkg/v1/bunt"
)

// TableOption defines options/settings for tables.
type TableOption func(*options)

// Alignment defines the text alignment option for a table cell.
type Alignment int

// Table cells support three types of alignment: left, right, center.
const (
	Left Alignment = iota
	Right
	Center
)

type options struct {
	filler          string
	separator       string
	desiredRowWidth int
	columnAlignment []Alignment
	errors          []error
}

func defaultOptions(cols int) options {
	alignments := make([]Alignment, cols)
	for i := 0; i < cols; i++ {
		alignments[i] = Left
	}

	return options{
		filler:          " ",
		separator:       " ",
		desiredRowWidth: -1,
		columnAlignment: alignments,
		errors:          []error{},
	}
}

// VertialBarSeparator sets a solid veritcal bar as the column separator.
func VertialBarSeparator() TableOption {
	return func(opts *options) {
		opts.separator = " │ "
	}
}

// CustomSeparator set a custom separator string (other than the default single space)
func CustomSeparator(separator string) TableOption {
	return func(opts *options) {
		opts.separator = separator
	}
}

// DesiredWidth sets the desired width of the table
func DesiredWidth(width int) TableOption {
	return func(opts *options) {
		opts.desiredRowWidth = width
	}
}

// AlignRight sets alignment to right for the given columns (referenced by index)
func AlignRight(cols ...int) TableOption {
	return func(opts *options) {
		for _, col := range cols {
			if col < 0 || col >= len(opts.columnAlignment) {
				opts.errors = append(opts.errors, &ColumnIndexIsOutOfBoundsError{col})
			} else {
				opts.columnAlignment[col] = Right
			}
		}
	}
}

// AlignCenter sets alignment to center for the given columns (referenced by index)
func AlignCenter(cols ...int) TableOption {
	return func(opts *options) {
		for _, col := range cols {
			if col < 0 || col >= len(opts.columnAlignment) {
				opts.errors = append(opts.errors, &ColumnIndexIsOutOfBoundsError{col})
			} else {
				opts.columnAlignment[col] = Center
			}
		}
	}
}

// Table renders a string with a well spaced and aligned table output
func Table(table [][]string, tableOptions ...TableOption) (string, error) {
	maxs, err := lookupMaxLengthPerColumn(table)
	if err != nil {
		return "", err
	}

	cols := len(maxs)
	options := defaultOptions(cols)

	for _, userOption := range tableOptions {
		userOption(&options)
	}

	if len(options.errors) > 0 {
		return "", options.errors[0]
	}

	var buf bytes.Buffer
	for _, row := range table {
		if options.desiredRowWidth > 0 {
			rawRowWidth := lookupPlainRowLength(row, maxs, options.separator)

			if rawRowWidth > options.desiredRowWidth {
				return "", &RowLengthExceedsDesiredWidthError{}
			}

			for y := range row {
				maxs[y] += (options.desiredRowWidth - rawRowWidth) / cols
			}
		}

		for y, cell := range row {
			notLastRow := y < len(row)-1
			fillment := strings.Repeat(
				options.filler,
				maxs[y]-bunt.PlainTextLength(cell),
			)

			switch options.columnAlignment[y] {
			case Left:
				buf.WriteString(cell)
				if notLastRow {
					buf.WriteString(fillment)
				}

			case Right:
				buf.WriteString(fillment)
				buf.WriteString(cell)

			case Center:
				x := bunt.PlainTextLength(fillment) / 2
				buf.WriteString(fillment[:x])
				buf.WriteString(cell)
				if notLastRow {
					buf.WriteString(fillment[x:])
				}
			}

			if notLastRow {
				buf.WriteString(options.separator)
			}
		}

		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func lookupMaxLengthPerColumn(table [][]string) ([]int, error) {
	if len(table) == 0 {
		return nil, &EmptyTableError{}
	}

	cols := len(table[0])
	for _, row := range table {
		if len(row) != cols {
			return nil, &ImbalancedTableError{}
		}
	}

	maxs := make([]int, cols)
	for _, row := range table {
		for y, cell := range row {
			if max := bunt.PlainTextLength(cell); max > maxs[y] {
				maxs[y] = max
			}
		}
	}

	return maxs, nil
}

func lookupPlainRowLength(row []string, maxs []int, separator string) int {
	var length int

	for i := range row {
		length += maxs[i] + bunt.PlainTextLength(separator)
	}

	return length
}

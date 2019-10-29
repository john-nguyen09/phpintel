package util

import "github.com/sourcegraph/go-lsp"

// IsInRange compare whether a position is within a range
func IsInRange(position lsp.Position, theRange lsp.Range) int {
	start := theRange.Start
	end := theRange.End

	if position.Line < start.Line ||
		(position.Line == start.Line && position.Character < start.Character) {
		return -1
	}
	if position.Line > end.Line ||
		(position.Line == end.Line && position.Character > end.Character) {
		return 1
	}
	return 0
}

// func PositionSearch(position lsp.Position, length int, getRange func(i int) lsp.Range) {
// 	i, j := 0, length
// 	for i < j {

// 	}
// }

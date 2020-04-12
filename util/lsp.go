package util

import (
	"strconv"
	"strings"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// IsInRange compare whether a position is within a range
func IsInRange(position protocol.Position, theRange protocol.Range) int {
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

// ComparePos returns -1, 0 or 1 indicating whether a is before, equal to or after b
func ComparePos(a protocol.Position, b protocol.Position) int {
	if a.Line < b.Line {
		return -1
	}
	if a.Line == b.Line {
		if a.Character < b.Character {
			return -1
		}
		if a.Character == b.Character {
			return 0
		}
		if a.Character > b.Character {
			return 1
		}
	}
	return 1
}

func CompareRange(a protocol.Range, b protocol.Range) int {
	if IsInRange(a.End, b) < 0 {
		return -1
	}
	if IsInRange(a.Start, b) > 0 {
		return 1
	}
	return 0
}

// PosFromString creates protocol.Position from string
func PosFromString(str string) protocol.Position {
	values := strings.Split(str, ":")
	line, err := strconv.Atoi(values[0])
	if err != nil {
		panic(err)
	}
	character, err := strconv.Atoi(values[1])
	if err != nil {
		panic(err)
	}
	return protocol.Position{
		Line:      line,
		Character: character,
	}
}

// RangeFromString creates protocol.Range from string
func RangeFromString(str string) protocol.Range {
	values := strings.Split(str, "-")
	start := PosFromString(values[0])
	end := PosFromString(values[1])
	return protocol.Range{
		Start: start,
		End:   end,
	}
}

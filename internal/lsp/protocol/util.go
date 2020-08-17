package protocol

import (
	"strconv"
	"strings"
)

// IsInRange compare whether a position is within a range
func IsInRange(position Position, theRange Range) int {
	start := theRange.Start
	end := theRange.End

	if position.Line < start.Line || (position.Line == start.Line && position.Character < start.Character) {
		return -1
	}
	if position.Line > end.Line || (position.Line == end.Line && position.Character >= end.Character) {
		if ComparePos(start, end) == 0 {
			return -1
		}
		return 1
	}
	return 0
}

// ComparePos returns -1, 0 or 1 indicating whether a is before, equal to or after b
func ComparePos(a Position, b Position) int {
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

// CompareRange compares 2 ranges
func CompareRange(a Range, b Range) int {
	if IsInRange(a.End, b) < 0 {
		return -1
	}
	if IsInRange(a.Start, b) > 0 {
		return 1
	}
	return 0
}

// PosFromString creates Position from string
func PosFromString(str string) Position {
	values := strings.Split(str, ":")
	line, err := strconv.Atoi(values[0])
	if err != nil {
		panic(err)
	}
	character, err := strconv.Atoi(values[1])
	if err != nil {
		panic(err)
	}
	return Position{
		Line:      line,
		Character: character,
	}
}

// RangeFromString creates Range from string
func RangeFromString(str string) Range {
	values := strings.Split(str, "-")
	start := PosFromString(values[0])
	end := PosFromString(values[1])
	return Range{
		Start: start,
		End:   end,
	}
}

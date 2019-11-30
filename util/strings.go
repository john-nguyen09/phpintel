package util

import "unicode/utf8"

// RunesToUTF8 converts []rune to []byte without converting to string first
func RunesToUTF8(rs []rune) []byte {
	bs := make([]byte, len(rs)*utf8.UTFMax)
	count := 0
	for _, r := range rs {
		count += utf8.EncodeRune(bs[count:], r)
	}
	return bs[:count]
}

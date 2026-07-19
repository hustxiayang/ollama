package mlxrunner

import "strings"

// findStop returns the byte index of the earliest occurrence of any stop string
// in text, and whether one was found. Empty stop strings are ignored.
func findStop(text string, stops []string) (int, bool) {
	idx := -1
	for _, s := range stops {
		if s == "" {
			continue
		}
		if i := strings.Index(text, s); i >= 0 && (idx < 0 || i < idx) {
			idx = i
		}
	}
	return idx, idx >= 0
}

// partialStopSuffix returns the length of the longest suffix of text that is a
// strict prefix of a stop string and therefore must be held for the next chunk.
func partialStopSuffix(text string, stops []string) int {
	keep := 0
	for _, s := range stops {
		if s == "" {
			continue
		}
		n := min(len(s)-1, len(text))
		for ; n > keep; n-- {
			if strings.HasSuffix(text, s[:n]) {
				keep = n
				break
			}
		}
	}
	return keep
}

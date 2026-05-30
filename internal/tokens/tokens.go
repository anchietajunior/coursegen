// Package tokens provides a rough, dependency-free token estimate.
//
// We deliberately avoid a real tokenizer: a ~4-chars-per-token heuristic is
// enough to make token economy *visible* (per-lesson and cumulative), which is
// the design goal — not billing-grade accuracy.
package tokens

import (
	"fmt"
	"math"
)

const charsPerToken = 4.0

// Estimate returns an approximate token count for the concatenation of texts.
func Estimate(texts ...string) int {
	chars := 0
	for _, t := range texts {
		chars += len(t)
	}
	return int(math.Ceil(float64(chars) / charsPerToken))
}

// Human formats a token count compactly: "920", "10.4k", "180.0k".
func Human(t int) string {
	if t < 1000 {
		return fmt.Sprintf("%d", t)
	}
	return fmt.Sprintf("%.1fk", float64(t)/1000.0)
}

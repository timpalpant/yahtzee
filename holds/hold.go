package holds

import (
	"github.com/timpalpant/yahtzee/dice"
)

var holdsCache = make([][][]int, dice.MaxHash)

func AllDistinctHolds(roll []int) [][]int {
	h := dice.Hash(roll)
	if result := holdsCache[h]; len(result) > 0 {
		return result
	}

	counts := make([]int, dice.NSides)
	for _, die := range roll {
		counts[die-1]++
	}

	result := enumerateDistinctHolds(counts, 0)
	holdsCache[h] = result
	return result
}

func enumerateDistinctHolds(counts []int, pos int) [][]int {
	if pos >= len(counts) {
		return [][]int{nil}
	}

	result := make([][]int, 0)
	for i := 0; i <= counts[pos]; i++ {
		toKeep := make([]int, i)
		for j := 0; j < i; j++ {
			toKeep[j] = pos + 1
		}

		for _, remaining := range enumerateDistinctHolds(counts, pos+1) {
			final := append(toKeep, remaining...)
			result = append(result, final)
		}
	}

	return result
}

package holds

import (
	"fmt"
)

// Enumerate all possible n-vectors of bools, representing
// all possible combinations of which dice to hold.
func AllPossibleHolds(n int) [][]bool {
	if n == 0 {
		return [][]bool{nil}
	}

	result := make([][]bool, 0)
	for _, option := range []bool{true, false} {
		for _, subHold := range AllPossibleHolds(n - 1) {
			hold := append(subHold, option)
			result = append(result, hold)
		}
	}

	return result
}

// Given the hold mask and dice roll, return a new vector with only
// the dice that were held.
func Keep(roll []int, hold []bool) []int {
	if len(roll) != len(hold) {
		panic(fmt.Errorf("cannot apply hold %v to roll %v", hold, roll))
	}

	result := make([]int, 0, len(roll))
	for i, held := range hold {
		if held {
			die := roll[i]
			result = append(result, die)
		}
	}

	return result
}

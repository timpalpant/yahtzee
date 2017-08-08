package tricks

import (
	"sort"
)

type Position int

const (
	Ones Position = 0
	Twos
	Threes
	Fours
	Fives
	Sixes
	ThreeOfAKind
	FourOfAKind
	FullHouse
	SmallStraight
	LargeStraight
	Chance
	Yahtzee
)

func (p Position) IsUpperHalf() bool {
	return p < 6
}

func (p Position) IsLowerHalf() bool {
	return p >= 6
}

func histogram(values []int) map[int]int {
	result := make(map[int]int, 6)
	for _, value := range values {
		result[value]++
	}

	return result
}

func isNOfAKind(dice []int, n int) bool {
	for _, count := range histogram(dice) {
		if count >= n {
			return true
		}
	}

	return false
}

func IsThreeOfAKind(dice []int) bool {
	return isNOfAKind(dice, 3)
}

func IsFourOfAKind(dice []int) bool {
	return isNOfAKind(dice, 4)
}

func IsFullHouse(dice []int) bool {
	h := histogram(dice)
	if len(h) != 2 {
		return false
	}

	for _, count := range h {
		if count != 2 && count != 3 {
			return false
		}
	}

	return true
}

// Either 1,2,3,4; 2,3,4,5; or 3,4,5,6.
func IsSmallStraight(dice []int) bool {
	case1 := map[int]struct{}{
		1: struct{}{},
		2: struct{}{},
		3: struct{}{},
		4: struct{}{},
	}

	case2 := map[int]struct{}{
		2: struct{}{},
		3: struct{}{},
		4: struct{}{},
		5: struct{}{},
	}

	case3 := map[int]struct{}{
		3: struct{}{},
		4: struct{}{},
		5: struct{}{},
		6: struct{}{},
	}

	for _, die := range dice {
		delete(case1, die)
		delete(case2, die)
		delete(case3, die)
	}

	return len(case1) == 0 || len(case2) == 0 || len(case3) == 0
}

func IsLargeStraight(dice []int) bool {
	sort.Ints(dice)
	first := dice[0]
	for i, value := range dice {
		if value != first+i {
			return false
		}
	}

	return true
}

func IsYahtzee(dice []int) bool {
	return isNOfAKind(dice, 5)
}

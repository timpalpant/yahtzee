package scoring

import (
	"github.com/timpalpant/yahtzee/tricks"
)

func PositionScore(dice []int, position tricks.Position) int {
	if position.IsUpperHalf() {
		dieFace := position + 1
		return dieFace * Count(dice, dieFace)
	}

	// Lower half.
	switch position {
	case tricks.ThreeOfAKind:
		if tricks.IsThreeOfAKind(dice) {
			return scoring.Sum(dice)
		}
	case tricks.FourOfAKind:
		if tricks.IsFourOfAKind(dice) {
			return scoring.Sum(dice)
		}
	case tricks.FullHouse:
		if tricks.IsFullHouse(dice) {
			return 25
		}
	case tricks.SmallStraight:
		if tricks.IsSmallStraight(dice) {
			return 30
		}
	case tricks.LargeStraight:
		if tricks.IsLargeStraight(dice) {
			return 40
		}
	case tricks.Chance:
		return scoring.Sum(dice)
	case tricks.Yahtzee:
		if tricks.IsYahtzee(dice) {
			return 50
		}
	}

	return 0
}

// Return the number of dice that equal the given value.
func Count(dice []int, value int) int {
	result := 0
	for _, die := range dice {
		if die == value {
			result++
		}
	}

	return result
}

// Return the sum of the values of the dice.
func Sum(dice []int) int {
	total := 0
	for _, die := range dice {
		total += die
	}

	return total
}

package yahtzee

import (
	"time"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee/dice"
	"github.com/timpalpant/yahtzee/holds"
)

const nDice = 5

var ExpectedScoreCache = make([]float64, 6400000)

func ExpectedScore(gs GameState) float64 {
	if gs.GameOver() {
		return 0.0 // Game over.
	}

	h := gs.Hash()
	if score := ExpectedScoreCache[h]; score != 0 {
		return score
	}

	remainingPositions := gs.AvailablePositions()

	glog.Infof("Computing expected score for: %v", gs.String())
	start := time.Now()
	countIter := 0
	expectedScore := expectedValue(nil, func(roll1 []int) float64 {
		glog.Infof("Roll 1: %v", roll1)
		return maxValue(roll1, func(hold1 []int) float64 {
			return expectedValue(hold1, func(roll2 []int) float64 {
				return maxValue(roll2, func(hold2 []int) float64 {
					return expectedValue(hold2, func(finalRoll []int) float64 {
						bestPlacement := 0.0
						for _, position := range remainingPositions {
							played, addedValue := gs.PlayPosition(finalRoll, position)
							expectedRemainingScore := ExpectedScore(played)
							expectedPositionValue := float64(addedValue) + expectedRemainingScore

							if expectedPositionValue > bestPlacement {
								bestPlacement = expectedPositionValue
							}

							countIter++
						}

						return bestPlacement
					})
				})
			})
		})
	})

	elapsed := time.Since(start)
	iterPerSec := float64(countIter) / elapsed.Seconds()
	glog.Infof("Expected score = %.2f for %v (%d iterations, %v, %.1f iter/s)",
		expectedScore, gs.String(), countIter, elapsed, iterPerSec)
	ExpectedScoreCache[h] = expectedScore
	return expectedScore
}

// Return the expected value of f over all rolls of 5 dice, constructed starting
// with the given initial dice (which may be nil).
func expectedValue(initialDice []int, f func(roll []int) float64) float64 {
	result := 0.0

	for _, roll := range dice.AllPossibleRolls(initialDice) {
		result += roll.Probability * f(roll.Dice)
	}

	return result
}

// Return the max value of f over all distinct holds of the 5 dice.
func maxValue(roll []int, f func(hold []int) float64) float64 {
	result := 0.0

	for _, kept := range holds.AllDistinctHolds(roll) {
		x := f(kept)
		if x > result {
			result = x
		}
	}

	return result
}

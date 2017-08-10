package yahtzee

import (
	"time"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee/dice"
	"github.com/timpalpant/yahtzee/holds"
)

const nDice = 5

var expectedScoreCache = map[GameState]float64{}

func ExpectedScore(gs GameState) float64 {
	remainingPositions := gs.AvailablePositions()
	if len(remainingPositions) == 0 {
		return 0.0 // Game over.
	}

	if score, ok := expectedScoreCache[gs]; ok {
		return score
	}

	glog.Infof("Computing expected score for: %v", gs)
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
							positionValue := gs.ValueAt(finalRoll, position)
							played := gs.PlayPosition(finalRoll, position)
							expectedRemainingScore := ExpectedScore(played)
							expectedPositionValue := float64(positionValue) + expectedRemainingScore

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
		expectedScore, gs, countIter, iterPerSec)
	expectedScoreCache[gs] = expectedScore
	return expectedScore
}

// Return the expected value of f over all rolls of 5 dice, constructed starting
// with the given initial dice (which may be nil).
func expectedValue(initialDice []int, f func(roll []int) float64) float64 {
	result := 0.0

	for _, remainingDice := range dice.AllPossibleRolls(nDice - len(initialDice)) {
		roll := append(initialDice, remainingDice...)
		p := dice.Probability(roll)

		result += p * f(roll)
	}

	return result
}

var allPossibleHolds = holds.AllPossibleHolds(nDice)

// Return the max value of f over all possible holds of the 5 dice.
func maxValue(roll []int, f func(hold []int) float64) float64 {
	result := 0.0

	for _, hold := range allPossibleHolds {
		kept := holds.Keep(roll, hold)
		x := f(kept)

		if x > result {
			result = x
		}
	}

	return result
}

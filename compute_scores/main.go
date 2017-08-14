package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

var (
	nGamesComputed = 0
	held1Caches    = make2DCache(yahtzee.NumTurns, yahtzee.MaxRoll)
	held2Caches    = make2DCache(yahtzee.NumTurns, yahtzee.MaxRoll)
)

// ScoreCache memoizes computed values. It is designed to be efficiently
// reusable by resetting the isSet array (which uses an efficient memset).
// Unset values are not defined.
type ScoreCache struct {
	values []float64
	isSet  []bool
}

func NewScoreCache(size int) *ScoreCache {
	return &ScoreCache{
		values: make([]float64, size),
		isSet:  make([]bool, size),
	}
}

func (sc *ScoreCache) Reset() {
	for i := range sc.isSet {
		sc.isSet[i] = false
	}
}

func (sc *ScoreCache) Set(key uint, value float64) {
	sc.values[key] = value
	sc.isSet[key] = true
}

func make2DCache(size1, size2 int) []*ScoreCache {
	result := make([]*ScoreCache, size1)
	for i := range result {
		result[i] = NewScoreCache(size2)
	}
	return result
}

func bestScoreForRoll(scores *ScoreCache, game yahtzee.GameState, roll yahtzee.Roll) float64 {
	best := 0.0
	for _, box := range game.AvailableBoxes() {
		newGame, addedValue := game.FillBox(box, roll)
		expectedRemainingScore := computeExpectedScores(scores, newGame)
		expectedPositionValue := float64(addedValue) + expectedRemainingScore

		if expectedPositionValue > best {
			best = expectedPositionValue
		}
	}

	return best
}

func expectedScoreForHold(heldCache *ScoreCache, held yahtzee.Roll, heldValue func(held yahtzee.Roll) float64) float64 {
	if heldCache.isSet[held] {
		return heldCache.values[held]
	}

	eValue := 0.0
	if held.NumDice() == yahtzee.NDice {
		eValue = heldValue(held)
	} else {
		for side := 1; side <= yahtzee.NSides; side++ {
			eValue += expectedScoreForHold(heldCache, held.Add(side), heldValue) / yahtzee.NSides
		}
	}

	heldCache.Set(uint(held), eValue)
	return eValue
}

func computeExpectedScores(scores *ScoreCache, game yahtzee.GameState) float64 {
	if game.GameOver() {
		return 0.0
	}

	if scores.isSet[game] {
		return scores.values[game]
	}

	expectedScore := 0.0
	remainingBoxes := game.AvailableBoxes()
	depth := yahtzee.NumTurns - len(remainingBoxes)
	held1Cache := held1Caches[depth]
	held1Cache.Reset()
	held2Cache := held2Caches[depth]
	held2Cache.Reset()

	for _, roll1 := range yahtzee.NewRoll().SubsequentRolls() {
		maxValue1 := 0.0
		for _, held1 := range roll1.PossibleHolds() {
			eValue2 := expectedScoreForHold(held1Cache, held1, func(roll2 yahtzee.Roll) float64 {
				maxValue2 := 0.0
				for _, held2 := range roll2.PossibleHolds() {
					eValue3 := expectedScoreForHold(held2Cache, held2, func(roll3 yahtzee.Roll) float64 {
						return bestScoreForRoll(scores, game, roll3)
					})

					if eValue3 > maxValue2 {
						maxValue2 = eValue3
					}
				}

				return maxValue2
			})

			if eValue2 > maxValue1 {
				maxValue1 = eValue2
			}
		}

		expectedScore += roll1.Probability() * maxValue1
	}

	nGamesComputed++
	if nGamesComputed%10000 == 0 {
		glog.Infof("%d games computed, current: %v = %g",
			nGamesComputed, game, expectedScore)
	}

	scores.Set(uint(game), expectedScore)
	return expectedScore
}

func main() {
	outputFilename := flag.String("output", "scores.txt", "Output filename")
	flag.Parse()

	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Info("Computing expected score table")
	game := yahtzee.NewGame()
	scores := NewScoreCache(yahtzee.MaxGame)
	computeExpectedScores(scores, game)

	glog.Infof("Expected score: %.2f", scores.values[game])
	glog.Infof("Writing score table to: %v", *outputFilename)
	f, err := os.Create(*outputFilename)
	if err != nil {
		glog.Fatal(err)
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()
	for game, isSet := range scores.isSet {
		if isSet {
			fmt.Fprintf(buf, "%v\t%v\n", game, scores.values[game])
		}
	}
}

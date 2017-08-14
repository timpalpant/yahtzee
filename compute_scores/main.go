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

const unset = -1

var (
	nGamesComputed = 0
	held1Caches    = make2DCache(yahtzee.NumTurns, yahtzee.MaxRoll)
	held2Caches    = make2DCache(yahtzee.NumTurns, yahtzee.MaxRoll)
)

func makeCache(size int) []float64 {
	result := make([]float64, size)
	resetCache(result)
	return result
}

func resetCache(c []float64) []float64 {
	for i := range c {
		c[i] = unset
	}
	return c
}

func make2DCache(size1, size2 int) [][]float64 {
	result := make([][]float64, size1)
	for i := range result {
		result[i] = makeCache(size2)
	}
	return result
}

func bestScoreForRoll(scores []float64, game yahtzee.GameState, roll yahtzee.Roll) float64 {
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

func expectedScoreForHold(heldCache []float64, held yahtzee.Roll, rollValue func(roll yahtzee.Roll) float64) float64 {
	eValue := heldCache[held]
	if eValue != unset {
		return eValue
	}

	eValue = 0.0
	if held.NumDice() == yahtzee.NDice {
		eValue = rollValue(held)
	} else {
		for side := 1; side <= yahtzee.NSides; side++ {
			eValue += expectedScoreForHold(heldCache, held.Add(side), rollValue) / yahtzee.NSides
		}
	}

	heldCache[held] = eValue
	return eValue
}

func computeExpectedScores(scores []float64, game yahtzee.GameState) float64 {
	if game.GameOver() {
		return 0.0
	}

	expectedScore := scores[game]
	if expectedScore != unset {
		return expectedScore
	}

	expectedScore = 0
	remainingBoxes := game.AvailableBoxes()
	depth := yahtzee.NumTurns - len(remainingBoxes)
	held1Cache := resetCache(held1Caches[depth])
	held2Cache := resetCache(held2Caches[depth])

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

	scores[game] = expectedScore
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
	scores := makeCache(yahtzee.MaxGame)
	computeExpectedScores(scores, game)

	glog.Infof("Expected score: %.2f", scores[game])
	glog.Infof("Writing score table to: %v", *outputFilename)
	f, err := os.Create(*outputFilename)
	if err != nil {
		glog.Fatal(err)
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()
	for game, score := range scores {
		if score != -1 {
			fmt.Fprintf(buf, "%v\t%v\n", game, score)
		}
	}
}

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
	roll2Caches    = make2DCache(yahtzee.NumTurns, yahtzee.MaxRoll)
	roll3Caches    = make2DCache(yahtzee.NumTurns, yahtzee.MaxRoll)
)

type cacheEntry struct {
	isSet bool
	value float64
}

func makeCache(size int) []cacheEntry {
	return make([]cacheEntry, size)
}

func resetCache(c []cacheEntry) []cacheEntry {
	for i := range c {
		c[i] = cacheEntry{}
	}
	return c
}

func make2DCache(size1, size2 int) [][]cacheEntry {
	result := make([][]cacheEntry, size1)
	for i := range result {
		result[i] = makeCache(size2)
	}
	return result
}

func computeExpectedScores(scores []cacheEntry, game yahtzee.GameState) float64 {
	if game.GameOver() {
		return 0.0
	}

	expectedScore := scores[game]
	if expectedScore.isSet {
		return expectedScore.value
	}

	remainingBoxes := game.AvailableBoxes()
	depth := yahtzee.NumTurns - len(remainingBoxes)
	roll2Cache := resetCache(roll2Caches[depth])
	roll3Cache := resetCache(roll3Caches[depth])

	for _, roll1 := range yahtzee.NewRoll().SubsequentRolls() {
		maxValue1 := 0.0
		for _, held1 := range roll1.PossibleHolds() {
			eValue2 := 0.0
			for _, roll2 := range held1.SubsequentRolls() {
				maxValue2 := roll2Cache[roll2]
				if !maxValue2.isSet {
					for _, held2 := range roll2.PossibleHolds() {
						eValue3 := 0.0
						for _, roll3 := range held2.SubsequentRolls() {
							maxValue3 := roll3Cache[roll3]
							if !maxValue3.isSet {
								for _, box := range remainingBoxes {
									newGame, addedValue := game.FillBox(box, roll3)
									expectedRemainingScore := computeExpectedScores(scores, newGame)
									expectedPositionValue := float64(addedValue) + expectedRemainingScore

									if expectedPositionValue > maxValue3.value {
										maxValue3.value = expectedPositionValue
									}
								}

								maxValue3.isSet = true
								roll3Cache[roll3] = maxValue3
							}

							// Conditional probability of this roll starting from the held one.
							p := (roll3 - held2).Probability()
							eValue3 += p * maxValue3.value
						}

						if eValue3 > maxValue2.value {
							maxValue2.value = eValue3
						}
					}

					maxValue2.isSet = true
					roll2Cache[roll2] = maxValue2
				}

				p := (roll2 - held1).Probability()
				eValue2 += p * maxValue2.value
			}

			if eValue2 > maxValue1 {
				maxValue1 = eValue2
			}
		}

		expectedScore.value += roll1.Probability() * maxValue1
	}

	nGamesComputed++
	if nGamesComputed%10000 == 0 {
		glog.Infof("%d games computed, current: %v = %g",
			nGamesComputed, game, expectedScore.value)
	}

	expectedScore.isSet = true
	scores[game] = expectedScore
	return expectedScore.value
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
		if score.isSet {
			fmt.Fprintf(buf, "%v\t%v\n", game, score)
		}
	}
}

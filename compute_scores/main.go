package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

const unset = -1

func makeCache(size int) []float64 {
	result := make([]float64, size)
	for i := range result {
		result[i] = unset
	}
	return result
}

func computeExpectedScores(scores []float64, game yahtzee.GameState) float64 {
	if game.GameOver() {
		return 0.0
	}

	expectedScore := scores[game]
	if expectedScore != unset {
		return expectedScore
	}

	glog.V(1).Infof("Computing expected score for game %s", game)
	expectedScore = 0
	remainingBoxes := game.AvailableBoxes()
	roll2Cache := makeCache(yahtzee.MaxRoll)
	roll3Cache := makeCache(yahtzee.MaxRoll)

	for _, roll1 := range yahtzee.NewRoll().SubsequentRolls() {
		maxValue1 := 0.0
		for _, held1 := range roll1.PossibleHolds() {
			eValue2 := 0.0
			for _, roll2 := range held1.SubsequentRolls() {
				maxValue2 := roll2Cache[roll2]
				if maxValue2 != unset {
					return maxValue2
				}

				maxValue2 = 0.0
				for _, held2 := range roll2.PossibleHolds() {
					eValue3 := 0.0
					for _, roll3 := range held2.SubsequentRolls() {
						maxValue3 := roll3Cache[roll3]
						if maxValue3 != unset {
							return maxValue3
						}

						maxValue3 = 0.0
						for _, box := range remainingBoxes {
							newGame, addedValue := game.FillBox(box, roll3)
							expectedRemainingScore := computeExpectedScores(scores, newGame)
							expectedPositionValue := float64(addedValue) + expectedRemainingScore

							if expectedPositionValue > maxValue3 {
								maxValue3 = expectedPositionValue
							}
						}

						roll3Cache[roll3] = maxValue3
						// Conditional probability of this roll starting from the held one.
						p := (roll3 - held2).Probability()
						eValue3 += p * maxValue3
					}

					if eValue3 > maxValue2 {
						maxValue2 = eValue3
					}
				}

				roll2Cache[roll2] = maxValue2
				p := (roll2 - held1).Probability()
				eValue2 += p * maxValue2
			}

			if eValue2 > maxValue1 {
				maxValue1 = eValue2
			}
		}

		expectedScore += roll1.Probability() * maxValue1
	}

	glog.V(1).Infof("Expected score for game %s = %.2f", game, expectedScore)
	scores[game] = expectedScore
	return expectedScore
}

func main() {
	outputFilename := flag.String("output", "scores.txt", "Output filename")
	flag.Parse()

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
		fmt.Fprintf(buf, "%v\t%v\n", game, score)
	}
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

var stdin = bufio.NewReader(os.Stdin)

func prompt(msg string) string {
	fmt.Print(msg)
	result, err := stdin.ReadString('\n')
	if err != nil {
		panic(err)
	}

	return strings.TrimRight(result, "\n")
}

func parseRoll(s string) (yahtzee.Roll, error) {
	roll := yahtzee.NewRoll()

	i, err := strconv.Atoi(s)
	if err != nil {
		return roll, err
	}

	for ; i > 0; i /= 10 {
		die := i % 10
		roll = roll.Add(die)
	}

	if roll.NumDice() != yahtzee.NDice {
		return roll, fmt.Errorf("Invalid number of dice: %v != %v",
			roll.NumDice(), yahtzee.NDice)
	}

	return roll, nil
}

func promptRoll() yahtzee.Roll {
	for {
		rollStr := prompt("Enter roll: ")
		roll, err := parseRoll(rollStr)
		if err != nil {
			fmt.Printf("Invalid roll: %v\n", err)
			continue
		}

		return roll
	}
}

func grValue(currentScore int, gr yahtzee.GameResult, scoreToBeat int) float64 {
	switch gr := gr.(type) {
	case yahtzee.ExpectedValue:
		return float64(gr) + float64(currentScore)
	case yahtzee.ScoreDistribution:
		return gr.GetProbability(scoreToBeat)
	}

	panic("Unknown game result type")
}

func bestHold(currentScore int, outcomes map[yahtzee.Roll]yahtzee.GameResult, scoreToBeat int) (yahtzee.Roll, float64) {
	var best yahtzee.Roll
	var bestValue float64
	for hold, gr := range outcomes {
		value := grValue(currentScore, gr, scoreToBeat)
		if value >= bestValue {
			best = hold
			bestValue = value
		}
	}

	return best, bestValue
}

func bestBox(currentScore int, outcomes map[yahtzee.Box]yahtzee.GameResult, scoreToBeat int) (yahtzee.Box, float64) {
	var best yahtzee.Box
	var bestValue float64
	for box, gr := range outcomes {
		value := grValue(currentScore, gr, scoreToBeat)
		if value >= bestValue {
			best = box
			bestValue = value
		}
	}

	return best, bestValue
}

func playGame(strat *yahtzee.Strategy, scoreToBeat int) {
	fmt.Println("Welcome to YAHTZEE!")
	game := yahtzee.NewGame()
	var currentScore int

	for !game.GameOver() {
		opt := yahtzee.NewTurnOptimizer(strat, game)
		roll1 := promptRoll()
		hold1Outcomes := opt.GetHold1Outcomes(roll1)
		bestHold1, value := bestHold(currentScore, hold1Outcomes, scoreToBeat)
		fmt.Printf("Best option is to hold: %v, value: %g\n",
			bestHold1, value)

		roll2 := promptRoll()
		hold2Outcomes := opt.GetHold2Outcomes(roll2)
		bestHold2, value := bestHold(currentScore, hold2Outcomes, scoreToBeat)
		fmt.Printf("Best option is to hold: %v, value: %g\n",
			bestHold2, value)

		roll3 := promptRoll()
		fillOutcomes := opt.GetFillOutcomes(roll3)
		box, value := bestBox(currentScore, fillOutcomes, scoreToBeat)
		var addValue int
		game, addValue = game.FillBox(box, roll3)
		currentScore += addValue
		fmt.Printf("Best option is to play: %v for %v points, final value: %g\n",
			box, addValue, value)
	}

	fmt.Printf("Game over! Final score: %v\n", currentScore)
}

func main() {
	expectedScores := flag.String(
		"expected_scores", "expected-scores.jsonlines",
		"File with expected scores to load")
	scoreDistributions := flag.String(
		"score_distributions", "score-distributions.jsonlines",
		"File with score distributions to load")
	scoreToBeat := flag.Int(
		"score_to_beat", -1,
		"High score to beat")
	flag.Parse()

	var strat *yahtzee.Strategy
	var err error
	if *scoreToBeat > 0 {
		glog.Info("Loading score distributions table")
		strat, err = yahtzee.LoadScoreDistributionsTable(*scoreDistributions)
	} else {
		glog.Info("Loading expected scores table")
		strat, err = yahtzee.LoadExpectedScoresTable(*expectedScores)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	playGame(strat, *scoreToBeat)
}

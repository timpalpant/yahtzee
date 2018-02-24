package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/optimization"
)

func main() {
	observable := flag.String("observable", "expected_value",
		"Observable to compute (expected_value, score_distribution, expected_work)")
	outputFilename := flag.String("output", "scores.gob.gz", "Output filename")
	scoreToBeat := flag.Int("score_to_beat", 200, "Score to beat for expected work calculation")
	flag.Parse()

	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Infof("Max game is: %v", yahtzee.MaxGame)
	glog.Infof("Max roll is: %v", yahtzee.MaxRoll)
	glog.Infof("Max score is: %v", yahtzee.MaxScore)

	glog.Info("Computing expected score table")
	var game yahtzee.Game = yahtzee.NewGame()
	var obs optimization.GameResult
	switch *observable {
	case "expected_value":
		obs = optimization.NewExpectedValue()
	case "score_distribution":
		obs = optimization.NewScoreDistribution()
	case "expected_work":
		obs = optimization.NewExpectedWork(*scoreToBeat, 10000)
		game = yahtzee.NewScoredGameState()
	default:
		glog.Fatal("Unknown observable: %v, options: expected_value, score_distribution")
	}

	s := optimization.NewStrategy(obs)
	result := s.Compute(game)

	glog.Infof("Expected score: %.2f", result)
	glog.Infof("Writing score table to: %v", *outputFilename)
	err := s.SaveToFile(*outputFilename)
	if err != nil {
		glog.Fatal(err)
	}
}

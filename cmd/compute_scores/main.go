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
	iter := flag.Int("iter", 1, "Number of iterations to perform")
	resume := flag.String("resume", "", "Resume calculation from given output")
	flag.Parse()

	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Infof("Max game is: %v", yahtzee.MaxGame)
	glog.Infof("Max roll is: %v", yahtzee.MaxRoll)
	glog.Infof("Max score is: %v", yahtzee.MaxScore)

	var obs optimization.GameResult
	switch *observable {
	case "expected_value":
		obs = optimization.NewExpectedValue()
	case "score_distribution":
		obs = optimization.NewScoreDistribution()
	case "expected_work":
		obs = optimization.NewExpectedWork(10000)
	default:
		glog.Fatal("Unknown observable: %v, options: expected_value, score_distribution")
	}

	var s *optimization.Strategy
	if *resume != "" {
		glog.Infof("Resuming training, loading cache from %v", *resume)
		s = optimization.NewStrategy(obs)
		s.LoadCache(*resume)
		obs = s.Compute(yahtzee.NewGame())
	}

	glog.Info("Computing expected score table")
	for i := 0; i < *iter; i++ {
		s = optimization.NewStrategy(obs)
		obs = s.Populate()
		glog.Infof("Expected score after iteration %v: %.2f", i, obs)

		glog.Infof("Writing score table to: %v", *outputFilename)
		err := s.SaveToFile(*outputFilename)
		if err != nil {
			glog.Fatal(err)
		}
	}
}

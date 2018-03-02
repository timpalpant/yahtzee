package main

import (
	"encoding/gob"
	"flag"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/golang/glog"
	gzip "github.com/klauspost/pgzip"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/optimization"
)

func loadGames(filename string) ([]yahtzee.GameState, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gzf.Close()

	dec := gob.NewDecoder(gzf)
	var games []yahtzee.GameState
	if err := dec.Decode(&games); err != nil && err != io.EOF {
		return nil, err
	}

	// TODO: Remove once enumerated games have been regenerated.
	games = append(games, yahtzee.NewGame())

	return games, nil
}

func getObservable(observable string) optimization.GameResult {
	switch observable {
	case "expected_value":
		return optimization.NewExpectedValue()
	case "score_distribution":
		return optimization.NewScoreDistribution()
	case "expected_work":
		return optimization.NewExpectedWork(10000)
	default:
		glog.Fatal("Unknown observable: %v, options: expected_value, score_distribution")
	}

	return nil
}

func main() {
	gamesFilename := flag.String("games", "", "File with all enumerated games")
	observable := flag.String("observable", "expected_value",
		"Observable to compute (expected_value, score_distribution, expected_work)")
	outputFilename := flag.String("output", "scores.gob.gz", "Output filename")
	iter := flag.Int("iter", 1, "Number of iterations to perform")
	resume := flag.String("resume", "", "Resume calculation from given output")
	flag.Parse()

	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Infof("Max game is: %d", yahtzee.MaxGame)
	glog.Infof("Max scored game is: %d", yahtzee.MaxScoredGame)
	glog.Infof("Max roll is: %d", yahtzee.MaxRoll)
	glog.Infof("Max score is: %d", yahtzee.MaxScore)

	glog.Infof("Loading games")
	games, err := loadGames(*gamesFilename)
	if err != nil {
		glog.Fatal(err)
	} else {
		glog.Infof("Loaded %v game states", len(games))
	}

	obs := getObservable(*observable)

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
		s.Populate(games, *outputFilename)
		obs = s.Compute(yahtzee.NewGame())
		glog.Infof("E_0 after iteration %v: %.2f", i, obs)

		glog.Infof("Writing results to: %v", *outputFilename)
		err := s.SaveToFile(*outputFilename)
		if err != nil {
			glog.Fatal(err)
		}
	}
}

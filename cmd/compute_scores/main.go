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

func loadGames(filename string, scoreDependent bool) ([]yahtzee.GameState, error) {
	if filename == "" {
		glog.Warning("-games not provided, enumerating games")
		return yahtzee.AllGameStates(scoreDependent), nil
	}

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

func main() {
	gamesFilename := flag.String("games", "", "File with all enumerated games")
	observable := flag.String("observable", "expected_value",
		"Observable to compute (expected_value, score_distribution, expected_work)")
	outputFilename := flag.String("output", "scores.gob.gz", "Output filename")
	iter := flag.Int("iter", 1, "Number of iterations to perform")
	resume := flag.String("resume", "", "Resume calculation from given output")
	scoreToBeat := flag.Int("score_to_beat", 200, "Desired score to beat (for expected_work)")
	e0 := flag.Float64("e0", 10000, "Guess for expected initial work at start of game")
	flag.Parse()

	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Infof("Max game is: %d", yahtzee.MaxGame)
	glog.Infof("Max scored game is: %d", yahtzee.MaxScoredGame)
	glog.Infof("Max roll is: %d", yahtzee.MaxRoll)
	glog.Infof("Max score is: %d", yahtzee.MaxScore)
	glog.Infof("Number of distinct rolls: %d", len(yahtzee.AllDistinctRolls()))

	var obs optimization.GameResult
	switch *observable {
	case "expected_value":
		obs = optimization.NewExpectedValue()
	case "score_distribution":
		obs = optimization.NewScoreDistribution()
	case "expected_work":
		obs = optimization.NewExpectedWork(*scoreToBeat, float32(*e0))
	default:
		glog.Fatalf("Unknown observable: %v, options: expected_value, score_distribution, expected_work", *observable)
	}

	glog.Infof("Loading games")
	games, err := loadGames(*gamesFilename, obs.ScoreDependent())
	if err != nil {
		glog.Fatal(err)
	} else {
		glog.Infof("Loaded %v game states", len(games))
	}

	var s *optimization.Strategy
	if *resume != "" {
		glog.Infof("Resuming training, loading cache from %v", *resume)
		s, err = optimization.LoadFromFile(*resume)
		if err != nil {
			glog.Fatal(err)
		}
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

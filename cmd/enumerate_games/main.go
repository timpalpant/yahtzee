package main

import (
	"encoding/gob"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/golang/glog"
	gzip "github.com/klauspost/pgzip"

	"github.com/timpalpant/yahtzee"
)

func saveGames(allGames []yahtzee.GameState, outputFilename string) {
	f, err := os.Create(outputFilename)
	if err != nil {
		glog.Fatal(err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	enc := gob.NewEncoder(gzw)
	if err := enc.Encode(allGames); err != nil {
		glog.Fatal(err)
	}
}

func main() {
	withScore := flag.Bool("scored", false, "Enumerate distinct gaame states with scores")
	outputFilename := flag.String("output", "games.gob.gz", "Output filename")
	flag.Parse()

	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Infof("Max game is: %d", yahtzee.MaxGame)
	glog.Infof("Max scored game is: %d", yahtzee.MaxScoredGame)
	glog.Infof("Max roll is: %d", yahtzee.MaxRoll)
	glog.Infof("Max score is: %d", yahtzee.MaxScore)

	allGames := yahtzee.AllGameStates(*withScore)
	saveGames(allGames, *outputFilename)
}

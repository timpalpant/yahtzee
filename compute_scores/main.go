package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

func main() {
	outputFilename := flag.String("output", "scores.txt", "Output filename")
	flag.Parse()

	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	glog.Info("Computing expected score table")
	ev := yahtzee.NewScoreDistribution()
	s := yahtzee.NewStrategy(ev)
	result := s.Compute(yahtzee.NewGame())

	glog.Infof("Expected score: %.2f", result)
	glog.Infof("Writing score table to: %v", *outputFilename)
	err := s.SaveToFile(*outputFilename)
	if err != nil {
		glog.Fatal(err)
	}
}

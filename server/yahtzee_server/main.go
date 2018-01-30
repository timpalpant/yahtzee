package main

import (
	"flag"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/server"
)

func main() {
	expectedScores := flag.String(
		"expected_scores", "data/expected-scores.gob.gz",
		"File with expected scores to load")
	scoreDistributions := flag.String(
		"score_distributions", "data/score-distributions.gob.gz",
		"File with score distributions to load")
	flag.Parse()

	glog.Info("Loading expected scores table")
	expectedScoreStrat := yahtzee.NewStrategy(yahtzee.NewExpectedValue())
	err := expectedScoreStrat.LoadCache(*expectedScores)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Info("Loading score distributions table")
	highScoreStrat := yahtzee.NewStrategy(yahtzee.NewScoreDistribution())
	err = highScoreStrat.LoadCache(*scoreDistributions)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Info("Starting server")
	server := server.NewYahtzeeServer(highScoreStrat, expectedScoreStrat)
	http.Handle("/",
		gziphandler.GzipHandler(http.HandlerFunc(server.Index)))
	http.Handle("/rest/v1/outcome_distribution",
		gziphandler.GzipHandler(http.HandlerFunc(server.OutcomeDistribution)))
	http.Handle("/static/", gziphandler.GzipHandler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
	glog.Fatal(http.ListenAndServe(":8080", nil))
}

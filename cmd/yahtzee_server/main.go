package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/server"
)

func main() {
	expectedScores := flag.String(
		"expected_scores", "../../data/expected-scores.gob.gz",
		"File with expected scores to load")
	scoreDistributions := flag.String(
		"score_distributions", "../../data/score-distributions.gob.gz",
		"File with score distributions to load")
	port := flag.Int("port", 8080, "Port to bind to")
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
	http.Handle("/rest/v1/score",
		gziphandler.GzipHandler(http.HandlerFunc(server.GetScore)))
	http.Handle("/rest/v1/optimal_move",
		gziphandler.GzipHandler(http.HandlerFunc(server.OptimalMove)))
	http.Handle("/rest/v1/outcome_distribution",
		gziphandler.GzipHandler(http.HandlerFunc(server.OutcomeDistribution)))
	http.Handle("/static/", gziphandler.GzipHandler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

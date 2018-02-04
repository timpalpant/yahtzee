package main

import (
	"flag"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/rpi/controller"
	"github.com/timpalpant/yahtzee/rpi/detector"
	"github.com/timpalpant/yahtzee/rpi/player"
)

func main() {
	scoreDistributions := flag.String(
		"score_distributions", "../data/score-distributions.gob.gz",
		"File with score distributions to load")
	scoreToBeat := flag.Int("score_to_beat", 300, "Score to beat")
	flag.Parse()

	glog.Info("Loading score distributions table")
	highScoreStrat := yahtzee.NewStrategy(yahtzee.NewScoreDistribution())
	err := highScoreStrat.LoadCache(*scoreDistributions)
	if err != nil {
		glog.Fatal(err)
	}

	detector := detector.NewYahtzeeDetector()
	controller, err := controller.NewYahtzeeController(controller.DefaultWiringConfig)
	if err != nil {
		glog.Fatal(err)
	}
	defer controller.Close()

	player := player.NewYahtzeePlayer(detector, controller)
	player.Play(highScoreStrat, *scoreToBeat)
}

package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/stianeikeland/go-rpio"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/rpi/controller"
	"github.com/timpalpant/yahtzee/rpi/detector"
)

var wiringConfig = &controller.YahtzeeControllerConfig{

}

func main() {
	scoreDistributions := flag.String(
		"score_distributions", "../data/score-distributions.gob.gz",
		"File with score distributions to load")
	scoreToBeat := flag.Int("score_to_beat", 300, "Score to beat")
	flag.Parse()

	err := rpio.Open()
	if err != nil {
		glog.Fatal(err)
	}
	defer rpio.Close()

	glog.Info("Loading score distributions table")
	highScoreStrat := yahtzee.NewStrategy(yahtzee.NewScoreDistribution())
	err = highScoreStrat.LoadCache(*scoreDistributions)
	if err != nil {
		glog.Fatal(err)
	}

	detector := detector.NewDetector()
	controller := controller.NewYahtzeeController(wiringConfig)
	player := player.NewYahtzeePlayer(detector, controller)
	player.Play(highScoreStrat, *scoreToBeat)
}

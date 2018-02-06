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
	uri := flag.String("uri", "http://localhost:8080", "URI of Yahtzee server")
	scoreToBeat := flag.Int("score_to_beat", 0, "Score to beat (if 0, maximize expected score)")
	flag.Parse()

	glog.Info("Initializing webcam detector")
	detector, err := detector.NewYahtzeeDetector()
	if err != nil {
		glog.Fatal(err)
	}
	defer detector.Close()

	glog.Info("Initializing GPIO controls")
	controller, err := controller.NewYahtzeeController(controller.DefaultWiringConfig)
	if err != nil {
		glog.Fatal(err)
	}
	defer controller.Close()

	glog.Info("Playing game")
	player := player.NewYahtzeePlayer(detector, controller)
	player.Play(highScoreStrat, *scoreToBeat)
}

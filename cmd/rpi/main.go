package main

import (
	"flag"
	"time"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee/client"
	"github.com/timpalpant/yahtzee/rpi"
	"github.com/timpalpant/yahtzee/rpi/controller"
	"github.com/timpalpant/yahtzee/rpi/detector"
)

func main() {
	uri := flag.String("uri", "http://localhost:8080", "URI of Yahtzee server")
	scoreToBeat := flag.Int("score_to_beat", 0, "Score to beat (if 0, maximize expected score)")
	flag.Parse()

	client := client.NewClient(*uri)

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

	// Turn Yahtzee on if it is off.
	controller.Roll()
	// Wait for roll to complete in case it was on.
	time.Sleep(3 * time.Second)

	for {
		glog.Info("Playing game")
		player := rpi.NewYahtzeePlayer(detector, client, controller)
		if err := player.Play(*scoreToBeat); err != nil {
			panic(err)
		}
	}
}

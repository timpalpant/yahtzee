package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee/client"
	"github.com/timpalpant/yahtzee/rpi"
	"github.com/timpalpant/yahtzee/rpi/controller"
	"github.com/timpalpant/yahtzee/rpi/detector"
)

var stdin = bufio.NewReader(os.Stdin)

func main() {
	dev := flag.String("d", "/dev/video0", "Video device to use")
	imageDir := flag.String("imagedir", "images", "Output directory for captured images")
	uri := flag.String("uri", "http://localhost:8080", "URI of Yahtzee server")
	scoreToBeat := flag.Int("score_to_beat", 0, "Score to beat (if 0, maximize expected score)")
	playContinuously := flag.Bool("play_continuously", false, "Continue to next game automatically")
	flag.Parse()

	client := client.NewClient(*uri)

	glog.Info("Initializing webcam detector")
	detector, err := detector.NewYahtzeeDetector(*dev, *imageDir)
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

	var result string
	for result != "q" {
		glog.Info("Playing game")
		player := rpi.NewYahtzeePlayer(detector, client, controller)
		if err = player.Play(*scoreToBeat); err != nil {
			panic(err)
		}

		if !(*playContinuously) {
			fmt.Print("Press ENTER to play again, q to quit")
			result, err = stdin.ReadString('\n')
			if err != nil {
				panic(err)
			}
		}
	}
}

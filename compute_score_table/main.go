package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

func main() {
	output := flag.String("output", "scores.txt", "Output file")
	flag.Parse()

	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	game := yahtzee.GameState{}
	glog.Infof("Expected score: %f\n", yahtzee.ExpectedScore(game))

	glog.Infof("Saving game state score table to %v", *output)
	f, err := os.Create(*output)
	if err != nil {
		glog.Fatal(err)
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()
	for gameHash, expectedScore := range yahtzee.ExpectedScoreCache {
		if expectedScore == 0 {
			continue
		}

		fmt.Fprintln(buf, gameHash, expectedScore)
	}
}

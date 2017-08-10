package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/timpalpant/yahtzee"
)

func main() {
	flag.Parse()

	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	newGame := &yahtzee.GameState{}
	fmt.Println(yahtzee.ExpectedScore(newGame))
}

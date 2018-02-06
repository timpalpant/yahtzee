package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/server"
)

var stdin = bufio.NewReader(os.Stdin)

func prompt(msg string) string {
	fmt.Print(msg)
	result, err := stdin.ReadString('\n')
	if err != nil {
		panic(err)
	}

	return strings.TrimRight(result, "\n")
}

func parseRoll(s string) (yahtzee.Roll, error) {
	roll := yahtzee.NewRoll()

	i, err := strconv.Atoi(s)
	if err != nil {
		return roll, err
	}

	for ; i > 0; i /= 10 {
		die := i % 10
		roll = roll.Add(die)
	}

	if roll.NumDice() != yahtzee.NDice {
		return roll, fmt.Errorf("Invalid number of dice: %v != %v",
			roll.NumDice(), yahtzee.NDice)
	}

	return roll, nil
}

func promptRoll() yahtzee.Roll {
	for {
		rollStr := prompt("Enter roll: ")
		roll, err := parseRoll(rollStr)
		if err != nil {
			fmt.Printf("Invalid roll: %v\n", err)
			continue
		}

		return roll
	}
}

func getOptimalMove(uri string, game yahtzee.GameState, step server.TurnStep, roll yahtzee.Roll, scoreToBeat int) (*server.OptimalMoveResponse, error) {
	req := &server.OptimalMoveRequest{
		GameState: server.FromYahtzeeGameState(game),
		TurnState: server.TurnState{
			Step: step,
			Dice: roll.Dice(),
		},
		ScoreToBeat: scoreToBeat,
	}

	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	if err := enc.Encode(req); err != nil {
		return nil, err
	}

	endpoint := uri + "/rest/v1/optimal_move"
	resp, err := http.Post(endpoint, "application/json; charset=utf-8", b)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned: %v", resp.Status)
	}

	result := &server.OptimalMoveResponse{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(result); err != nil {
		return nil, err
	}

	return result, err
}

func playGame(uri string, scoreToBeat int) {
	fmt.Println("Welcome to YAHTZEE!")
	game := yahtzee.NewGame()
	var currentScore int

	for !game.GameOver() {
		roll1 := promptRoll()
		resp1, err := getOptimalMove(uri, game, server.Hold1, roll1, scoreToBeat)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Best option is to hold: %v, value: %g\n",
			resp1.HeldDice, resp1.Value)

		roll2 := promptRoll()
		resp2, err := getOptimalMove(uri, game, server.Hold2, roll2, scoreToBeat)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Best option is to hold: %v, value: %g\n",
			resp2.HeldDice, resp2.Value)

		roll3 := promptRoll()
		resp3, err := getOptimalMove(uri, game, server.FillBox, roll3, scoreToBeat)
		if err != nil {
			fmt.Println(err)
			continue
		}

		box := yahtzee.Box(resp3.BoxFilled)
		var addValue int
		game, addValue = game.FillBox(box, roll3)
		currentScore += addValue
		if resp3.NewGame {
			fmt.Printf("P = %g; best option is to give up and start a new game\n\n", resp3.Value)
			return
		} else {
			fmt.Printf("Best option is to play: %v for %v points, final value: %g\n",
				box, addValue, resp3.Value)
		}
	}

	fmt.Printf("Game over! Final score: %v\n", currentScore)
}

func main() {
	uri := flag.String("uri", "http://localhost:8080", "URI of Yahtzee server")
	scoreToBeat := flag.Int("score_to_beat", 0, "High score to try to beat")
	flag.Parse()

	for {
		playGame(*uri, *scoreToBeat)
	}
}

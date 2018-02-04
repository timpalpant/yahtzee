package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/timpalpant/yahtzee/rpi/controller"
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

func main() {
	controller, err := controller.NewYahtzeeController(controller.DefaultWiringConfig)
	if err != nil {
		panic(err)
	}
	defer controller.Close()

	fmt.Println("Play YAHTZEE!")
	fmt.Println("N - New game")
	fmt.Println("L - Left")
	fmt.Println("R - Right")
	fmt.Println("X - Roll")
	fmt.Println("E - Enter")
	for i := 1; i <= 5; i++ {
		fmt.Printf("%d - Hold die %d\n", i, i)
	}

	for {
		choice := strings.ToUpper(prompt("Enter a selection: "))
		switch choice {
		case "N":
			controller.NewGame()
		case "L":
			controller.Left(1)
		case "R":
			controller.Right(1)
		case "X":
			controller.Roll()
		case "E":
			controller.Enter()
		case "1":
			fallthrough
		case "2":
			fallthrough
		case "3":
			fallthrough
		case "4":
			fallthrough
		case "5":
			die, err := strconv.Atoi(choice)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if err := controller.Hold(die-1); err != nil {
				fmt.Println(err)
			}
		default:
			fmt.Println("Invalid choice")
		}
	}
}

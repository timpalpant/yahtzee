package rpi

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/client"
	"github.com/timpalpant/yahtzee/rpi/controller"
	"github.com/timpalpant/yahtzee/rpi/detector"
)

type YahtzeePlayer struct {
	detector   *detector.YahtzeeDetector
	client     *client.Client
	controller *controller.YahtzeeController

	game         yahtzee.GameState
	turnStep     yahtzee.TurnStep
	held         []bool
	currentScore int
}

func NewYahtzeePlayer(detector *detector.YahtzeeDetector,
	client *client.Client, controller *controller.YahtzeeController) *YahtzeePlayer {
	return &YahtzeePlayer{
		detector:   detector,
		client:     client,
		controller: controller,
		game:       yahtzee.NewGame(),
		turnStep:   yahtzee.Hold1,
		held:       make([]bool, yahtzee.NDice),
	}
}

func (yp *YahtzeePlayer) Play(scoreToBeat int) error {
	yp.controller.NewGame()

	for !yp.game.GameOver() {
		time.Sleep(time.Second)
		yp.controller.Roll()
		// Wait for roll to complete.
		time.Sleep(6 * time.Second)
		roll, err := yp.detector.GetCurrentRoll()
		if err != nil {
			return err
		}

		resp, err := yp.client.GetOptimalMove(yp.game, yp.turnStep, roll, scoreToBeat)
		if err != nil {
			return err
		}

		switch yp.turnStep {
		case yahtzee.Hold1:
			fallthrough
		case yahtzee.Hold2:
			if len(resp.HeldDice) < yahtzee.NDice {
				glog.Infof("Best option is to hold: %v, value: %g",
					resp.HeldDice, resp.Value)
				if err := yp.hold(roll, resp.HeldDice); err != nil {
					return err
				}

				break
			}

			// Hold all dice, i.e. skip to fill box.
			resp, err = yp.client.GetOptimalMove(yp.game, yahtzee.FillBox, roll, scoreToBeat)
			if err != nil {
				return err
			}
			fallthrough
		case yahtzee.FillBox:
			box := yahtzee.Box(resp.BoxFilled)
			yp.fillBox(box, roll)
		}

		if resp.NewGame {
			glog.Info("Giving up and starting a new game")
			break
		}
	}

	return nil
}

func (yp *YahtzeePlayer) hold(roll, diceToKeep []int) error {
	if len(roll) != len(yp.held) {
		return fmt.Errorf(
			"wrong number of dice in roll: expected: %v, got: %v",
			len(yp.held), len(roll))
	}

	keepSet := make(map[int]int, len(diceToKeep))
	for _, side := range diceToKeep {
		keepSet[side]++
	}

	// TODO: Minimize changes.
	desired := make([]bool, len(roll))
	for die := range roll {
		side := roll[die]
		if n := keepSet[side]; n > 0 {
			glog.V(2).Infof("Keeping die %v with side %v", die, side)
			desired[die] = true
			keepSet[side]--
		}
	}

	glog.V(1).Infof("Current held dice: %v", yp.held)
	glog.V(1).Infof("Desired held dice: %v", desired)

	for die := range yp.held {
		if yp.held[die] != desired[die] {
			if err := yp.controller.Hold(die); err != nil {
				return err
			}

			yp.held[die] = desired[die]
		}
	}

	yp.turnStep++
	return nil
}

func (yp *YahtzeePlayer) fillBox(box yahtzee.Box, dice []int) {
	roll := yahtzee.NewRollFromDice(dice)
	game, addValue := yp.game.FillBox(box, roll)
	glog.Infof("Best option is to play: %v for %v points", box, addValue)

	buttonPressSequence := yp.computeFillPresses(box, roll)
	yp.controller.Perform(buttonPressSequence)

	yp.currentScore += addValue
	yp.game = game
	// Held dice reset for the next turn when box is filled.
	for die := range yp.held {
		yp.held[die] = false
	}
}

func (yp *YahtzeePlayer) computeFillPresses(box yahtzee.Box, roll yahtzee.Roll) []controller.YahtzeeButton {
	result := make([]controller.YahtzeeButton, 0)

	// If roll is the first Yahtzee then the Yahtzee box is highlighted automatically.
	if yahtzee.IsYahtzee(roll) && !yp.game.BonusEligible() {
		result = append(result, controller.Enter)
		return result
	}

	// Cursor initializes automatically if we are in the fill box step.
	// Otherwise, we need to trigger it by pushing right/left once.
	if yp.turnStep != yahtzee.FillBox {
		result = append(result, controller.Right)
	}

	moves := getMovesToBox(yp.game, box)
	result = append(result, moves...)

	result = append(result, controller.Enter)
	return result
}

func getMovesToBox(game yahtzee.GameState, desired yahtzee.Box) []controller.YahtzeeButton {
	result := make([]controller.YahtzeeButton, 0)

	// TODO: Go left when it is more efficient.
	for _, box := range game.AvailableBoxes() {
		if box == desired {
			break
		}

		result = append(result, controller.Right)
	}

	return result
}

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
	prevRoll     []int
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
		glog.Infof("Turn %d, step %v, current score: %v", yp.game.Turn(), yp.turnStep, yp.currentScore)
		time.Sleep(time.Second)
		yp.controller.Roll()
		// Wait for roll to complete.
		time.Sleep(6 * time.Second)
		roll, err := yp.detector.GetCurrentRoll()
		if err != nil {
			return err
		} else if err := yp.checkUnexpectedRollChange(roll); err != nil {
			glog.Warning(err)
		}

		glog.Infof("Detected roll: %v", roll)
		resp, err := yp.client.GetOptimalMove(yp.game, yp.turnStep, roll, scoreToBeat)
		if err != nil {
			return err
		}

		if resp.NewGame {
			glog.Info("Giving up and starting a new game")
			break
		}

		switch yp.turnStep {
		case yahtzee.Hold1:
			fallthrough
		case yahtzee.Hold2:
			if len(resp.HeldDice) == yahtzee.NDice {
				if err := yp.fillBoxEarly(roll, scoreToBeat); err != nil {
					return err
				}
			} else {
				glog.Infof("Best option is to hold: %v, value: %g",
					resp.HeldDice, resp.Value)
				if err := yp.hold(roll, resp.HeldDice); err != nil {
					return err
				}
			}
		case yahtzee.FillBox:
			box := yahtzee.Box(resp.BoxFilled)
			yp.fillBox(box, roll)
		}

		yp.prevRoll = roll
	}

	glog.Infof("Final score: %v", yp.currentScore)
	return nil
}

func (yp *YahtzeePlayer) checkUnexpectedRollChange(roll []int) error {
	if yp.lastRoll == nil {
		return nil
	}

	if len(roll) != len(held) {
		return fmt.Errorf(
			"unexpectednumber of dice in roll: got %v, expected %v",
			len(roll), len(held))
	}

	for die, held := range yp.held {
		if held {
			if roll[die] != yp.lastRoll[die] {
				return fmt.Errorf(
					"unexpected roll change: die %v [HELD], was: %v, now: %v",
					die, yp.lastRoll[die], roll[die])
			}
		}
	}
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

func (yp *YahtzeePlayer) fillBoxEarly(roll []int, scoreToBeat int) error {
	// Hold all dice, i.e. skip to fill box.
	resp, err := yp.client.GetOptimalMove(yp.game, yahtzee.FillBox, roll, scoreToBeat)
	if err != nil {
		return err
	}

	box := yahtzee.Box(resp.BoxFilled)
	yp.fillBox(box, roll)
	return nil
}

func (yp *YahtzeePlayer) fillBox(box yahtzee.Box, dice []int) {
	roll := yahtzee.NewRollFromDice(dice)
	game, addValue := yp.game.FillBox(box, roll)
	glog.Infof("Best option is to play: %v for %v points", box, addValue)

	// Last box plays itself automatically.
	if len(yp.game.AvailableBoxes()) > 1 {
		buttonPressSequence := yp.computeFillPresses(box, roll)
		yp.controller.Perform(buttonPressSequence)
	}

	// Next turn. Note: Held dice reset.
	yp.currentScore += addValue
	yp.game = game
	yp.turnStep = yahtzee.Hold1
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
	glog.V(1).Infof("Moving right %d times to select %v", len(moves), box)
	result = append(result, moves...)

	result = append(result, controller.Enter)
	return result
}

func getMovesToBox(game yahtzee.GameState, desired yahtzee.Box) []controller.YahtzeeButton {
	glog.V(2).Infof("Computing moves to select desired box %v", desired)
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

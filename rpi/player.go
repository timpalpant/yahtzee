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
	gameValue, err := yp.client.GetGameValue(yp.game, scoreToBeat)
	if err != nil {
		return err
	}

	sleep := 3 * time.Second
	for !yp.game.GameOver() {
		glog.Infof("Turn %d, step %v, current score: %v", yp.game.Turn(), yp.turnStep, yp.currentScore)
		yp.controller.Roll()
		// Wait for roll to complete.
		time.Sleep(sleep)
		roll, err := yp.detector.GetCurrentRoll()
		if err != nil {
			return err
		} else if err := yp.checkUnexpectedRollChange(roll); err != nil {
			glog.Warning(err)
		}

		glog.Infof("Detected roll: %v", roll)
		remainingScore := scoreToBeat - yp.currentScore
		resp, err := yp.client.GetOptimalMove(yp.game, yp.turnStep, roll, remainingScore)
		if err != nil {
			return err
		}

		if scoreToBeat > 0 && resp.Value < gameValue {
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

				nHeld := len(resp.HeldDice)
				// Discount sleep time based on number of dice held.
				sleep = 3*time.Second - time.Duration(1e9*float64(nHeld)/5.0)
			}
		case yahtzee.FillBox:
			box := yahtzee.Box(resp.BoxFilled)
			scoreAdded := yp.fillBox(box, roll)
			// Sleep extra long proportionally to score to be added.
			sleep = 4*time.Second + time.Duration(1e9*float64(scoreAdded)/50.0)
		}

		yp.prevRoll = roll
	}

	glog.Infof("Final score: %v", yp.currentScore)
	return nil
}

func (yp *YahtzeePlayer) checkUnexpectedRollChange(roll []int) error {
	if yp.prevRoll == nil {
		return nil
	}

	if len(roll) != len(yp.held) {
		return fmt.Errorf(
			"unexpectednumber of dice in roll: got %v, expected %v",
			len(roll), len(yp.held))
	}

	for die, held := range yp.held {
		if held {
			if roll[die] != yp.prevRoll[die] {
				return fmt.Errorf(
					"unexpected roll change: die %v [HELD], was: %v, now: %v",
					die, yp.prevRoll[die], roll[die])
			}
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

func (yp *YahtzeePlayer) fillBox(box yahtzee.Box, dice []int) int {
	roll := yahtzee.NewRollFromDice(dice)
	game, addValue := yp.game.FillBox(box, roll)
	glog.Infof("Best option is to play: %v for %v points", box, addValue)

	// Last box plays itself automatically.
	if len(yp.game.AvailableBoxes()) > 1 {
		buttonPressSequence := yp.computeFillPresses(box, roll)
		yp.controller.Perform(buttonPressSequence)
		// Need a small sleep after Enter, otherwise Roll press is not
		// detected correctly.
		time.Sleep(200 * time.Millisecond)
	}

	// Next turn. Note: Held dice reset.
	yp.currentScore += addValue
	yp.game = game
	yp.turnStep = yahtzee.Hold1
	for die := range yp.held {
		yp.held[die] = false
	}

	return addValue
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

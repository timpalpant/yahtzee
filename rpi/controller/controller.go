package controller

import (
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio"
)

// Button represents a single button on the electronic
// hand-held Yahtzee game.
type Button struct {
	rpio.Pin
}

func NewButton(pin int) Button {
	return Button{rpio.Pin(pin)}
}

// Press pushes and releases the button.
func (btn Button) Press() {
	btn.High()
	time.Sleep(50 * time.Millisecond)
	btn.Low()
}

// YahtzeeControllerConfig configures the mapping between GPIO
// pins and the buttons they are wired to on the controller. Note
// that these are BCM2835 pinouts, see https://godoc.org/github.com/stianeikeland/go-rpio
// for more information.
type YahtzeeControllerConfig struct {
	HoldButtonPins [5]int
	NewGamePin     int
	RollPin        int
	LeftPin        int
	RightPin       int
	EnterPin       int
}

// YahtzeeController encapsulates the behavior used to control the Yahtzee game.
type YahtzeeController struct {
	holdButtons   [5]Button
	newGameButton Button
	rollButton    Button
	leftButton    Button
	rightButton   Button
	enterButton   Button
}

func NewYahtzeeController(config YahtzeeControllerConfig) *YahtzeeController {
	controller := &YahtzeeController{
		holdButtons: [5]Button{
			NewButton(config.HoldButtonPins[0]),
			NewButton(config.HoldButtonPins[1]),
			NewButton(config.HoldButtonPins[2]),
			NewButton(config.HoldButtonPins[3]),
			NewButton(config.HoldButtonPins[4]),
		},
		newGameButton: NewButton(config.NewGamePin),
		rollButton:    NewButton(config.RollPin),
		leftButton:    NewButton(config.LeftPin),
		rightButton:   NewButton(config.RightPin),
		enterButton:   NewButton(config.EnterPin),
	}

	for _, btn := range controller.holdButtons {
		btn.Output()
	}
	controller.newGameButton.Output()
	controller.rollButton.Output()
	controller.enterButton.Output()

	return controller
}

func (yc *YahtzeeController) Hold(die int) error {
	if die < 0 || die >= len(yc.holdButtons) {
		return fmt.Errorf("die must be between 0 - %d", len(yc.holdButtons))
	}

	yc.holdButtons[die].Press()
	return nil
}

func (yc *YahtzeeController) NewGame() {
	yc.newGameButton.Press()
}

func (yc *YahtzeeController) Right(n int) {
	for i := 0; i < n; i++ {
		yc.rightButton.Press()
	}
}

func (yc *YahtzeeController) Left(n int) {
	for i := 0; i < n; i++ {
		yc.leftButton.Press()
	}
}

func (yc *YahtzeeController) Enter() {
	yc.enterButton.Press()
}

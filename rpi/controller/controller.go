package controller

import (
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio"
)

const buttonPressTime = 50 * time.Millisecond

var DefaultWiringConfig = YahtzeeControllerConfig{
	HoldButtonPins: [5]int{22, 27, 23, 15, 18},
	NewGamePin:     14,
	RollPin:        4,
	LeftPin:        10,
	RightPin:       9,
	EnterPin:       11,
}

// Button represents a single button on the electronic
// hand-held Yahtzee game.
type Button struct {
	rpio.Pin
}

func NewButton(pin int) Button {
	p := rpio.Pin(pin)
	p.Output()
	p.High()  // Relays are normally-open.
	return Button{p}
}

// Press pushes and releases the button.
func (btn Button) Press() {
	btn.Low()
	time.Sleep(buttonPressTime)
	btn.High()
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

func NewYahtzeeController(config YahtzeeControllerConfig) (*YahtzeeController, error) {
	err := rpio.Open()
	if err != nil {
		return nil, err
	}

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

	return controller, nil
}

func (yc *YahtzeeController) Close() {
	rpio.Close()
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

func (yc *YahtzeeController) Roll() {
	yc.rollButton.Press()
}

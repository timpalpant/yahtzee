package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/stianeikeland/go-rpio"
)

const defaultButtonPressDuration = 100 * time.Millisecond

// YahtzeeButton is an enum of the available buttons on the
// YahtzeeController.
type YahtzeeButton int

const (
	Hold1 YahtzeeButton = iota
	Hold2
	Hold3
	Hold4
	Hold5
	NewGame
	Left
	Right
	Enter
	Roll
)

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
	p.High() // Relays are normally-open.
	return Button{p}
}

// Press pushes and releases the button, holding it for the specified duration.
func (btn Button) Press(d time.Duration) {
	btn.Low()
	time.Sleep(d)
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

	glog.V(1).Infof("Holding die %v", die+1)
	yc.holdButtons[die].Press(defaultButtonPressDuration)
	return nil
}

func (yc *YahtzeeController) NewGame() {
	// NOTE: New game button has to be held down for longer
	// to be recognized as a press.
	glog.V(1).Infof("Pressing new game")
	yc.newGameButton.Press(time.Second)
}

func (yc *YahtzeeController) Right(n int) {
	glog.V(1).Infof("Moving right %v", n)
	for i := 0; i < n; i++ {
		yc.rightButton.Press(defaultButtonPressDuration)
	}
}

func (yc *YahtzeeController) Left(n int) {
	glog.V(1).Infof("Moving left %v", n)
	for i := 0; i < n; i++ {
		yc.leftButton.Press(defaultButtonPressDuration)
	}
}

func (yc *YahtzeeController) Enter() {
	glog.V(1).Infof("Pressing enter")
	yc.enterButton.Press(defaultButtonPressDuration)
}

func (yc *YahtzeeController) Roll() {
	glog.V(1).Infof("Pressing roll")
	yc.rollButton.Press(defaultButtonPressDuration)
}

func (yc *YahtzeeController) Press(btn YahtzeeButton) error {
	switch btn {
	case Hold1:
		return yc.Hold(0)
	case Hold2:
		return yc.Hold(1)
	case Hold3:
		return yc.Hold(2)
	case Hold4:
		return yc.Hold(3)
	case Hold5:
		return yc.Hold(4)
	case NewGame:
		yc.NewGame()
	case Left:
		yc.Left(1)
	case Right:
		yc.Right(1)
	case Enter:
		yc.Enter()
	default:
		return fmt.Errorf("unsupported button: %v", btn)
	}

	return nil
}

// Execute the given sequence of button presses.
func (yc *YahtzeeController) Perform(buttonPressSequence []YahtzeeButton) error {
	for i, btn := range buttonPressSequence {
		if err := yc.Press(btn); err != nil {
			return err
		}

		time.Sleep(200 * time.Millisecond)
		// NOTE: Longer wait if pressing the same button repeatedly
		// to ensure that relay responds and we don't miss a press.
		if i > 0 && buttonPressSequence[i-1] == btn {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}

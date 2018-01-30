package player

import (
	"github.com/timpalpant/yahtzee"
)

type YahtzeePlayer struct {
	detector *detector.YahtzeeDetector
	controller *controller.YahtzeeController
}

func NewYahtzeePlayer(detector *detector.YahtzeeDetector, controller *controller.YahtzeeController) *YahtzeePlayer {
	return &YahtzeePlayer{detector, controller}
}

func (yp *YahtzeePlayer) Play(strategy *yahtzee.Strategy, scoreToBeat int) error {

}

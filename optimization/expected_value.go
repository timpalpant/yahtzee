package optimization

import (
	"encoding/gob"
	"fmt"

	"github.com/timpalpant/yahtzee"
)

func init() {
	gob.Register(ExpectedValue(0))
}

// ExpectedValue implements GameResult, and represents
// maximizing your expected score.
type ExpectedValue float32

func NewExpectedValue() ExpectedValue {
	return ExpectedValue(0)
}

func (ew ExpectedValue) ScoreDependent() bool {
	return false
}

func (ev ExpectedValue) New() GameResult {
	return NewExpectedValue()
}

func (ev ExpectedValue) GameValue(game yahtzee.GameState) GameResult {
	if !game.GameOver() {
		panic(fmt.Errorf("trying to get endgame value of game that is not over: %v", game))
	}

	return ExpectedValue(0)
}

func (ev ExpectedValue) CopyInto(other GameResult) GameResult {
	return ev
}

func (ev ExpectedValue) Zero(other GameResult) GameResult {
	return NewExpectedValue()
}

func (ev ExpectedValue) Add(other GameResult, weight float32) GameResult {
	otherEV := other.(ExpectedValue)
	return ev + ExpectedValue(weight)*otherEV
}

func (ev ExpectedValue) Max(other GameResult) GameResult {
	otherEV := other.(ExpectedValue)
	if otherEV > ev {
		return otherEV
	}

	return ev
}

func (ev ExpectedValue) Shift(offset int) GameResult {
	return ev + ExpectedValue(offset)
}

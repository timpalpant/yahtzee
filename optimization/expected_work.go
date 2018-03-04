package optimization

import (
	"encoding/gob"
	"fmt"

	"github.com/timpalpant/yahtzee"
)

func init() {
	gob.Register(ExpectedWork{})
}

// ExpectedWork implements GameResult, and represents
// minimizing the work to achiewe a particular score. Unlike
// ExpectedWork, it computes the result for only a single score.
type ExpectedWork struct {
	ScoreToBeat int
	Value       float32
}

func NewExpectedWork(scoreToBeat int, e0 float32) *ExpectedWork {
	return &ExpectedWork{scoreToBeat, e0}
}

func (ew *ExpectedWork) ScoreDependent() bool {
	return true
}

func (ew *ExpectedWork) New() GameResult {
	return &ExpectedWork{ew.ScoreToBeat, ew.Value}
}

func (ew *ExpectedWork) GameValue(game yahtzee.GameState) GameResult {
	if !game.GameOver() {
		panic(fmt.Errorf("trying to get endgame value of game that is not over: %v", game))
	}

	if game.TotalScore() >= ew.ScoreToBeat {
		// Achieved desired score, no additional work required.
		return &ExpectedWork{ew.ScoreToBeat, 0}
	} else {
		// Didn't achieve desired score, have to start over.
		return &ExpectedWork{ew.ScoreToBeat, ew.Value}
	}
}

func (ew *ExpectedWork) CopyInto(other GameResult) GameResult {
	otherEW := other.(*ExpectedWork)
	otherEW.Value = ew.Value
	return otherEW
}

func (ew *ExpectedWork) Zero(other GameResult) GameResult {
	otherEW := other.(*ExpectedWork)
	otherEW.Value = 1 // +1 work for every roll.
	return otherEW
}

func (ew *ExpectedWork) Add(other GameResult, weight float32) GameResult {
	otherEW := other.(*ExpectedWork)
	ew.Value += weight * otherEW.Value
	return ew
}

func (ew *ExpectedWork) Max(other GameResult) GameResult {
	otherEW := other.(*ExpectedWork)
	if otherEW.Value < ew.Value {
		ew.Value = otherEW.Value
	}
	return ew
}

func (ew *ExpectedWork) Shift(offset int) GameResult {
	ew.Value += float32(offset)
	return ew
}

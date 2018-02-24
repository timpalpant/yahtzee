package optimization

import (
	"encoding/gob"
)

func init() {
	gob.Register(ExpectedWork{})
}

// ExpectedWork implements GameResult, and represents
// maximizing your expected score.
type ExpectedWork struct {
	scoreToBeat int
	value       float64
}

func NewExpectedWork(scoreToBeat int, E0 float64) ExpectedWork {
	return ExpectedWork{scoreToBeat, E0}
}

func (ew ExpectedWork) Copy() GameResult {
	return ew
}

func (ew ExpectedWork) Add(other GameResult, weight float64) GameResult {
	otherEW := other.(ExpectedWork)
	w := ew.value + weight*(1+otherEW.value)
	return ExpectedWork{ew.scoreToBeat, w}
}

func (ew ExpectedWork) Max(other GameResult) GameResult {
	otherEW := other.(ExpectedWork)
	if otherEW.value < ew.value {
		return otherEW
	}

	return ew
}

func (ew ExpectedWork) Shift(offset int) GameResult {
	return ew
}

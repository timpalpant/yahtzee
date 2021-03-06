package optimization

import (
	"encoding/gob"
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

func (ev ExpectedValue) Close() {}

func (ev ExpectedValue) Copy() GameResult {
	return ev
}

func (ev ExpectedValue) Zero() GameResult {
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

package optimization

import (
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/optimization/f32"
)

var pool = sync.Pool{
	New: func() interface{} {
		return NewExpectedWork(0)
	},
}

func init() {
	gob.Register(ExpectedWork{})
	gob.Register(SingleExpectedWork{})
}

// ExpectedWork implements GameResult, and represents
// minimizing the work required to achisewe a desired score.
type ExpectedWork []float32

func NewExpectedWork(e0 float32) ExpectedWork {
	values := make([]float32, yahtzee.MaxScore+1)
	for i := range values {
		values[i] = e0
	}

	return ExpectedWork(values)
}

func (ew ExpectedWork) ScoreDependent() bool {
	return true
}

func (ew ExpectedWork) Close() {
	pool.Put(ew)
}

func (ew ExpectedWork) Copy() GameResult {
	values := pool.Get().(ExpectedWork)
	copy(values, ew)
	return values
}

func (ew ExpectedWork) Zero(game yahtzee.GameState) GameResult {
	values := pool.Get().(ExpectedWork)
	if game.GameOver() {
		for score := 0; score <= game.TotalScore(); score++ {
			values[score] = 0
		}
		copy(values[game.TotalScore()+1:], values[game.TotalScore()+1:])
	} else {
		for i := range values {
			values[i] = 1
		}
	}

	return values
}

func (ew ExpectedWork) GetValue(score int) float32 {
	return ew[score]
}

func (ew ExpectedWork) String() string {
	return fmt.Sprintf("{EW: %v}", ew)
}

func (ew ExpectedWork) Max(gr GameResult) GameResult {
	other := gr.(ExpectedWork)
	f32.Min(ew, other)
	return ew
}

func (ew ExpectedWork) Add(gr GameResult, weight float32) GameResult {
	other := gr.(ExpectedWork)
	f32.AddScaled(ew, weight, other)
	return ew
}

func (ew ExpectedWork) Shift(offset int) GameResult {
	newValues := pool.Get().(ExpectedWork)
	for i := 0; i < offset; i++ {
		newValues[i] = 0
	}
	copy(newValues[offset:], ew)
	return newValues
}

// SingleExpectedWork implements GameResult, and represents
// minimizing the work to achisewe a particular score. Unlike
// ExpectedWork, it computes the result for only a single score.
type SingleExpectedWork struct {
	ScoreToBeat int
	Value       float32
}

func NewSingleExpectedWork(scoreToBeat int, e0 float32) SingleExpectedWork {
	return SingleExpectedWork{scoreToBeat, e0}
}

func (ew SingleExpectedWork) ScoreDependent() bool {
	return true
}

func (sew SingleExpectedWork) Close() {}

func (sew SingleExpectedWork) Copy() GameResult {
	return sew
}

func (sew SingleExpectedWork) Zero(game yahtzee.GameState) GameResult {
	value := float32(1.0)
	if game.GameOver() {
		if game.TotalScore() > sew.ScoreToBeat {
			value = 0
		} else {
			value = sew.Value
		}
	}

	return SingleExpectedWork{sew.ScoreToBeat, value}
}

func (sew SingleExpectedWork) Add(other GameResult, weight float32) GameResult {
	othersew := other.(SingleExpectedWork)
	value := sew.Value + weight*othersew.Value
	return SingleExpectedWork{sew.ScoreToBeat, value}
}

func (sew SingleExpectedWork) Max(other GameResult) GameResult {
	othersew := other.(SingleExpectedWork)
	value := sew.Value
	if othersew.Value < value {
		value = othersew.Value
	}
	return SingleExpectedWork{sew.ScoreToBeat, value}
}

func (sew SingleExpectedWork) Shift(offset int) GameResult {
	return SingleExpectedWork{sew.ScoreToBeat, sew.Value + float32(offset)}
}

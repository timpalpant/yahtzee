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
}

// ExpectedWork implements GameResult, and represents
// minimizing the work required to achieve a desired score.
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

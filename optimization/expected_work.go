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
		return make([]float32, yahtzee.MaxScore)
	},
}

func init() {
	gob.Register(ExpectedWork{})
}

// ExpectedWork implements GameResult, and represents
// minimizing the work required to achieve a desired score.
type ExpectedWork struct {
	// The expected work at the start of a game.
	E0 float32
	// [W(start + 1), W(start + 2), ..., W(start + N)]
	// where W(start + N) is the largest integer such that W(start + N) < E0.
	Values []float32
}

func NewExpectedWork(e0 float32) ExpectedWork {
	values := pool.Get().([]float32)
	for i := range values {
		values[i] = e0
	}

	return ExpectedWork{
		E0:     e0,
		Values: values,
	}
}

func (ew ExpectedWork) Close() {
	pool.Put(ew.Values)
}

func (ew ExpectedWork) Copy() GameResult {
	values := pool.Get().([]float32)
	copy(values, ew.Values)
	return ExpectedWork{
		E0:     ew.E0,
		Values: values,
	}
}

func (ew ExpectedWork) Zero() GameResult {
	values := pool.Get().([]float32)
	for i := range values {
		values[i] = 1
	}

	return ExpectedWork{
		E0:     ew.E0,
		Values: values,
	}
}

func (ew ExpectedWork) String() string {
	return fmt.Sprintf("{Dist: %v}", ew.Values)
}

func (ew ExpectedWork) Max(gr GameResult) GameResult {
	other := gr.(ExpectedWork)
	f32.Min(ew.Values, other.Values)
	return ew
}

func (ew ExpectedWork) Add(gr GameResult, weight float32) GameResult {
	other := gr.(ExpectedWork)
	f32.AddScaled(ew.Values, weight, other.Values)
	return ew
}

func (ew ExpectedWork) Shift(offset int) GameResult {
	newValues := pool.Get().([]float32)
	for i := 0; i < offset; i++ {
		newValues[i] = 0
	}
	copy(newValues[offset:], ew.Values)
	ew.Values = newValues
	return ew
}

package optimization

import (
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/timpalpant/yahtzee"
)

var valuesPool = sync.Pool{
	New: func() interface{} {
		return make([]float64, yahtzee.MaxScore)
	},
}

func init() {
	gob.Register(ExpectedWork{})
}

// ExpectedWork implements GameResult, and represents
// minimizing the work required to achieve a desired score.
type ExpectedWork struct {
	// The expected work at the start of a game.
	E0 float64
	// [W(start + 1), W(start + 2), ..., W(start + N)]
	// where W(start + N) is the largest integer such that W(start + N) < E0.
	Values []float64
}

func NewExpectedWork(e0 float64) ExpectedWork {
	return ExpectedWork{E0: e0}
}

func (ew ExpectedWork) Copy() GameResult {
	values := valuesPool.Get().([]float64)
	for i := range values {
		values[i] = ew.E0
	}

	return ExpectedWork{
		E0:     ew.E0,
		Values: values,
	}
}

func (ew ExpectedWork) Zero() GameResult {
	values := valuesPool.Get().([]float64)
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

func clear(v []float64) {
	for i := range v {
		v[i] = 0
	}
}

func (ew ExpectedWork) Max(gr GameResult) GameResult {
	other := gr.(ExpectedWork)
	newValues := valuesPool.Get().([]float64)
	clear(newValues)
	for s := 0; s < len(newValues); s++ {
		x1 := ew.Values[s]
		x2 := other.Values[s]
		newValues[s] = min(x1, x2)
	}

	if len(ew.Values) == yahtzee.MaxScore {
		valuesPool.Put(ew.Values)
	}

	ew.Values = newValues
	return ew
}

func min(x1, x2 float64) float64 {
	if x1 < x2 {
		return x1
	}

	return x2
}

func (ew ExpectedWork) Add(gr GameResult, weight float64) GameResult {
	other := gr.(ExpectedWork)
	newValues := valuesPool.Get().([]float64)
	copy(newValues, ew.Values)
	for s := 0; s < len(newValues); s++ {
		newValues[s] += weight * other.Values[s]
	}

	valuesPool.Put(ew.Values)
	ew.Values = newValues
	return ew
}

func (ew ExpectedWork) Shift(offset int) GameResult {
	newValues := valuesPool.Get().([]float64)
	for i := 0; i < offset; i++ {
		newValues[i] = 0
	}

	for s := offset; s < len(newValues); s++ {
		newValues[s] = ew.Values[s-offset]
	}

	valuesPool.Put(ew.Values)
	ew.Values = newValues
	return ew
}

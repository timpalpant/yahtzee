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
// minimizing the work required to achieve a desired score.
type ExpectedWork struct {
	// The expected work at the start of a game.
	E0 float64
	// The smallest integer such that W(start) > 0.
	Start int
	// [W(start + 1), W(start + 2), ..., W(start + N)]
	// where W(start + N) is the largest integer such that W(start + N) < E0.
	Values []float64
}

func NewExpectedWork(e0 float64) ExpectedWork {
	return ExpectedWork{E0: e0}
}

func (ew ExpectedWork) Copy() GameResult {
	var vCopy []float64
	if len(ew.Values) > 0 {
		vCopy = make([]float64, len(ew.Values))
		copy(vCopy, ew.Values)
	}

	return ExpectedWork{
		E0:     ew.E0,
		Start:  ew.Start,
		Values: vCopy,
	}
}

func (ew ExpectedWork) Zero() GameResult {
	values := make([]float64, yahtzee.MaxScore)
	for i := range values {
		values[i] = 1
	}

	return ExpectedWork{
		E0:     ew.E0,
		Values: values,
	}
}

func (ew ExpectedWork) Stop() int {
	return ew.Start + len(ew.Values)
}

func (ew ExpectedWork) GetValue(score int) float64 {
	if score <= ew.Start {
		return 0
	}

	idx := score - ew.Start
	if idx >= len(ew.Values) {
		return ew.E0
	}

	return ew.Values[idx]
}

func (ew ExpectedWork) String() string {
	return fmt.Sprintf("{Start: %v, Stop: %v, Dist: %v}",
		ew.Start, ew.Stop(), ew.Values)
}

func (ew ExpectedWork) Max(gr GameResult) GameResult {
	other := gr.(ExpectedWork)
	newStart := ew.Start
	if other.Start > newStart {
		newStart = other.Start
	}

	newStop := ew.Stop()
	if other.Stop() > newStop {
		newStop = other.Stop()
	}

	newValues := make([]float64, newStop-newStart)
	for s := newStart; s < newStop; s++ {
		x1 := ew.GetValue(s)
		x2 := other.GetValue(s)
		newValues[s-newStart] = min(x1, x2)

		// NOTE: Uses the fact that expected work is monotonic.
		if newValues[s-newStart] >= ew.E0 {
			newValues = newValues[:s-newStart]
			break
		}
	}

	return ExpectedWork{
		E0:     ew.E0,
		Start:  newStart,
		Values: newValues,
	}
}

func min(x1, x2 float64) float64 {
	if x1 < x2 {
		return x1
	}

	return x2
}

func (ew ExpectedWork) Add(gr GameResult, weight float64) GameResult {
	other := gr.(ExpectedWork)
	newStart := ew.Start
	if other.Start < newStart {
		newStart = other.Start
	}

	newStop := yahtzee.MaxScore
	newValues := make([]float64, newStop-newStart)
	copy(newValues[ew.Start-newStart:], ew.Values)
	for s := other.Start; s < yahtzee.MaxScore; s++ {
		newValues[s-newStart] += weight * other.GetValue(s)

		// NOTE: Uses the fact that expected work is monotonic.
		if newValues[s-newStart] >= ew.E0 {
			newValues = newValues[:s-newStart]
			break
		}
	}

	return ExpectedWork{
		E0:     ew.E0,
		Start:  ew.Start,
		Values: newValues,
	}
}

func (ew ExpectedWork) Shift(offset int) GameResult {
	return ExpectedWork{
		E0:     ew.E0,
		Start:  ew.Start + offset,
		Values: ew.Values,
	}
}

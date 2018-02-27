package optimization

import (
	"encoding/gob"
	"fmt"
)

func init() {
	gob.Register(ScoreDistribution{})
}

// ScoreDistribution implements GameResult, and represents
// maximizing your probability of obtaining score s for all s.
type ScoreDistribution struct {
	// The largest integer such that P(start) == 1.
	Start int
	// [P(start + 1), P(start + 2), ..., P(start + N)]
	// where P(start + N) is the largest integer such that P(s1) > 0.
	Probabilities []float32
}

func NewScoreDistribution() ScoreDistribution {
	return ScoreDistribution{}
}

func (sd ScoreDistribution) Close() {}

func (sd ScoreDistribution) Copy() GameResult {
	pCopy := make([]float32, len(sd.Probabilities))
	copy(pCopy, sd.Probabilities)

	return ScoreDistribution{
		Start:         sd.Start,
		Probabilities: pCopy,
	}
}

func (sd ScoreDistribution) Zero() GameResult {
	return NewScoreDistribution()
}

func (sd ScoreDistribution) Stop() int {
	return sd.Start + len(sd.Probabilities)
}

func (sd ScoreDistribution) GetProbability(score int) float32 {
	if score <= sd.Start {
		return 1
	}

	idx := score - sd.Start
	if idx >= len(sd.Probabilities) {
		return 0
	}

	return sd.Probabilities[idx]
}

func (sd ScoreDistribution) String() string {
	return fmt.Sprintf("{Start: %v, Stop: %v, Dist: %v}",
		sd.Start, sd.Stop(), sd.Probabilities)
}

func (sd ScoreDistribution) Max(gr GameResult) GameResult {
	other := gr.(ScoreDistribution)
	maxStart := sd.Start
	if other.Start > maxStart {
		maxStart = other.Start
	}

	maxStop := sd.Stop()
	if other.Stop() > maxStop {
		maxStop = other.Stop()
	}

	newProbabilities := make([]float32, maxStop-maxStart)
	for s := maxStart; s < maxStop; s++ {
		//x1 := sd.Probabilities[s-sd.Start]
		//x2 := other.Probabilities[s-other.Start]
		x1 := sd.GetProbability(s)
		x2 := other.GetProbability(s)
		newProbabilities[s-maxStart] = max(x1, x2)
	}

	return ScoreDistribution{
		Start:         maxStart,
		Probabilities: newProbabilities,
	}
}

func max(x1, x2 float32) float32 {
	if x1 > x2 {
		return x1
	}

	return x2
}

func (sd ScoreDistribution) Add(gr GameResult, weight float32) GameResult {
	other := gr.(ScoreDistribution)
	// other.Start must be > sd.Start since otherwise we will be adding
	// to a probability that is already 1.
	if other.Start < sd.Start {
		panic(fmt.Errorf("adding will create probability > 1: %v <= %v", other.Start, sd.Start))
	}

	maxStop := sd.Stop()
	if other.Stop() > maxStop {
		maxStop = other.Stop()
	}

	newProbabilities := make([]float32, maxStop-sd.Start)
	copy(newProbabilities, sd.Probabilities)
	for s := sd.Start; s < other.Stop(); s++ {
		newProbabilities[s-sd.Start] += weight * other.GetProbability(s)
	}

	return ScoreDistribution{
		Start:         sd.Start,
		Probabilities: newProbabilities,
	}
}

func (sd ScoreDistribution) Shift(offset int) GameResult {
	return ScoreDistribution{
		Start:         sd.Start + offset,
		Probabilities: sd.Probabilities,
	}
}

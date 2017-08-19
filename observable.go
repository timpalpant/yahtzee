package yahtzee

import (
	"fmt"
)

type ExpectedValue float64

func NewExpectedValue() ExpectedValue {
	return ExpectedValue(0)
}

func (ev ExpectedValue) Copy() GameResult {
	return ev
}

func (ev ExpectedValue) Add(other GameResult, weight float64) GameResult {
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

type ScoreDistribution struct {
	// The largest integer such that P(start) == 1.
	Start int
	// [P(start + 1), P(start + 2), ..., P(start + N)]
	// where P(start + N) is the largest integer such that P(s1) > 0.
	Probabilities []float64
}

func NewScoreDistribution() ScoreDistribution {
	return ScoreDistribution{}
}

func (sd ScoreDistribution) Copy() GameResult {
	pCopy := make([]float64, len(sd.Probabilities))
	copy(pCopy, sd.Probabilities)

	return ScoreDistribution{
		Start:         sd.Start,
		Probabilities: pCopy,
	}
}

func (sd ScoreDistribution) Stop() int {
	return sd.Start + len(sd.Probabilities)
}

func (sd ScoreDistribution) GetProbability(score int) float64 {
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

	newProbabilities := make([]float64, maxStop-maxStart)
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

func max(x1, x2 float64) float64 {
	result := x1
	if x2 > x1 {
		result = x2
	}
	return result
}

func (sd ScoreDistribution) Add(gr GameResult, weight float64) GameResult {
	other := gr.(ScoreDistribution)
	maxStop := sd.Stop()
	if other.Stop() > maxStop {
		maxStop = other.Stop()
	}

	newProbabilities := make([]float64, maxStop-sd.Start)
	copy(newProbabilities, sd.Probabilities)
	for s := sd.Start; s < other.Start; s++ {
		newProbabilities[s-sd.Start] += weight
	}
	for s := other.Start; s < other.Stop(); s++ {
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

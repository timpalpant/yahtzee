package optimization

import (
	"encoding/gob"
	"fmt"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/optimization/f32"
)

func init() {
	gob.Register(ScoreDistribution{})
}

// ScoreDistribution implements GameResult, and represents
// minimizing the work required to achieve a desired score.
type ScoreDistribution []float32

func NewScoreDistribution() ScoreDistribution {
	result := make(ScoreDistribution, yahtzee.MaxScore+1)
	for i := range result {
		result[i] = 0
	}

	result[0] = 1
	return result
}

func (ew ScoreDistribution) ScoreDependent() bool {
	return false
}

func (sd ScoreDistribution) New() GameResult {
	return NewScoreDistribution()
}

func (sd ScoreDistribution) GameValue(game yahtzee.GameState) GameResult {
	if !game.GameOver() {
		panic(fmt.Errorf("trying to get endgame value of game that is not over: %v", game))
	}

	return NewScoreDistribution()
}

func (sd ScoreDistribution) CopyInto(other GameResult) GameResult {
	otherSD := other.(ScoreDistribution)
	copy(otherSD, sd)
	return otherSD
}

func (sd ScoreDistribution) Zero(other GameResult) GameResult {
	otherSD := other.(ScoreDistribution)
	for i := range otherSD {
		otherSD[i] = 0
	}
	return otherSD
}

func (sd ScoreDistribution) GetProbability(score int) float32 {
	return sd[score]
}

func (sd ScoreDistribution) Max(gr GameResult) GameResult {
	other := gr.(ScoreDistribution)
	f32.Max(sd, other)
	return sd
}

func (sd ScoreDistribution) Add(gr GameResult, weight float32) GameResult {
	other := gr.(ScoreDistribution)
	f32.AddScaled(sd, weight, other)
	return sd
}

func (sd ScoreDistribution) Shift(offset int) GameResult {
	copy(sd[offset+1:], sd)
	for i := 0; i <= offset; i++ {
		sd[i] = 1
	}
	return sd
}

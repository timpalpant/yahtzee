package optimization

import (
	"encoding/gob"
	"sync"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/optimization/f32"
)

var sdPool = sync.Pool{
	New: func() interface{} {
		return make(ScoreDistribution, yahtzee.MaxScore)
	},
}

func init() {
	gob.Register(ScoreDistribution{})
}

// ScoreDistribution implements GameResult, and represents
// minimizing the work required to achieve a desired score.
type ScoreDistribution []float32

func NewScoreDistribution() ScoreDistribution {
	sd := sdPool.Get().(ScoreDistribution)
	for i := range sd {
		sd[i] = 0
	}

	sd[0] = 1
	return sd
}

func (ew ScoreDistribution) ScoreDependent() bool {
	return false
}

func (sd ScoreDistribution) HashCode() string {
	hasher := floatHasherPool.Get().(*floatHasher)
	defer floatHasherPool.Put(hasher)
	return hasher.HashSlice([]float32(sd))
}

func (sd ScoreDistribution) Close() {
	sdPool.Put(sd)
}

func (sd ScoreDistribution) Copy() GameResult {
	newSD := sdPool.Get().(ScoreDistribution)
	copy(newSD, sd)
	return newSD
}

func (sd ScoreDistribution) Zero(game yahtzee.GameState) GameResult {
	return NewScoreDistribution()
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
	newSD := sdPool.Get().(ScoreDistribution)
	for i := 0; i <= offset; i++ {
		newSD[i] = 1
	}
	copy(newSD[offset+1:], sd)
	return newSD
}

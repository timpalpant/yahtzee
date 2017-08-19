package yahtzee

import (
	"fmt"
	"math"

	"github.com/golang/glog"
)

type MaxExpectedScoreStrategy struct {
	expectedScores  *ScoreCache
	nExpectedScores int

	held1Caches []*ScoreCache
	held2Caches []*ScoreCache
}

func NewMaxExpectedScoreStrategy() *MaxExpectedScoreStrategy {
	return &MaxExpectedScoreStrategy{
		expectedScores: NewScoreCache(MaxGame),
		held1Caches:    make2DCache(NumTurns, MaxRoll),
		held2Caches:    make2DCache(NumTurns, MaxRoll),
	}
}

func make2DCache(size1, size2 int) []*ScoreCache {
	result := make([]*ScoreCache, size1)
	for i := range result {
		result[i] = NewScoreCache(size2)
	}
	return result
}

func (s *MaxExpectedScoreStrategy) SaveToFile(filename string) error {
	return s.expectedScores.SaveToFile(filename)
}

func LoadExpectedScoreTable(filename string) (*MaxExpectedScoreStrategy, error) {
	scores, err := LoadScoreCache(filename)
	if err != nil {
		return nil, err
	}

	return &MaxExpectedScoreStrategy{
		expectedScores: scores,
	}, nil
}

func (s *MaxExpectedScoreStrategy) Compute(game GameState) float64 {
	if game.GameOver() {
		return 0.0
	}

	if s.expectedScores.IsSet(uint(game)) {
		return s.expectedScores.Get(uint(game))
	}

	expectedScore := 0.0
	remainingBoxes := game.AvailableBoxes()
	depth := NumTurns - len(remainingBoxes)
	held1Cache := s.held1Caches[depth]
	held1Cache.Reset()
	held2Cache := s.held2Caches[depth]
	held2Cache.Reset()

	for _, roll1 := range NewRoll().SubsequentRolls() {
		maxValue1 := 0.0
		for _, held1 := range roll1.PossibleHolds() {
			eValue2 := expectedScoreForHold(held1Cache, held1, func(roll2 Roll) float64 {
				maxValue2 := 0.0
				for _, held2 := range roll2.PossibleHolds() {
					eValue3 := expectedScoreForHold(held2Cache, held2, func(roll3 Roll) float64 {
						return s.bestScoreForRoll(game, roll3)
					})

					if eValue3 > maxValue2 {
						maxValue2 = eValue3
					}
				}

				return maxValue2
			})

			if eValue2 > maxValue1 {
				maxValue1 = eValue2
			}
		}

		expectedScore += roll1.Probability() * maxValue1
	}

	s.expectedScores.Set(uint(game), expectedScore)
	s.nExpectedScores++
	if s.nExpectedScores%1000 == 0 {
		glog.Infof("Computed %v games", s.nExpectedScores)
	}
	return expectedScore
}

func (s *MaxExpectedScoreStrategy) bestScoreForRoll(game GameState, roll Roll) float64 {
	best := 0.0
	for _, box := range game.AvailableBoxes() {
		newGame, addedValue := game.FillBox(box, roll)
		expectedRemainingScore := s.Compute(newGame)
		expectedPositionValue := float64(addedValue) + expectedRemainingScore

		if expectedPositionValue > best {
			best = expectedPositionValue
		}
	}

	return best
}

func expectedScoreForHold(heldCache *ScoreCache, held Roll, heldValue func(held Roll) float64) float64 {
	if heldCache.IsSet(uint(held)) {
		return heldCache.Get(uint(held))
	}

	eValue := 0.0
	if held.NumDice() == NDice {
		eValue = heldValue(held)
	} else {
		for side := 1; side <= NSides; side++ {
			eValue += expectedScoreForHold(heldCache, held.Add(side), heldValue) / NSides
		}
	}

	heldCache.Set(uint(held), eValue)
	return eValue
}

type ScoreDistribution struct {
	// The smallest integer such that P(start) < 1.
	Start int
	// [P(start), P(start + 1), ..., P(start + N)]
	// where P(start + N) is the largest integer such that P(s1) > 0.
	Probabilities []float64
}

func (sd *ScoreDistribution) Stop() int {
	return sd.Start + len(sd.Probabilities)
}

func (sd *ScoreDistribution) GetProbability(score int) float64 {
	if score < sd.Start {
		return 1
	}

	idx := score - sd.Start
	if idx >= len(sd.Probabilities) {
		return 0
	}

	return sd.Probabilities[idx]
}

func (sd *ScoreDistribution) String() string {
	return fmt.Sprintf("{Start: %v, Stop: %v, Dist: %v ...}",
		sd.Start, sd.Stop(), sd.Probabilities[:3])
}

func maxDistribution(sd, other *ScoreDistribution) *ScoreDistribution {
	newStart := sd.Start
	if other.Start > newStart {
		newStart = other.Start
	}

	newStop := sd.Stop()
	if other.Stop() > newStop {
		newStop = other.Stop()
	}

	newProbabilities := make([]float64, newStop-newStart)
	for s := newStart; s < newStop; s++ {
		newProbabilities[s-newStart] = max(sd.GetProbability(s), other.GetProbability(s))
	}

	return &ScoreDistribution{
		Start:         newStart,
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

func addDistribution(sd, other *ScoreDistribution, weight float64) *ScoreDistribution {
	newStart := sd.Start
	newStop := sd.Stop()
	if other.Stop() > newStop {
		newStop = other.Stop()
	}

	newProbabilities := make([]float64, newStop-newStart)
	for s := newStart; s < newStop; s++ {
		p := sd.GetProbability(s) + weight*other.GetProbability(s)
		newProbabilities[s-newStart] = p
	}

	for i, p := range newProbabilities {
		if math.Abs(1-p) >= 0.000001 {
			newProbabilities = newProbabilities[i:]
			newStart += i
			break
		}
	}

	return &ScoreDistribution{
		Start:         newStart,
		Probabilities: newProbabilities,
	}
}

func shiftDistribution(sd *ScoreDistribution, shift int) *ScoreDistribution {
	return &ScoreDistribution{
		Start:         sd.Start + shift,
		Probabilities: sd.Probabilities,
	}
}

type BeatHighScoreStrategy struct {
	scoreDistributions *DistributionCache
	nExpectedScores    int

	held1Caches []*DistributionCache
	held2Caches []*DistributionCache
}

func NewBeatHighScoreStrategy() *BeatHighScoreStrategy {
	return &BeatHighScoreStrategy{
		scoreDistributions: NewDistributionCache(MaxGame),
		held1Caches:        make2DDistributionCache(NumTurns, MaxRoll),
		held2Caches:        make2DDistributionCache(NumTurns, MaxRoll),
	}
}

func make2DDistributionCache(size1, size2 int) []*DistributionCache {
	result := make([]*DistributionCache, size1)
	for i := range result {
		result[i] = NewDistributionCache(size2)
	}
	return result
}

func LoadDistributionStrategy(filename string) (*BeatHighScoreStrategy, error) {
	scores, err := LoadDistributionCache(filename)
	if err != nil {
		return nil, err
	}

	return &BeatHighScoreStrategy{
		scoreDistributions: scores,
	}, nil
}

func (s *BeatHighScoreStrategy) SaveToFile(filename string) error {
	return s.scoreDistributions.SaveToFile(filename)
}

func (s *BeatHighScoreStrategy) Compute(game GameState) *ScoreDistribution {
	if game.GameOver() {
		return &ScoreDistribution{}
	}

	if s.scoreDistributions.IsSet(uint(game)) {
		return s.scoreDistributions.Get(uint(game))
	}

	//glog.V(1).Infof("Computing score distribution for game %v", game)
	scoreDistribution := &ScoreDistribution{}
	remainingBoxes := game.AvailableBoxes()
	depth := NumTurns - len(remainingBoxes)
	held1Cache := s.held1Caches[depth]
	held1Cache.Reset()
	held2Cache := s.held2Caches[depth]
	held2Cache.Reset()

	for _, roll1 := range NewRoll().SubsequentRolls() {
		maxValue1 := &ScoreDistribution{}
		for _, held1 := range roll1.PossibleHolds() {
			eValue2 := scoreDistributionForHold(held1Cache, held1, func(roll2 Roll) *ScoreDistribution {
				maxValue2 := &ScoreDistribution{}
				for _, held2 := range roll2.PossibleHolds() {
					eValue3 := scoreDistributionForHold(held2Cache, held2, func(roll3 Roll) *ScoreDistribution {
						return s.scoreDistributionForRoll(game, roll3)
					})

					maxValue2 = maxDistribution(eValue3, maxValue2)
				}

				return maxValue2
			})

			maxValue1 = maxDistribution(eValue2, maxValue1)
		}

		scoreDistribution = addDistribution(scoreDistribution, maxValue1, roll1.Probability())
	}

	//glog.V(1).Infof("Score distribution for game %v: %v (took: %v)", game, scoreDistribution, elapsed)
	s.scoreDistributions.Set(uint(game), scoreDistribution)
	s.nExpectedScores++
	if s.nExpectedScores%1000 == 0 {
		glog.Infof("Computed %v games", s.nExpectedScores)
	}
	return scoreDistribution
}

func (s *BeatHighScoreStrategy) scoreDistributionForRoll(game GameState, roll Roll) *ScoreDistribution {
	best := &ScoreDistribution{}
	for _, box := range game.AvailableBoxes() {
		newGame, addedValue := game.FillBox(box, roll)
		expectedFinalDistribution := shiftDistribution(s.Compute(newGame), addedValue)
		best = maxDistribution(best, expectedFinalDistribution)
	}

	return best
}

func scoreDistributionForHold(heldCache *DistributionCache, held Roll, heldValue func(held Roll) *ScoreDistribution) *ScoreDistribution {
	if heldCache.IsSet(uint(held)) {
		return heldCache.Get(uint(held))
	}

	var eValue *ScoreDistribution
	if held.NumDice() == NDice {
		eValue = heldValue(held)
	} else {
		eValue = &ScoreDistribution{}
		for side := 1; side <= NSides; side++ {
			holdDist := scoreDistributionForHold(heldCache, held.Add(side), heldValue)
			addDistribution(eValue, holdDist, 1.0/NSides)
		}
	}

	heldCache.Set(uint(held), eValue)
	return eValue
}

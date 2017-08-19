package yahtzee

import (
	"github.com/golang/glog"
)

type GameResult interface {
	Copy() GameResult
	Add(other GameResult, weight float64) GameResult
	Max(other GameResult) GameResult
	Shift(offset int) GameResult
}

type Strategy struct {
	observable GameResult
	results    *Cache
	nResults   int

	held1Caches []*Cache
	held2Caches []*Cache
}

func NewStrategy(observable GameResult) *Strategy {
	return &Strategy{
		observable:  observable,
		results:     NewCache(MaxGame),
		held1Caches: New2DCache(NumTurns, MaxRoll),
		held2Caches: New2DCache(NumTurns, MaxRoll),
	}
}

func (s *Strategy) SaveToFile(filename string) error {
	return s.results.SaveToFile(filename)
}

func (s *Strategy) Compute(game GameState) GameResult {
	if game.GameOver() {
		return s.observable.Copy()
	}

	if s.results.IsSet(uint(game)) {
		return s.results.Get(uint(game))
	}

	result := s.observable.Copy()
	remainingBoxes := game.AvailableBoxes()
	depth := NumTurns - len(remainingBoxes)
	held1Cache := s.held1Caches[depth]
	held1Cache.Reset()
	held2Cache := s.held2Caches[depth]
	held2Cache.Reset()

	for _, roll1 := range NewRoll().SubsequentRolls() {
		maxValue1 := s.observable.Copy()
		for _, held1 := range roll1.PossibleHolds() {
			eValue2 := s.expectedResultForHold(held1Cache, held1, func(roll2 Roll) GameResult {
				maxValue2 := s.observable.Copy()
				for _, held2 := range roll2.PossibleHolds() {
					eValue3 := s.expectedResultForHold(held2Cache, held2, func(roll3 Roll) GameResult {
						return s.bestScoreForRoll(game, roll3)
					})

					maxValue2 = maxValue2.Max(eValue3)
				}

				return maxValue2
			})

			maxValue1 = maxValue1.Max(eValue2)
		}

		result = result.Add(maxValue1, roll1.Probability())
	}

	s.results.Set(uint(game), result)
	s.nResults++
	if s.nResults%1000 == 0 {
		glog.Infof("Computed %v games, current: %v, result: %v", s.nResults, game, result)
	}
	return result
}

func (s *Strategy) expectedResultForHold(heldCache *Cache, held Roll, heldValue func(held Roll) GameResult) GameResult {
	if heldCache.IsSet(uint(held)) {
		return heldCache.Get(uint(held))
	}

	eValue := s.observable.Copy()
	if held.NumDice() == NDice {
		eValue = heldValue(held)
	} else {
		for side := 1; side <= NSides; side++ {
			holdResult := s.expectedResultForHold(heldCache, held.Add(side), heldValue)
			eValue = eValue.Add(holdResult, 1.0/NSides)
		}
	}

	heldCache.Set(uint(held), eValue)
	return eValue
}

func (s *Strategy) bestScoreForRoll(game GameState, roll Roll) GameResult {
	best := s.observable.Copy()
	for _, box := range game.AvailableBoxes() {
		newGame, addedValue := game.FillBox(box, roll)
		expectedRemainingScore := s.Compute(newGame)
		expectedPositionValue := expectedRemainingScore.Shift(addedValue)
		best = best.Max(expectedPositionValue)
	}

	return best
}

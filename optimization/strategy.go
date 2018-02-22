package optimization

import (
	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/cache"
)

// GameResult is an observable to maximize.
type GameResult interface {
	Copy() GameResult
	Add(other GameResult, weight float64) GameResult
	Max(other GameResult) GameResult
	Shift(offset int) GameResult
}

// Strategy maximizes an observable GameResult through
// retrograde analysis.
type Strategy struct {
	observable GameResult
	results    *cache.Cache

	held1Caches []*cache.Cache
	held2Caches []*cache.Cache
}

func NewStrategy(observable GameResult) *Strategy {
	return &Strategy{
		observable:  observable,
		results:     cache.New(yahtzee.MaxGame),
		held1Caches: cache.New2D(yahtzee.NumTurns, yahtzee.MaxRoll),
		held2Caches: cache.New2D(yahtzee.NumTurns, yahtzee.MaxRoll),
	}
}

func (s *Strategy) LoadCache(filename string) error {
	return s.results.LoadFromFile(filename)
}

func (s *Strategy) SaveToFile(filename string) error {
	return s.results.SaveToFile(filename)
}

func (s *Strategy) Compute(game yahtzee.GameState) GameResult {
	if game.GameOver() {
		return s.observable.Copy()
	}

	if s.results.IsSet(uint(game)) {
		return s.results.Get(uint(game)).(GameResult)
	}

	// We re-use pre-allocated caches to avoid repeated allocations
	// during strategy table computation.
	h1Cache := s.held1Caches[game.Turn()]
	h1Cache.Reset()
	h2Cache := s.held2Caches[game.Turn()]
	h2Cache.Reset()

	opt := TurnOptimizer{
		strategy:   s,
		game:       game,
		held1Cache: h1Cache,
		held2Cache: h2Cache,
	}

	result := opt.GetOptimalTurnOutcome()
	s.results.Set(uint(game), result)
	return result
}

// TurnOptimizer computes optimal choices for a single turn.
// Once the strategy results table is fully populated, TurnOptimizer
// is thread-safe as long as the caches are not shared.
type TurnOptimizer struct {
	strategy   *Strategy
	game       yahtzee.GameState
	held1Cache *cache.Cache
	held2Cache *cache.Cache
}

func NewTurnOptimizer(strategy *Strategy, game yahtzee.GameState) *TurnOptimizer {
	return &TurnOptimizer{
		strategy:   strategy,
		game:       game,
		held1Cache: cache.New(yahtzee.MaxRoll),
		held2Cache: cache.New(yahtzee.MaxRoll),
	}
}

func (t *TurnOptimizer) GetOptimalTurnOutcome() GameResult {
	result := t.strategy.observable.Copy()
	for _, roll1 := range yahtzee.NewRoll().SubsequentRolls() {
		maxValue1 := t.GetBestHold1(roll1)
		result = result.Add(maxValue1, roll1.Probability())
	}

	return result
}

func (t *TurnOptimizer) GetBestHold1(roll1 yahtzee.Roll) GameResult {
	maxValue1 := t.strategy.observable.Copy()
	for _, held1 := range roll1.PossibleHolds() {
		eValue2 := t.expectedResultForHold(t.held1Cache, held1, t.GetBestHold2)
		maxValue1 = maxValue1.Max(eValue2)
	}

	return maxValue1
}

func (t *TurnOptimizer) GetHold1Outcomes(roll1 yahtzee.Roll) map[yahtzee.Roll]GameResult {
	possibleHolds := roll1.PossibleHolds()
	result := make(map[yahtzee.Roll]GameResult, len(possibleHolds))
	for _, held1 := range possibleHolds {
		result[held1] = t.expectedResultForHold(t.held1Cache, held1, t.GetBestHold2)
	}

	return result
}

func (t *TurnOptimizer) GetBestHold2(roll2 yahtzee.Roll) GameResult {
	maxValue2 := t.strategy.observable.Copy()
	for _, held2 := range roll2.PossibleHolds() {
		eValue3 := t.expectedResultForHold(t.held2Cache, held2, t.GetBestFill)
		maxValue2 = maxValue2.Max(eValue3)
	}

	return maxValue2
}

func (t *TurnOptimizer) GetHold2Outcomes(roll2 yahtzee.Roll) map[yahtzee.Roll]GameResult {
	possibleHolds := roll2.PossibleHolds()
	result := make(map[yahtzee.Roll]GameResult, len(possibleHolds))
	for _, held2 := range possibleHolds {
		result[held2] = t.expectedResultForHold(t.held2Cache, held2, t.GetBestFill)
	}

	return result
}

func (t *TurnOptimizer) GetBestFill(roll yahtzee.Roll) GameResult {
	best := t.strategy.observable.Copy()
	for _, box := range t.game.AvailableBoxes() {
		newGame, addedValue := t.game.FillBox(box, roll)
		expectedRemainingScore := t.strategy.Compute(newGame)
		expectedPositionValue := expectedRemainingScore.Shift(addedValue)
		best = best.Max(expectedPositionValue)
	}

	return best
}

func (t *TurnOptimizer) GetFillOutcomes(roll yahtzee.Roll) map[yahtzee.Box]GameResult {
	availableBoxes := t.game.AvailableBoxes()
	result := make(map[yahtzee.Box]GameResult, len(availableBoxes))
	for _, box := range availableBoxes {
		newGame, addedValue := t.game.FillBox(box, roll)
		expectedRemainingScore := t.strategy.Compute(newGame)
		expectedPositionValue := expectedRemainingScore.Shift(addedValue)
		result[box] = expectedPositionValue
	}

	return result
}

func (t *TurnOptimizer) expectedResultForHold(heldCache *cache.Cache, held yahtzee.Roll, heldValue func(held yahtzee.Roll) GameResult) GameResult {
	if heldCache.IsSet(uint(held)) {
		return heldCache.Get(uint(held)).(GameResult)
	}

	eValue := t.strategy.observable.Copy()
	if held.NumDice() == yahtzee.NDice {
		eValue = heldValue(held)
	} else {
		for side := 1; side <= yahtzee.NSides; side++ {
			holdResult := t.expectedResultForHold(heldCache, held.Add(side), heldValue)
			eValue = eValue.Add(holdResult, 1.0/yahtzee.NSides)
		}
	}

	heldCache.Set(uint(held), eValue)
	return eValue
}

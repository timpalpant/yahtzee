package optimization

import (
	"sync"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

// GameResult is an observable to maximize.
type GameResult interface {
	Zero() GameResult
	Copy() GameResult
	Add(other GameResult, weight float64) GameResult
	Max(other GameResult) GameResult
	Shift(offset int) GameResult
}

// Strategy maximizes an observable GameResult through
// retrograde analysis.
type Strategy struct {
	observable GameResult
	results    *Cache
}

func NewStrategy(observable GameResult) *Strategy {
	return &Strategy{
		observable: observable,
		results:    NewCache(yahtzee.MaxGame),
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
		return s.results.Get(uint(game))
	}

	opt := NewTurnOptimizer(s, game)
	result := opt.GetOptimalTurnOutcome()
	s.results.Set(uint(game), result)
	if s.results.Count()%1000 == 0 {
		glog.V(1).Infof("Computed results for %v games", s.results.Count())
	}
	return result
}

// cachePool maintains a reusable set of caches for TurnOptimizer,
// to reduce memory pressure on the GC during calculation.
var cachePool = sync.Pool{
	New: func() interface{} {
		return NewCache(yahtzee.MaxRoll)
	},
}

// TurnOptimizer computes optimal choices for a single turn.
// Once the strategy results table is fully populated, TurnOptimizer
// is thread-safe as long as the caches are not shared.
type TurnOptimizer struct {
	strategy   *Strategy
	game       yahtzee.GameState
	held1Cache *Cache
	held2Cache *Cache
}

func NewTurnOptimizer(strategy *Strategy, game yahtzee.GameState) *TurnOptimizer {
	held1Cache := cachePool.Get().(*Cache)
	held1Cache.Reset()
	held2Cache := cachePool.Get().(*Cache)
	held2Cache.Reset()

	return &TurnOptimizer{
		strategy:   strategy,
		game:       game,
		held1Cache: held1Cache,
		held2Cache: held2Cache,
	}
}

func (t *TurnOptimizer) GetOptimalTurnOutcome() GameResult {
	glog.V(2).Infof("Computing outcome for game %v", t.game)
	result := t.strategy.observable.Zero()
	for _, roll1 := range yahtzee.AllDistinctRolls() {
		maxValue1 := t.GetBestHold1(roll1)
		result = result.Add(maxValue1, roll1.Probability())
	}

	glog.V(2).Infof("Outcome for game %v = %v", t.game, result)
	return result
}

func (t *TurnOptimizer) GetBestHold1(roll1 yahtzee.Roll) GameResult {
	return t.maxOverHolds(roll1, func(held1 yahtzee.Roll) GameResult {
		return t.expectationOverRolls(t.held1Cache, held1, t.GetBestHold2)
	})
}

func (t *TurnOptimizer) GetHold1Outcomes(roll1 yahtzee.Roll) map[yahtzee.Roll]GameResult {
	possibleHolds := roll1.PossibleHolds()
	result := make(map[yahtzee.Roll]GameResult, len(possibleHolds))
	for _, held1 := range possibleHolds {
		result[held1] = t.expectationOverRolls(t.held1Cache, held1, t.GetBestHold2)
	}

	return result
}

func (t *TurnOptimizer) GetBestHold2(roll2 yahtzee.Roll) GameResult {
	return t.maxOverHolds(roll2, func(held2 yahtzee.Roll) GameResult {
		return t.expectationOverRolls(t.held2Cache, held2, t.GetBestFill)
	})
}

func (t *TurnOptimizer) GetHold2Outcomes(roll2 yahtzee.Roll) map[yahtzee.Roll]GameResult {
	possibleHolds := roll2.PossibleHolds()
	result := make(map[yahtzee.Roll]GameResult, len(possibleHolds))
	for _, held2 := range possibleHolds {
		result[held2] = t.expectationOverRolls(t.held2Cache, held2, t.GetBestFill)
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

func (t *TurnOptimizer) expectationOverRolls(cache *Cache, held yahtzee.Roll, rollValue func(roll yahtzee.Roll) GameResult) GameResult {
	if cache.IsSet(uint(held)) {
		return cache.Get(uint(held))
	}

	eValue := t.strategy.observable.Zero()
	if held.NumDice() == yahtzee.NDice {
		eValue = rollValue(held)
	} else {
		for side := 1; side <= yahtzee.NSides; side++ {
			value := t.expectationOverRolls(cache, held.Add(side), rollValue)
			eValue = eValue.Add(value, 1.0/yahtzee.NSides)
		}
	}

	cache.Set(uint(held), eValue)
	return eValue
}

func (t *TurnOptimizer) maxOverHolds(roll yahtzee.Roll, heldValue func(held yahtzee.Roll) GameResult) GameResult {
	result := t.strategy.observable.Copy()
	for _, held := range roll.PossibleHolds() {
		value := heldValue(held)
		result = result.Max(value)
	}

	return result
}

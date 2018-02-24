package optimization

import (
	"compress/gzip"
	"encoding/gob"
	"io"
	"os"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/cache"
)

// GameResult is an observable to maximize.
type GameResult interface {
	Copy() GameResult
	Add(other GameResult, weight float64) GameResult
	Max(other GameResult) GameResult
	Shift(offset int) GameResult
	IsOver(yahtzee.Game) bool
	Value(yahtzee.Game) GameResult
}

// Strategy maximizes an observable GameResult through
// retrograde analysis.
type Strategy struct {
	observable GameResult
	results    map[yahtzee.Game]GameResult

	held1Caches []*cache.Cache
	held2Caches []*cache.Cache
}

func NewStrategy(observable GameResult) *Strategy {
	return &Strategy{
		observable:  observable,
		results:     make(map[yahtzee.Game]GameResult, yahtzee.MaxGame),
		held1Caches: cache.New2D(yahtzee.NumTurns, yahtzee.MaxRoll),
		held2Caches: cache.New2D(yahtzee.NumTurns, yahtzee.MaxRoll),
	}
}

type GameValue struct {
	Game  yahtzee.Game
	Value GameResult
}

func (s *Strategy) LoadCache(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzf.Close()

	dec := gob.NewDecoder(gzf)
	for {
		var result GameValue
		if err := dec.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		s.results[result.Game] = result.Value
	}

	return nil
}

func (s *Strategy) SaveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	enc := gob.NewEncoder(gzw)
	for game, value := range s.results {
		result := GameValue{game, value}
		if err := enc.Encode(result); err != nil {
			return err
		}
	}

	return nil
}

func (s *Strategy) Compute(game yahtzee.Game) GameResult {
	if result, ok := s.results[game]; ok {
		return result
	}

	if s.observable.IsOver(game) {
		return s.observable.Value(game)
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
	s.results[game] = result
	if len(s.results)%10000 == 0 {
		glog.V(1).Infof("%v game states computed", len(s.results))
	}

	return result
}

// TurnOptimizer computes optimal choices for a single turn.
// Once the strategy results table is fully populated, TurnOptimizer
// is thread-safe as long as the caches are not shared.
type TurnOptimizer struct {
	strategy   *Strategy
	game       yahtzee.Game
	held1Cache *cache.Cache
	held2Cache *cache.Cache
}

func NewTurnOptimizer(strategy *Strategy, game yahtzee.Game) *TurnOptimizer {
	return &TurnOptimizer{
		strategy:   strategy,
		game:       game,
		held1Cache: cache.New(yahtzee.MaxRoll),
		held2Cache: cache.New(yahtzee.MaxRoll),
	}
}

func (t *TurnOptimizer) GetOptimalTurnOutcome() GameResult {
	glog.V(2).Infof("Computing outcome for game %v", t.game)
	result := t.strategy.observable.Copy()
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

func (t *TurnOptimizer) expectationOverRolls(cache *cache.Cache, held yahtzee.Roll, rollValue func(roll yahtzee.Roll) GameResult) GameResult {
	if cache.IsSet(uint(held)) {
		return cache.Get(uint(held)).(GameResult)
	}

	eValue := t.strategy.observable.Copy()
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

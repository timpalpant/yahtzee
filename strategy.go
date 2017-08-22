package yahtzee

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/golang/glog"
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

func LoadExpectedScoresTable(filename string) (*Strategy, error) {
	ev := NewExpectedValue()
	s := NewStrategy(ev)

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var result struct {
			Key   uint
			Value ExpectedValue
		}

		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, err
		}

		s.results.Set(result.Key, result.Value)
	}

	return s, scanner.Err()
}

func LoadScoreDistributionsTable(filename string) (*Strategy, error) {
	obs := NewScoreDistribution()
	s := NewStrategy(obs)

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var result struct {
			Key   uint
			Value ScoreDistribution
		}

		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, err
		}

		s.results.Set(result.Key, result.Value)
	}

	return s, scanner.Err()
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
	s.nResults++
	if s.nResults%1000 == 0 {
		glog.Infof("Computed %v games, current: %v, result: %v", s.nResults, game, result)
	}
	return result
}

// TurnOptimizer computes optimal choices for a single turn.
// Once the strategy results table is fully populated, TurnOptimizer
// is thread-safe as long as the caches are not shared.
type TurnOptimizer struct {
	strategy   *Strategy
	game       GameState
	held1Cache *Cache
	held2Cache *Cache
}

func NewTurnOptimizer(strategy *Strategy, game GameState) *TurnOptimizer {
	return &TurnOptimizer{
		strategy:   strategy,
		game:       game,
		held1Cache: NewCache(MaxRoll),
		held2Cache: NewCache(MaxRoll),
	}
}

func (t *TurnOptimizer) GetOptimalTurnOutcome() GameResult {
	result := t.strategy.observable.Copy()
	for _, roll1 := range NewRoll().SubsequentRolls() {
		maxValue1 := t.GetBestHold1(roll1)
		result = result.Add(maxValue1, roll1.Probability())
	}

	return result
}

func (t *TurnOptimizer) GetBestHold1(roll1 Roll) GameResult {
	maxValue1 := t.strategy.observable.Copy()
	for _, held1 := range roll1.PossibleHolds() {
		eValue2 := t.expectedResultForHold(t.held1Cache, held1, func(roll2 Roll) GameResult {
			return t.GetBestHold2(roll2)
		})

		maxValue1 = maxValue1.Max(eValue2)
	}

	return maxValue1
}

func (t *TurnOptimizer) GetBestHold2(roll2 Roll) GameResult {
	maxValue2 := t.strategy.observable.Copy()
	for _, held2 := range roll2.PossibleHolds() {
		eValue3 := t.expectedResultForHold(t.held2Cache, held2, func(roll3 Roll) GameResult {
			return t.GetBestFill(roll3)
		})

		maxValue2 = maxValue2.Max(eValue3)
	}

	return maxValue2
}

func (t *TurnOptimizer) GetBestFill(roll Roll) GameResult {
	best := t.strategy.observable.Copy()
	for _, box := range t.game.AvailableBoxes() {
		newGame, addedValue := t.game.FillBox(box, roll)
		expectedRemainingScore := t.strategy.Compute(newGame)
		expectedPositionValue := expectedRemainingScore.Shift(addedValue)
		best = best.Max(expectedPositionValue)
	}

	return best
}

func (t *TurnOptimizer) expectedResultForHold(heldCache *Cache, held Roll, heldValue func(held Roll) GameResult) GameResult {
	if heldCache.IsSet(uint(held)) {
		return heldCache.Get(uint(held))
	}

	eValue := t.strategy.observable.Copy()
	if held.NumDice() == NDice {
		eValue = heldValue(held)
	} else {
		for side := 1; side <= NSides; side++ {
			holdResult := t.expectedResultForHold(heldCache, held.Add(side), heldValue)
			eValue = eValue.Add(holdResult, 1.0/NSides)
		}
	}

	heldCache.Set(uint(held), eValue)
	return eValue
}

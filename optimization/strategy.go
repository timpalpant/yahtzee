package optimization

import (
	"container/heap"
	"runtime"
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
	observable     GameResult
	results        *Cache
	resultsPending []bool
	queue          *GameHeap
	cond           *sync.Cond
}

func NewStrategy(observable GameResult) *Strategy {
	s := &Strategy{
		observable:     observable,
		results:        NewCache(yahtzee.MaxGame),
		resultsPending: make([]bool, yahtzee.MaxGame),
		queue:          NewGameHeap(),
		cond:           sync.NewCond(&sync.Mutex{}),
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go s.computeWorker(i)
	}

	for game := yahtzee.NewGame(); game <= yahtzee.MaxGame; game++ {
		if game.GameOver() {
			s.results.Set(uint(game), observable.Copy())
		}
	}

	return s
}

// LoadCache loads the results table for this strategy from the
// given filename.
func (s *Strategy) LoadCache(filename string) error {
	return s.results.LoadFromFile(filename)
}

// SaveToFile serializes the results table for this strategy to
// the given filename.
func (s *Strategy) SaveToFile(filename string) error {
	return s.results.SaveToFile(filename)
}

func (s *Strategy) Populate() GameResult {
	game := yahtzee.NewGame()
	opt := NewTurnOptimizer(s, game, true)
	defer opt.Close()
	result := opt.GetOptimalTurnOutcome()
	s.results.Set(uint(game), result)
	return result
}

// Compute calculates the value of the given GameState for
// the observable that is maximized by this Strategy.
func (s *Strategy) Compute(game yahtzee.GameState) GameResult {
	if s.results.IsSet(uint(game)) {
		return s.results.Get(uint(game))
	}

	s.cond.L.Lock()
	s.queue.Remove(game)
	s.resultsPending[game] = true
	s.cond.L.Unlock()

	return s.computeGame(game)
}

func (s *Strategy) enqueueAsync(game yahtzee.GameState) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	// Enqueue computation for this game if it isn't already in progress.
	if !s.resultsPending[game] {
		s.resultsPending[game] = true
		heap.Push(s.queue, game)
		s.cond.Signal()
	}
}

func (s *Strategy) computeGame(game yahtzee.GameState) GameResult {
	var result GameResult
	if game.GameOver() {
		// Optimization to avoid allocation of terminal state games.
		result = s.observable.Copy()
	} else {
		opt := NewTurnOptimizer(s, game, false)
		defer opt.Close()
		result = opt.GetOptimalTurnOutcome()
	}

	s.results.Set(uint(game), result)
	if s.results.Count()%10000 == 0 {
		glog.V(1).Infof("Computed %v game results", s.results.Count())
	}
	s.cond.Broadcast()
	return result
}

func (s *Strategy) computeUntilReady(game yahtzee.GameState) GameResult {
	for !s.results.IsSet(uint(game)) {
		s.cond.L.Lock()
		for s.queue.Len() == 0 {
			s.cond.Wait()
			if s.results.IsSet(uint(game)) {
				s.cond.L.Unlock()
				return s.results.Get(uint(game))
			}
		}

		other := heap.Pop(s.queue).(yahtzee.GameState)
		s.cond.L.Unlock()

		s.computeGame(other)
	}

	return s.results.Get(uint(game))
}

func (s *Strategy) computeWorker(i int) {
	glog.V(1).Infof("Compute worker %v starting", i)
	defer glog.V(1).Infof("Compute worker %v shutting down", i)
	s.computeUntilReady(yahtzee.NewGame())
}

// ComputeAll runs Compute for each GameState in the given slice,
// and returns the results for each. Each element of the returned slice
// corresponds to the same element i in the input.
//
// The advantage of ComputeAll is that the calculatations may be performed
// in parallel, reducing the total time to calculate all results.
func (s *Strategy) ComputeAll(games []yahtzee.GameState) []GameResult {
	for _, game := range games {
		s.enqueueAsync(game)
	}

	result := make([]GameResult, len(games))
	for i, game := range games {
		result[i] = s.computeUntilReady(game)
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
	parallel   bool
}

func NewTurnOptimizer(strategy *Strategy, game yahtzee.GameState, parallel bool) *TurnOptimizer {
	held1Cache := cachePool.Get().(*Cache)
	held1Cache.Reset()
	held2Cache := cachePool.Get().(*Cache)
	held2Cache.Reset()

	return &TurnOptimizer{
		strategy:   strategy,
		game:       game,
		held1Cache: held1Cache,
		held2Cache: held2Cache,
		parallel:   parallel,
	}
}

func (t *TurnOptimizer) Close() {
	cachePool.Put(t.held1Cache)
	cachePool.Put(t.held2Cache)
}

func (t *TurnOptimizer) GetOptimalTurnOutcome() GameResult {
	if t.game.GameOver() {
		return t.strategy.observable.Copy()
	}

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

type boxResult struct {
	newGame    yahtzee.GameState
	addedValue int
}

func (t *TurnOptimizer) GetBestFill(roll yahtzee.Roll) GameResult {
	if t.parallel {
		return t.getBestFillParallel(roll)
	} else {
		best := t.strategy.observable.Copy()
		for _, box := range t.game.AvailableBoxes() {
			newGame, addedValue := t.game.FillBox(box, roll)
			expectedRemainingScore := t.strategy.Compute(newGame)
			expectedPositionValue := expectedRemainingScore.Shift(addedValue)
			best = best.Max(expectedPositionValue)
		}
		return best
	}
}

func (t *TurnOptimizer) getBestFillParallel(roll yahtzee.Roll) GameResult {
	availableBoxes := t.game.AvailableBoxes()
	boxResults := make([]boxResult, 0, len(availableBoxes))
	toCompute := make([]yahtzee.GameState, 0, len(availableBoxes))
	for _, box := range availableBoxes {
		newGame, addedValue := t.game.FillBox(box, roll)
		boxResults = append(boxResults, boxResult{
			newGame:    newGame,
			addedValue: addedValue,
		})
		toCompute = append(toCompute, newGame)
	}

	subGameResults := t.strategy.ComputeAll(toCompute)
	best := t.strategy.observable.Copy()
	for i, boxResult := range boxResults {
		expectedRemainingScore := subGameResults[i]
		expectedPositionValue := expectedRemainingScore.Shift(boxResult.addedValue)
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

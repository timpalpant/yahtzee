package optimization

import (
	"runtime"
	"sort"
	"sync"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

// GameResult is an observable to maximize.
type GameResult interface {
	Close()
	Zero() GameResult
	Copy() GameResult
	Add(other GameResult, weight float64) GameResult
	Max(other GameResult) GameResult
	Shift(offset int) GameResult
}

type Status int

const (
	NotStarted Status = iota
	InProgress
	Done
)

// Strategy maximizes an observable GameResult through
// retrograde analysis.
type Strategy struct {
	observable GameResult
	results    *Cache
	status     []Status
	queue      []yahtzee.GameState
	cond       *sync.Cond
}

func NewStrategy(observable GameResult) *Strategy {
	return &Strategy{
		observable: observable,
		results:    NewCache(yahtzee.MaxGame),
		status:     make([]Status, yahtzee.MaxGame),
		cond:       sync.NewCond(&sync.Mutex{}),
	}
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
	s.initQueue()
	glog.Infof("%v games in queue", len(s.queue))

	wg := sync.WaitGroup{}
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func(i int) {
			s.computeWorker(i)
			wg.Done()
		}(i)
	}

	wg.Wait()
	return s.Compute(yahtzee.NewGame())
}

func (s *Strategy) initQueue() {
	// Figure out the games we need to compute.
	s.queue = make([]yahtzee.GameState, 0, yahtzee.MaxGame)
	for game := yahtzee.NewGame(); game <= yahtzee.MaxGame; game++ {
		if game.IsValid() {
			s.queue = append(s.queue, game)
		}
	}

	// Sort games to compute by number of turns remaining.
	// i.e. Start at the end games and then proceed to earlier ones.
	sort.Slice(s.queue, func(i, j int) bool {
		return s.queue[i].TurnsRemaining() < s.queue[j].TurnsRemaining()
	})
}

// Compute calculates the value of the given GameState for
// the observable that is maximized by this Strategy.
func (s *Strategy) Compute(game yahtzee.GameState) GameResult {
	if game.GameOver() {
		return s.observable
	}

	if result, ok := s.results.GetIfSet(uint(game)); ok {
		return result
	}

	s.cond.L.Lock()
	switch s.status[game] {
	case NotStarted:
		s.status[game] = InProgress
		s.cond.L.Unlock()
		return s.computeGame(game)
	case InProgress:
		for s.status[game] != Done {
			s.cond.Wait()
		}
	}

	s.cond.L.Unlock()
	return s.results.Get(uint(game))
}

func (s *Strategy) computeWorker(i int) {
	glog.V(1).Infof("Compute worker %v starting", i)
	defer glog.V(1).Infof("Compute worker %v shutting down", i)
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	for len(s.queue) > 0 {
		next := s.queue[0]
		s.queue = s.queue[1:]
		if s.status[next] == NotStarted {
			s.status[next] = InProgress
			s.cond.L.Unlock()
			s.computeGame(next)
			s.cond.L.Lock()
		}
	}
}

func (s *Strategy) computeGame(game yahtzee.GameState) GameResult {
	opt := NewTurnOptimizer(s, game)
	defer opt.Close()
	result := opt.GetOptimalTurnOutcome()
	s.results.Set(uint(game), result)
	if s.results.Count()%10000 == 0 {
		glog.V(1).Infof("Computed %v games", s.results.Count())
	}

	// Wake up anybody waiting for this game to be Done.
	s.status[game] = Done
	s.cond.Broadcast()
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

func (t *TurnOptimizer) Close() {
	cachePool.Put(t.held1Cache)
	cachePool.Put(t.held2Cache)
}

func (t *TurnOptimizer) GetOptimalTurnOutcome() GameResult {
	glog.V(3).Infof("Computing outcome for game %v", t.game)
	result := t.strategy.observable.Zero()
	for _, roll1 := range yahtzee.AllDistinctRolls() {
		maxValue1 := t.GetBestHold1(roll1)
		result = result.Add(maxValue1, roll1.Probability())
		maxValue1.Close()
	}

	glog.V(3).Infof("Outcome for game %v = %v", t.game, result)
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
	best := t.strategy.observable.Copy()
	for _, box := range t.game.AvailableBoxes() {
		newGame, addedValue := t.game.FillBox(box, roll)
		expectedRemainingScore := t.strategy.Compute(newGame)
		expectedPositionValue := expectedRemainingScore.Shift(addedValue)
		best = best.Max(expectedPositionValue)
		expectedPositionValue.Close()
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
	if result, ok := cache.GetIfSet(uint(held)); ok {
		return result
	}

	var eValue GameResult
	if held.NumDice() == yahtzee.NDice {
		eValue = rollValue(held)
	} else {
		eValue = t.strategy.observable.Zero()
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

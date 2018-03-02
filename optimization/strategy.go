package optimization

import (
	"math"
	"runtime"
	"sync"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

// GameResult is an observable to maximize.
type GameResult interface {
	// Whether or not this GameResult is score-sensitive.
	ScoreDependent() bool
	// Release any resources allocated to this GameResult.
	// Note: After calling Close(), the GameResult may no longer be used.
	Close()
	// Return a zero value for the given game.
	Zero(game yahtzee.GameState) GameResult
	// Deep copy this GameResult.
	Copy() GameResult
	// Add the given GameResult to this one, with the given weight scale factor.
	// Store the result in this GameResult.
	Add(other GameResult, weight float32) GameResult
	// Take the max between this and another GameResult,
	// storing the result in this one.
	Max(other GameResult) GameResult
	// Shift this GameResult by the given score offset.
	Shift(offset int) GameResult
	// Return a cryptographic hash of this GameResult value.
	HashCode() string
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
		results:    NewCache(),
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

func (s *Strategy) Populate(games []yahtzee.GameState) {
	queues := asQueues(bucketByTurn(games))
	// Retrograde analysis: process end games first and work backward
	// so that later game results can be reused.
	for turn := maxKey(queues); turn >= minKey(queues); turn-- {
		if toProcess, ok := queues[turn]; ok {
			glog.Infof("Processing turn %v games", turn)
			s.processGames(toProcess)
			glog.Infof("Compressing results cache")
			s.results.Compress()
		} else {
			glog.Warningf("No games for turn %v", turn)
		}
	}
}

func maxKey(m map[int]chan yahtzee.GameState) int {
	result := math.MinInt64
	for key := range m {
		if key > result {
			result = key
		}
	}

	return result
}

func minKey(m map[int]chan yahtzee.GameState) int {
	result := math.MaxInt64
	for key := range m {
		if key < result {
			result = key
		}
	}

	return result
}

func (s *Strategy) processGames(queue chan yahtzee.GameState) {
	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				select {
				case game := <-queue:
					s.computeGame(game)
				default:
					wg.Done()
					return
				}
			}
		}()
	}

	wg.Wait()
	close(queue)
}

// Bucket games to compute by number of turns remaining.
// We want to start at the end games and then proceed to earlier ones,
// so that previous results can be used.
func bucketByTurn(toCompute []yahtzee.GameState) map[int][]yahtzee.GameState {
	glog.Infof("Bucketing %v games to compute by turn", len(toCompute))
	byTurn := make(map[int][]yahtzee.GameState, yahtzee.NumTurns)
	for _, game := range toCompute {
		t := game.Turn()
		byTurn[t] = append(byTurn[t], game)
	}

	return byTurn
}

func asQueues(gamesByTurn map[int][]yahtzee.GameState) map[int]chan yahtzee.GameState {
	glog.Infof("Initializing queues")
	result := make(map[int]chan yahtzee.GameState, len(gamesByTurn))
	for turn, games := range gamesByTurn {
		result[turn] = make(chan yahtzee.GameState, len(games))
		for _, game := range games {
			result[turn] <- game
		}
	}

	return result
}

func (s *Strategy) computeGame(game yahtzee.GameState) GameResult {
	opt := NewTurnOptimizer(s, game)
	defer opt.Close()
	result := opt.GetOptimalTurnOutcome()

	s.results.Set(uint(game), result)
	if s.results.Count()%10000 == 0 {
		glog.V(1).Infof("Computed %v games", s.results.Count())
	}

	return result
}

// Compute calculates the value of the given GameState for
// the observable that is maximized by this Strategy.
func (s *Strategy) Compute(game yahtzee.GameState) GameResult {
	if result, ok := s.results.Get(uint(game)); ok {
		return result
	}

	return s.computeGame(game)
}

// turnOptimizerPool maintains a reusable set of TurnOptimizers,
// to reduce memory pressure on the GC during calculation.
var turnOptimizerPool = sync.Pool{
	New: func() interface{} {
		return &TurnOptimizer{
			held1Cache: NewCache(),
			held2Cache: NewCache(),
		}
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
	opt := turnOptimizerPool.Get().(*TurnOptimizer)
	opt.strategy = strategy
	opt.game = game
	opt.held1Cache.Reset()
	opt.held2Cache.Reset()
	return opt
}

func (t *TurnOptimizer) Close() {
	turnOptimizerPool.Put(t)
}

func (t *TurnOptimizer) GetOptimalTurnOutcome() GameResult {
	if t.game.GameOver() {
		return t.strategy.observable.Zero(t.game)
	}

	glog.V(3).Infof("Computing outcome for game %v", t.game)
	result := t.strategy.observable.Zero(t.game)
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

func (t *TurnOptimizer) GetBestFill(roll yahtzee.Roll) GameResult {
	best := t.strategy.observable.Copy()
	for _, box := range t.game.AvailableBoxes() {
		newGame, addedValue := t.game.FillBox(box, roll)
		// If the observable is not score-dependent, clear the score
		// and apply shift to reduce the game state space significantly.
		var expectedPositionValue GameResult
		if t.strategy.observable.ScoreDependent() {
			expectedPositionValue = t.strategy.Compute(newGame)
		} else {
			expectedRemainingScore := t.strategy.Compute(newGame.Unscored())
			expectedPositionValue = expectedRemainingScore.Shift(addedValue)
		}
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
	if result, ok := cache.Get(uint(held)); ok {
		return result
	}

	var eValue GameResult
	if held.NumDice() == yahtzee.NDice {
		eValue = rollValue(held)
	} else {
		eValue = t.strategy.observable.Zero(t.game)
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

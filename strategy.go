package yahtzee

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

func (s *Strategy) LoadCache(filename string) error {
	return s.results.LoadFromFile(filename)
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
		eValue2 := t.expectedResultForHold(t.held1Cache, held1, t.GetBestHold2)
		maxValue1 = maxValue1.Max(eValue2)
	}

	return maxValue1
}

func (t *TurnOptimizer) GetHold1Outcomes(roll1 Roll) map[Roll]GameResult {
	possibleHolds := roll1.PossibleHolds()
	result := make(map[Roll]GameResult, len(possibleHolds))
	for _, held1 := range possibleHolds {
		result[held1] = t.expectedResultForHold(t.held1Cache, held1, t.GetBestHold2)
	}

	return result
}

func (t *TurnOptimizer) GetBestHold2(roll2 Roll) GameResult {
	maxValue2 := t.strategy.observable.Copy()
	for _, held2 := range roll2.PossibleHolds() {
		eValue3 := t.expectedResultForHold(t.held2Cache, held2, t.GetBestFill)
		maxValue2 = maxValue2.Max(eValue3)
	}

	return maxValue2
}

func (t *TurnOptimizer) GetHold2Outcomes(roll2 Roll) map[Roll]GameResult {
	possibleHolds := roll2.PossibleHolds()
	result := make(map[Roll]GameResult, len(possibleHolds))
	for _, held2 := range possibleHolds {
		result[held2] = t.expectedResultForHold(t.held2Cache, held2, t.GetBestFill)
	}

	return result
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

func (t *TurnOptimizer) GetFillOutcomes(roll Roll) map[Box]GameResult {
	availableBoxes := t.game.AvailableBoxes()
	result := make(map[Box]GameResult, len(availableBoxes))
	for _, box := range availableBoxes {
		newGame, addedValue := t.game.FillBox(box, roll)
		expectedRemainingScore := t.strategy.Compute(newGame)
		expectedPositionValue := expectedRemainingScore.Shift(addedValue)
		result[box] = expectedPositionValue
	}

	return result
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

package server

import (
	"github.com/timpalpant/yahtzee"
)

// GameState represents the current state of the game at the beginning
// of the turn.
type GameState struct {
	Filled               []bool
	YahtzeeBonusEligible bool
	UpperHalfScore       int
}

func FromYahtzeeGameState(game yahtzee.GameState) GameState {
	filled := make([]bool, yahtzee.NumTurns)
	for i := range filled {
		filled[i] = true
	}
	for _, box := range game.AvailableBoxes() {
		filled[int(box)] = false
	}

	return GameState{
		Filled:               filled,
		YahtzeeBonusEligible: game.BonusEligible(),
		UpperHalfScore:       game.UpperHalfScore(),
	}
}

func (gs GameState) ToYahtzeeGameState() yahtzee.GameState {
	game := yahtzee.NewGame()
	for box, filled := range gs.Filled {
		if filled {
			game = game.SetBoxFilled(yahtzee.Box(box))
		}
	}

	if gs.YahtzeeBonusEligible {
		game = game.SetBonusEligible()
	}

	game = game.AddUpperHalfScore(gs.UpperHalfScore)
	return game
}

type TurnStep int

const (
	Hold1 TurnStep = iota
	Hold2
	FillBox
)

// TurnState represents the current progress through a turn.
// In all cases, Dice are the current 5 dice after the previous roll.
type TurnState struct {
	Step TurnStep
	Dice []int
}

// OptimalMoveRequest gets the best move to make given the current
// turn state.
type OptimalMoveRequest struct {
	GameState GameState
	TurnState TurnState

	// If ScoreToBeat is provided, then the result will be the move
	// that maximizes the probability of achieving a final score
	// greater than the given score.
	ScoreToBeat int
}

// Optimal move response returns the best move to make.
type OptimalMoveResponse struct {
	// HeldDice are returned if the TurnState of the request is
	// Hold1 or Hold2.
	HeldDice []int
	// BoxFilled is returned if the TurnState of the request is FillBox.
	BoxFilled int
	// NewGame is set if the optimal move is to quit and start a new game.
	NewGame bool
	// Value is the value attributed to this move.
	// For expected value, it is the expected remaining value.
	// For ScoreToBeat, it is the probability of beating the score.
	Value float64
}

// OutcomeDistributionRequest gets the range of possible outcomes
// for all possible choices at the current turn state.
type OutcomeDistributionRequest struct {
	GameState GameState
	TurnState TurnState
}

// HoldChoices are populated if TurnState.Step is Hold1 or Hold2.
// FillChoices are populated if TurnState.Step is FillBox (after third roll).
type OutcomeDistributionResponse struct {
	HoldChoices []HoldChoice
	FillChoices []FillChoice
}

// HoldChoice represents one possible choice of dice to hold,
// and the associated final outcome if that choice is made.
type HoldChoice struct {
	HeldDice               []int
	ExpectedFinalScore     float64
	FinalScoreDistribution []float64
}

// FillChoice represents the outcome of filling a particular box with
// a given roll.
type FillChoice struct {
	BoxFilled              int
	ExpectedFinalScore     float64
	FinalScoreDistribution []float64
}

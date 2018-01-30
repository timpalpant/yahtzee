package server

import (
	"github.com/timpalpant/yahtzee"
)

// GameState represents the current state of the game at the beginning
// of the turn.
type GameState struct {
	Filled               [13]bool
	YahtzeeBonusEligible bool
	UpperHalfScore       int
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
	Dice [5]int
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

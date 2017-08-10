package yahtzee

import (
	"fmt"

	"github.com/timpalpant/yahtzee/scoring"
	"github.com/timpalpant/yahtzee/tricks"
)

const (
	upperHalfBonusThreshold = 63
	upperHalfBonus          = 35
	yahtzeeBonus            = 100
)

// GameState represents the current state of a game of Yahtzee.
// Note that equivalent states are condensed in this representation.
type GameState struct {
	// Boolean representing whether the box has been played.
	// Index 0-5: Upper half (ones, twos, ... sixes)
	// Index 6-12: Lower half (three of a kind ... yahtzee)
	Played [13]bool
	// The number of boxes that have been played.
	// Invariant: Must remain equal to sum(p for p in Played).
	// This is kept for efficiency.
	NPlayed int

	// Ranges from 0-63. Any value >= 63 means you get the
	// upper half bonus, so all such states are equivalent.
	UpperHalfScore int
	// If true, then the player has played the Yahtzee for
	// points and is eligible for the bonus if another Yahtzee.
	// Note that BonusEligible => Played[12].
	BonusEligible bool
}

// Get the positions that have not yet been played in this game.
func (gs *GameState) AvailablePositions() []tricks.Position {
	if gs.NPlayed == len(gs.Played) {
		return nil
	}

	result := make([]tricks.Position, 0, len(gs.Played)-gs.NPlayed)
	for i, alreadyPlayed := range gs.Played {
		if !alreadyPlayed {
			result = append(result, tricks.Position(i))
		}
	}

	return result
}

// Get the change in the score if the given dice are played
// at the given position in this game. Returns -1 if the
// position has already been played.
func (gs *GameState) ValueAt(roll []int, position tricks.Position) int {
	if gs.Played[position] {
		return -1
	}

	score := scoring.PositionScore(roll, position)

	if position.IsUpperHalf() &&
		gs.UpperHalfScore < upperHalfBonusThreshold &&
		gs.UpperHalfScore+score >= upperHalfBonusThreshold {
		score += upperHalfBonus
	}

	if gs.BonusEligible && tricks.IsYahtzee(roll) {
		score += 100

		// Joker rule.
		switch position {
		case tricks.FullHouse:
			return 25
		case tricks.SmallStraight:
			return 30
		case tricks.LargeStraight:
			return 40
		}
	}

	return score
}

// Return a new GameState constructed by playing the given dice on a position.
func (gs *GameState) PlayPosition(roll []int, position tricks.Position) GameState {
	if gs.Played[position] {
		panic(fmt.Errorf("position has already been played: %v", position))
	}

	value := scoring.PositionScore(roll, position)

	result := *gs
	result.Played[position] = true
	result.NPlayed++

	if position.IsUpperHalf() {
		result.UpperHalfScore += value
		// Cap upper half score at bonus threshold since all values > threshold
		// are equivalent in terms of getting the bonus.
		if result.UpperHalfScore > upperHalfBonusThreshold {
			result.UpperHalfScore = upperHalfBonusThreshold
		}
	}

	if position == tricks.Yahtzee && tricks.IsYahtzee(roll) {
		result.BonusEligible = true
	}

	return result
}

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
	gameOver                = 8191
	maxHash                 = 6400000
)

// GameState represents the current state of a game of Yahtzee.
// Note that equivalent states are condensed in this representation.
type GameState struct {
	// Bitset representing whether the box has been played.
	// Index 0-5: Upper half (ones, twos, ... sixes)
	// Index 6-12: Lower half (three of a kind ... yahtzee)
	// Index 13: Bonus eligible
	Played int

	// Ranges from 0-63. Any value >= 63 means you get the
	// upper half bonus, so all such states are equivalent.
	UpperHalfScore int
}

func (gs *GameState) String() string {
	return fmt.Sprintf("{Remaining: %v, UpperHalf: %v, BonusEligible: %v}",
		gs.AvailablePositions(), gs.UpperHalfScore, gs.BonusEligible())
}

func (gs *GameState) Hash() int {
	return 100000*gs.UpperHalfScore + gs.Played
}

func (gs *GameState) GameOver() bool {
	v := gs.Played &^ (1 << 13)
	return v == gameOver
}

func (gs *GameState) BonusEligible() bool {
	return gs.Played&(1<<13) != 0
}

// Get the positions that have not yet been played in this game.
func (gs *GameState) AvailablePositions() []tricks.Position {
	result := make([]tricks.Position, 0, 13)
	for i := uint(0); i < 13; i++ {
		alreadyPlayed := gs.Played & (1 << i)
		if alreadyPlayed == 0 {
			result = append(result, tricks.Position(i))
		}
	}

	return result
}

// Return a new GameState constructed by playing the given dice on a position.
func (gs *GameState) PlayPosition(roll []int, position tricks.Position) (GameState, int) {
	value := scoring.PositionScore(roll, position)

	result := *gs
	result.Played |= (1 << uint(position))
	if position == tricks.Yahtzee && value != 0 {
		result.Played |= (1 << 13)
	}

	if position.IsUpperHalf() && gs.UpperHalfScore < upperHalfBonusThreshold {
		result.UpperHalfScore += value
		// Cap upper half score at bonus threshold since all values > threshold
		// are equivalent in terms of getting the bonus.
		if result.UpperHalfScore >= upperHalfBonusThreshold {
			result.UpperHalfScore = upperHalfBonusThreshold
			value += upperHalfBonus
		}
	}

	if value != 0 && gs.BonusEligible() && tricks.IsYahtzee(roll) {
		// Second Yahtzee bonus.
		value += 100

		// Joker rule.
		switch position {
		case tricks.FullHouse:
			value += 25
		case tricks.SmallStraight:
			value += 30
		case tricks.LargeStraight:
			value += 40
		}
	}

	return result, value
}

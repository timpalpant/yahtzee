package yahtzee

import (
	"fmt"
)

const (
	MaxGame  = 6400000
	NumTurns = int(Yahtzee + 1)

	UpperHalfBonusThreshold = 63
	UpperHalfBonus          = 35
	YahtzeeBonus            = 100
)

const (
	bonusBit  uint = uint(Yahtzee + 1)
	shiftUHS  uint = bonusBit + 1
	boxesMask      = (1 << bonusBit) - 1
)

// Each distinct game is represented by an integer as follows:
//
//   1. The lowest 13 bits represent whether a box has been filled.
//      Bits 0-5 are the Upper half (ones, twos, ... sixes).
//      Bits 6-12 are the Lower half (three of a kind ... yahtzee)
//   2. Bit 13 represents whether you are eligible for the bonus,
//      meaning that you have previously filled the Yahtzee for points.
//      Therefore bit 13 can only be set if bit 12 is also set.
//   3. Bits 14-19 represent the upper half score in
//      the range [0, 63]. Since for all upper half scores >= 63 you
//      get the upper half bonus, they are equivalent and the upper
//      half score is capped at 63.
//
// This means that all games are represented by an integer < 6.4mm (MaxGame).
type GameState uint

func NewGame() GameState {
	return GameState(0)
}

func (game GameState) GameOver() bool {
	return (game & boxesMask) == boxesMask
}

func (game GameState) BoxFilled(b Box) bool {
	return (game & (1 << b)) != 0
}

func (game GameState) BonusEligible() bool {
	return (game & (1 << bonusBit)) != 0
}

func (game GameState) UpperHalfScore() int {
	return int(game >> shiftUHS)
}

func (game GameState) AvailableBoxes() []Box {
	result := make([]Box, 0)
	for box := Ones; box <= Yahtzee; box++ {
		if !game.BoxFilled(box) {
			result = append(result, box)
		}
	}
	return result
}

func (game GameState) FillBox(box Box, roll Roll) (GameState, int) {
	if game.BoxFilled(box) {
		panic(fmt.Errorf("trying to play already filled box %v", box))
	} else if roll.NumDice() != NDice {
		panic(fmt.Errorf("trying to play incomplete roll with %v dice", roll.NumDice()))
	}

	newGame := game
	value := box.Score(roll)

	newGame |= (1 << box)
	if box == Yahtzee && value != 0 {
		newGame |= (1 << bonusBit)
	}

	prevUHS := game.UpperHalfScore()
	if value != 0 && box.IsUpperHalf() && prevUHS < UpperHalfBonusThreshold {
		newGame += GameState(value << shiftUHS)

		// Cap upper half score at bonus threshold since all values > threshold
		// are equivalent in terms of getting the bonus.
		if prevUHS+value >= UpperHalfBonusThreshold {
			newGame -= GameState((prevUHS + value - UpperHalfBonusThreshold) << shiftUHS)
			value += UpperHalfBonus
		}
	}

	if game.BonusEligible() && IsYahtzee(roll) {
		value += YahtzeeBonus

		// Joker rule: Roll can be played in any box for points,
		// if the corresponding upper half box is already filled.
		nativeBox := nativeUpperHalfBox(roll)
		if game.BoxFilled(nativeBox) {
			switch box {
			case FullHouse:
				value += 25
			case SmallStraight:
				value += 30
			case LargeStraight:
				value += 40
			}
		}
	}

	return newGame, value
}

func (game GameState) String() string {
	return fmt.Sprintf("{ID: %d, Available: %v, BonusEligible: %v, UpperHalf: %v}",
		game, game.AvailableBoxes(), game.BonusEligible(), game.UpperHalfScore())
}

func nativeUpperHalfBox(yahtzeeRoll Roll) Box {
	for box, count := range yahtzeeRoll.Counts() {
		if count > 0 {
			return Box(box)
		}
	}

	panic(fmt.Errorf("error trying to get UH box for: %s", yahtzeeRoll))
}

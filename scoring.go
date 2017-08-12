package yahtzee

type Box uint

const (
	Ones Box = iota
	Twos
	Threes
	Fours
	Fives
	Sixes
	ThreeOfAKind
	FourOfAKind
	FullHouse
	SmallStraight
	LargeStraight
	Chance
	Yahtzee
)

func (b Box) IsUpperHalf() bool {
	return b <= Sixes
}

// BoxScore returns the score that would be received for
// playing the given roll in the given box.
//
// Note it does not include any bonuses.
func (b Box) Score(roll Roll) int {
	if b.IsUpperHalf() {
		side := int(b + 1)
		return side * roll.CountOf(side)
	}

	switch b {
	case ThreeOfAKind:
		return roll.SumOfDice()
	case FourOfAKind:
		return roll.SumOfDice()
	case FullHouse:
		if IsFullHouse(roll) {
			return 25
		}
	case SmallStraight:
		if IsSmallStraight(roll) {
			return 30
		}
	case LargeStraight:
		if IsLargeStraight(roll) {
			return 40
		}
	case Chance:
		return roll.SumOfDice()
	case Yahtzee:
		if IsYahtzee(roll) {
			return 50
		}
	}

	return 0
}

func IsThreeOfAKind(roll Roll) bool {
	return roll.HasNOfAKind(3)
}

func IsFourOfAKind(roll Roll) bool {
	return roll.HasNOfAKind(4)
}

func IsFullHouse(roll Roll) bool {
	return roll.IsFullHouse()
}

func IsSmallStraight(roll Roll) bool {
	return roll.HasNInARow(4)
}

func IsLargeStraight(roll Roll) bool {
	return roll.HasNInARow(5)
}

func IsYahtzee(roll Roll) bool {
	return roll.HasNOfAKind(5)
}

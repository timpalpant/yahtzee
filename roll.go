package yahtzee

import (
	"fmt"
)

const (
	NDice   = 5
	NSides  = 6
	MaxRoll = 500000 + 1
)

var (
	rolls         = enumerateAllRolls()
	holds         = enumerateAllHolds()
	probabilities = computeAllProbabilities()
	pow10         = []int{1, 10, 100, 1000, 10000, 100000}
)

// Type Roll encodes a roll of five dice as an integer.
// The first digit represents the number of ones, the second the
// number of twos, and so on.
//
// Example: [1, 1, 2, 3, 6] => 100112.
//
// This means that all rolls of five dice are represented by
// an integer <= 500000, and permutations are considered equivalent.
type Roll int

func NewRoll() Roll {
	return Roll(0)
}

// Return a new Roll constructed by adding the given die to this one.
func (r Roll) Add(die int) Roll {
	return r + Roll(pow10[die-1])
}

func (r Roll) Remove(die int) Roll {
	if r.CountOf(die) <= 0 {
		panic(fmt.Errorf("Trying to remove die %v from %v", die, r))
	}

	return r - Roll(pow10[die-1])
}

// Count returns the total number of dice in this roll.
func (r Roll) NumDice() int {
	result := 0
	for ; r > 0; r /= 10 {
		count := int(r % 10)
		result += count
	}
	return result
}

func (r Roll) Counts() []int {
	counts := make([]int, NSides)
	for side := 1; side <= NSides; side++ {
		counts[side-1] = int(r % 10)
		r /= 10
	}
	return counts
}

// Return the side of one of the dice in this roll.
func (r Roll) One() int {
	for side := 1; side <= NSides; side++ {
		count := int(r % 10)
		if count > 0 {
			return side
		}
		r /= 10
	}

	return -1
}

// CountOf returns the number of a particular side in this roll.
func (r Roll) CountOf(side int) int {
	return (int(r) / pow10[side-1]) % 10
}

// Return all possible subsequent rolls starting from this one.
// If this roll contains two dice, it will return all possible
// combinations of these two with three others rolled.
// The returned rolls will always contain NDice.
func (r Roll) SubsequentRolls() []Roll {
	return rolls[r]
}

// Return all possible distict kept subsets of this roll.
func (r Roll) PossibleHolds() []Roll {
	return holds[r]
}

func (r Roll) Probability() float64 {
	return probabilities[r]
}

// SumOfDice returns the sum of the sides of all dice.
func (r Roll) SumOfDice() int {
	result := 0
	for side := 1; side <= NSides; side++ {
		count := int(r % 10)
		result += side * count
		r /= 10
	}
	return result
}

// HasNOfAKind checks whether there are at least N of any side in this roll.
func (r Roll) HasNOfAKind(n int) bool {
	for ; r > 0; r /= 10 {
		count := int(r % 10)
		if count >= n {
			return true
		}
	}

	return false
}

// HasNInARow checks whether there is a sequence of N sides in a row.
func (r Roll) HasNInARow(n int) bool {
	nInARow := 0
	for ; r > 0; r /= 10 {
		count := int(r % 10)
		if count > 0 {
			nInARow++
			if nInARow >= n {
				return true
			}
		} else {
			nInARow = 0
		}
	}

	return false
}

func (r Roll) IsFullHouse() bool {
	if r == 0 {
		return false
	}

	for ; r > 0; r /= 10 {
		count := int(r % 10)
		if count != 0 && count != 2 && count != 3 {
			return false
		}
	}

	return true
}

// Dice unpacks this Roll into its natural representation
// as a list of sides.
func (r Roll) Dice() []int {
	dice := make([]int, 0)
	for side := 1; side <= NSides; side++ {
		count := int(r % 10)
		for i := 0; i < count; i++ {
			dice = append(dice, side)
		}
		r /= 10
	}
	return dice
}

func (r Roll) String() string {
	return fmt.Sprintf("%v", r.Dice())
}

func enumerateAllRolls() [][]Roll {
	result := make([][]Roll, MaxRoll)

	for roll := Roll(0); roll < MaxRoll; roll++ {
		if roll.NumDice() > NDice {
			continue // Not a valid Yahtzee roll.
		}

		result[roll] = enumerateRolls(roll, 1)
	}

	return result
}

func enumerateRolls(roll Roll, die int) []Roll {
	numNeeded := NDice - roll.NumDice()
	result := enumerateRollHelper(numNeeded, 1, NSides)
	for i := range result {
		result[i] += roll
	}

	return result
}

func enumerateRollHelper(n, j, k int) []Roll {
	if n == 0 {
		return []Roll{0}
	}

	result := make([]Roll, 0)
	for die := j; die <= k; die++ {
		for _, subRoll := range enumerateRollHelper(n-1, die, k) {
			roll := subRoll.Add(die)
			result = append(result, roll)
		}
	}

	return result
}

func enumerateAllHolds() [][]Roll {
	result := make([][]Roll, MaxRoll)

	for roll := Roll(0); roll < MaxRoll; roll++ {
		if roll.NumDice() > NDice {
			continue // Not a valid Yahtzee roll.
		}

		result[roll] = enumerateHolds(roll, 1)
	}

	return result
}

func enumerateHolds(roll Roll, die int) []Roll {
	if die > NSides {
		return []Roll{0}
	}

	result := make([]Roll, 0)
	// Enumerate in order of most least -> most held so that
	// we can compute expected values over the held multiset efficiently.
	// See Pawlewicz, Appendix B.
	for i := 0; i <= roll.CountOf(die); i++ {
		kept := Roll(i * pow10[die-1])
		for _, remaining := range enumerateHolds(roll, die+1) {
			finalRoll := kept + remaining
			result = append(result, finalRoll)
		}
	}

	return result
}

func pow(n, k int) int {
	result := 1
	for i := 0; i < k; i++ {
		result *= n
	}
	return result
}

func factorial(k int) int {
	result := 1
	for i := 2; i <= k; i++ {
		result *= i
	}
	return result
}

func multinomial(n int, k []int) int {
	result := factorial(n)
	for _, kI := range k {
		result /= factorial(kI)
	}
	return result
}

func computeProbability(roll Roll) float64 {
	n := multinomial(roll.NumDice(), roll.Counts())
	d := pow(NSides, roll.NumDice())
	return float64(n) / float64(d)
}

func computeAllProbabilities() []float64 {
	result := make([]float64, MaxRoll)
	for roll := Roll(0); roll < MaxRoll; roll++ {
		if roll.NumDice() > NDice {
			continue // Not a valid Yahtzee roll.
		}

		result[roll] = computeProbability(roll)
	}

	return result
}

package yahtzee

import (
	"fmt"
)

const (
	NDice  = 5
	NSides = 6

	// The smallest integer such that 2^i >= NDice.
	bitsPerSide = uint(3)
	// Mask to select the count of a single (lowest position) die.
	// e.g. taking roll & dieMask will return the count of ones.
	dieMask = (1 << bitsPerSide) - 1
	// The largest possible roll integer (+1). This can be used
	// to pre-allocate arrays indexed by roll.
	MaxRoll = (NDice << (bitsPerSide * (NSides - 1))) + 1
)

var (
	rolls         = enumerateAllRolls()
	holds         = enumerateAllHolds()
	probabilities = computeAllProbabilities()
)

// Type Roll encodes a roll of five dice as the concatenation of 6 octal integers.
// Each 3 bits represent the number of ones, the number of twos, etc.
type Roll uint

func NewRoll() Roll {
	return Roll(0)
}

// Construct a new roll from the given counts in base-10
// (as opposed to the canonical representation in base-8).
func NewRollFromBase10Counts(counts int) Roll {
	r := NewRoll()

	for side := 1; side <= NSides; side++ {
		count := counts % 10
		r += NewRollOfDie(side, count)
		counts /= 10
	}

	return r
}

// Construct a new Roll with count of the given die.
func NewRollOfDie(die, count int) Roll {
	return Roll(count << (uint(die-1) * bitsPerSide))
}

// Construct a new Roll from the given dice.
func NewRollFromDice(dice []int) Roll {
	r := NewRoll()
	for _, die := range dice {
		r += NewRollOfDie(die, 1)
	}

	return r
}

// Return a new Roll constructed by adding the given die to this one.
func (r Roll) Add(die int) Roll {
	return r + NewRollOfDie(die, 1)
}

func (r Roll) Remove(die int) Roll {
	if r.CountOf(die) <= 0 {
		panic(fmt.Errorf("Trying to remove die %v from %v", die, r))
	}

	return r - NewRollOfDie(die, 1)
}

// NumDice returns the total number of dice in this roll.
func (r Roll) NumDice() int {
	result := 0
	for ; r > 0; r >>= bitsPerSide {
		count := int(r & dieMask)
		result += count
	}
	return result
}

// Return the side of one of the dice in this roll.
func (r Roll) One() int {
	for side := 1; side <= NSides; side++ {
		count := int(r & dieMask)
		if count > 0 {
			return side
		}
		r >>= bitsPerSide
	}

	return -1
}

// CountOf returns the number of a particular side in this roll.
func (r Roll) CountOf(side int) int {
	return int(r>>(uint(side-1)*bitsPerSide)) & dieMask
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
		count := int(r & dieMask)
		result += side * count
		r >>= bitsPerSide
	}
	return result
}

// HasNOfAKind checks whether there are at least N of any side in this roll.
func (r Roll) HasNOfAKind(n int) bool {
	for ; r > 0; r >>= bitsPerSide {
		count := int(r & dieMask)
		if count >= n {
			return true
		}
	}

	return false
}

// HasNInARow checks whether there is a sequence of N sides in a row.
func (r Roll) HasNInARow(n int) bool {
	nInARow := 0
	for ; r > 0; r >>= bitsPerSide {
		count := int(r & dieMask)
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

	for ; r > 0; r >>= bitsPerSide {
		count := int(r & dieMask)
		if count != 0 && count != 2 && count != 3 {
			return false
		}
	}

	return true
}

func (r Roll) Counts() []int {
	counts := make([]int, NSides)
	for side := 1; side <= NSides; side++ {
		counts[side-1] = int(r & dieMask)
		r >>= bitsPerSide
	}
	return counts
}

// Dice unpacks this Roll into its natural representation
// as a list of sides.
func (r Roll) Dice() []int {
	dice := make([]int, 0)
	for side := 1; side <= NSides; side++ {
		count := int(r & dieMask)
		for i := 0; i < count; i++ {
			dice = append(dice, side)
		}
		r >>= bitsPerSide
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
		kept := NewRollOfDie(die, i)
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

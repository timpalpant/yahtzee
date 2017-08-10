package dice

const nSides = 6

// Hash roll into an integer by relying on the fact
// that each die is in the range 1-nSides, with nSides < 10.
func Hash(roll []int) int {
	result := 0
	digit := 1
	for _, die := range roll {
		result += digit * die
		digit *= 10
	}
	return result
}

var rollsCache = map[int][][]int{}

// Return all possible (distinct) rolls of N 6-sided dice.
// Returns an (n+6-1 choose n) slice.
func AllPossibleRolls(n int) [][]int {
	if result, ok := rollsCache[n]; ok {
		return result
	}

	result := enumerateRolls(n, 1, nSides)

	rollsCache[n] = result
	return result
}

func enumerateRolls(n, j, k int) [][]int {
	if n == 0 {
		return [][]int{nil}
	}

	result := make([][]int, 0)
	for i := j; i <= k; i++ {
		for _, subroll := range enumerateRolls(n-1, i, k) {
			roll := append(subroll, i)
			result = append(result, roll)
		}
	}

	return result
}

var pCache = map[int]float64{}

// Return the probability of the given roll amongst
// all possible rolls of len(roll) dice.
func Probability(roll []int) float64 {
	h := Hash(roll)
	if p, ok := pCache[h]; ok {
		return p
	}

	counts := make([]int, nSides)
	for _, die := range roll {
		counts[die-1]++
	}

	n := multinomial(len(roll), counts)
	d := pow(nSides, len(roll))
	p := float64(n) / float64(d)

	pCache[h] = p
	return p
}

func multinomial(n int, k []int) int {
	result := factorial(n)
	for _, ki := range k {
		result /= factorial(ki)
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

func pow(n, k int) int {
	result := 1
	for i := 0; i < k; i++ {
		result *= n
	}

	return result
}

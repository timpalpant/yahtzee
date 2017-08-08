package dice

// Return all possible (distinct) rolls of N k-sided dice.
// Returns an (n+k-1 choose n) slice of [n]int slices.
func AllPossibleRolls(n, k int) [][]int {
	return enumerateRolls(n, 1, k)
}

func enumerateRolls(n, j, k int) [][]int {
	if n == 0 {
		return [][]int{nil}
	}

	result := make([][]int, 0)
	for i := j; i <= k; i++ {
		for _, subroll := range enumerateRolls(n-1, i, k) {
			roll := append([]int{i}, subroll...)
			result = append(result, roll)
		}
	}

	return result
}

// Return the probability of the given roll amongst
// all possible rolls of len(roll) k-sided dice.
func Probability(roll []int, k int) float64 {
	counts := make([]int, k)
	for _, k := range roll {
		counts[k]++
	}

	n := multinomial(len(roll), counts)
	d := pow(k, len(roll))
	return float64(n) / float64(d)
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

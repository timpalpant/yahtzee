package yahtzee

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ScoreCache memoizes computed values. It is designed to be efficiently
// reusable by resetting the isSet array (which uses an efficient memset).
// Unset values are not defined.
type ScoreCache struct {
	values []float64
	isSet  []bool
}

func NewScoreCache(size int) *ScoreCache {
	return &ScoreCache{
		values: make([]float64, size),
		isSet:  make([]bool, size),
	}
}

func LoadScoreCache(filename string) (*ScoreCache, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scores := NewScoreCache(MaxGame)
	for scanner.Scan() {
		entry := strings.Split(scanner.Text(), "\t")
		if len(entry) != 2 {
			return nil, fmt.Errorf("Invalid score table line: %v", scanner.Text())
		}

		game, err := strconv.ParseUint(entry[0], 10, 64)
		if err != nil {
			return nil, err
		}

		score, err := strconv.ParseFloat(entry[0], 64)
		if err != nil {
			return nil, err
		}

		scores.Set(uint(game), score)
	}

	return scores, scanner.Err()
}

func (sc *ScoreCache) SaveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()
	for game, isSet := range sc.isSet {
		if isSet {
			score := sc.Get(uint(game))
			fmt.Fprintf(buf, "%v\t%v\n", game, score)
		}
	}

	return nil
}

func (sc *ScoreCache) Reset() {
	for i := range sc.isSet {
		sc.isSet[i] = false
	}
}

func (sc *ScoreCache) Set(key uint, value float64) {
	sc.values[key] = value
	sc.isSet[key] = true
}

func (sc *ScoreCache) Get(key uint) float64 {
	return sc.values[key]
}

func (sc *ScoreCache) IsSet(key uint) bool {
	return sc.isSet[key]
}

// DistributionCache memoizes computed distributions.
type DistributionCache struct {
	values []*ScoreDistribution
	isSet  []bool
}

func NewDistributionCache(size int) *DistributionCache {
	return &DistributionCache{
		values: make([]*ScoreDistribution, size),
		isSet:  make([]bool, size),
	}
}

func (dc *DistributionCache) Reset() {
	for i := range dc.isSet {
		dc.isSet[i] = false
	}
}

func (dc *DistributionCache) Set(key uint, value *ScoreDistribution) {
	dc.values[key] = value
	dc.isSet[key] = true
}

func (dc *DistributionCache) Get(key uint) *ScoreDistribution {
	return dc.values[key]
}

func (dc *DistributionCache) IsSet(key uint) bool {
	return dc.isSet[key]
}

func LoadDistributionCache(filename string) (*DistributionCache, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scores := NewDistributionCache(MaxGame)
	for scanner.Scan() {
		var result struct {
			Game         uint
			Distribution *ScoreDistribution
		}

		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, err
		}

		scores.Set(result.Game, result.Distribution)
	}

	return scores, scanner.Err()
}

func (dc *DistributionCache) SaveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()
	for game, isSet := range dc.isSet {
		if isSet {
			distribution := dc.Get(uint(game))

			result := struct {
				Game         int
				Distribution *ScoreDistribution
			}{game, distribution}

			jsonBuf, err := json.Marshal(result)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(buf, jsonBuf); err != nil {
				return err
			}
		}
	}

	return nil
}

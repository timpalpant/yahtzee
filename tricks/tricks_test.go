package tricks

import (
	"testing"
)

func TestIsThreeOfAKind(t *testing.T) {
	cases := []struct {
		dice     []int
		expected bool
	}{
		{
			dice:     []int{1, 1, 3, 4, 5},
			expected: false,
		},
		{
			dice:     []int{1, 4, 1, 5, 1},
			expected: true,
		},
		{
			dice:     []int{2, 2, 2, 2, 2},
			expected: true,
		},
	}

	for _, tc := range cases {
		result := IsThreeOfAKind(tc.dice)
		if result != tc.expected {
			t.Errorf("Dice: %v, got %v expected %v", tc.dice, result, tc.expected)
		}
	}
}

func TestIsFourOfAKind(t *testing.T) {
	cases := []struct {
		dice     []int
		expected bool
	}{
		{
			dice:     []int{1, 1, 1, 4, 5},
			expected: false,
		},
		{
			dice:     []int{2, 2, 2, 2, 2},
			expected: true,
		},
		{
			dice:     []int{2, 2, 3, 2, 2},
			expected: true,
		},
	}

	for _, tc := range cases {
		result := IsFourOfAKind(tc.dice)
		if result != tc.expected {
			t.Errorf("Dice: %v, got %v expected %v", tc.dice, result, tc.expected)
		}
	}
}

func TestIsFullHouse(t *testing.T) {
	cases := []struct {
		dice     []int
		expected bool
	}{
		{
			dice:     []int{1, 1, 1, 4, 5},
			expected: false,
		},
		{
			dice:     []int{2, 2, 2, 2, 2},
			expected: false,
		},
		{
			dice:     []int{2, 2, 3, 2, 2},
			expected: false,
		},
		{
			dice:     []int{3, 2, 3, 2, 2},
			expected: true,
		},
	}

	for _, tc := range cases {
		result := IsFullHouse(tc.dice)
		if result != tc.expected {
			t.Errorf("Dice: %v, got %v expected %v", tc.dice, result, tc.expected)
		}
	}
}

func TestIsSmallStraight(t *testing.T) {
	cases := []struct {
		dice     []int
		expected bool
	}{
		{
			dice:     []int{1, 1, 1, 4, 5},
			expected: false,
		},
		{
			dice:     []int{2, 1, 3, 4, 1},
			expected: true,
		},
		{
			dice:     []int{3, 2, 4, 5, 6},
			expected: true,
		},
		{
			dice:     []int{1, 3, 6, 5, 4},
			expected: true,
		},
	}

	for _, tc := range cases {
		result := IsSmallStraight(tc.dice)
		if result != tc.expected {
			t.Errorf("Dice: %v, got %v expected %v", tc.dice, result, tc.expected)
		}
	}
}

func TestIsLargeStraight(t *testing.T) {
	cases := []struct {
		dice     []int
		expected bool
	}{
		{
			dice:     []int{1, 1, 1, 4, 5},
			expected: false,
		},
		{
			dice:     []int{2, 1, 3, 4, 1},
			expected: false,
		},
		{
			dice:     []int{2, 3, 1, 4, 5},
			expected: true,
		},
		{
			dice:     []int{3, 2, 4, 6, 5},
			expected: true,
		},
	}

	for _, tc := range cases {
		result := IsLargeStraight(tc.dice)
		if result != tc.expected {
			t.Errorf("Dice: %v, got %v expected %v", tc.dice, result, tc.expected)
		}
	}
}

func TestIsYahtzee(t *testing.T) {
	cases := []struct {
		dice     []int
		expected bool
	}{
		{
			dice:     []int{1, 1, 1, 1, 5},
			expected: false,
		},
		{
			dice:     []int{2, 2, 2, 2, 2},
			expected: true,
		},
		{
			dice:     []int{6, 6, 6, 6, 6},
			expected: true,
		},
	}

	for _, tc := range cases {
		result := IsYahtzee(tc.dice)
		if result != tc.expected {
			t.Errorf("Dice: %v, got %v expected %v", tc.dice, result, tc.expected)
		}
	}
}

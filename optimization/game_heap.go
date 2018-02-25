package optimization

import (
	"container/heap"

	"github.com/timpalpant/yahtzee"
)

// GameHeap implements heap.Interface and holds yahtzee.GameStates.
// The GameStates are ordered by the number of turns remaining;
// games with fewer turns remaining have higher priority.
type GameHeap struct {
	heap        []yahtzee.GameState
	heapIndices map[yahtzee.GameState]int
}

func NewGameHeap() *GameHeap {
	return &GameHeap{
		heap:        make([]yahtzee.GameState, 0, yahtzee.MaxGame),
		heapIndices: make(map[yahtzee.GameState]int),
	}
}

func (gh GameHeap) Len() int { return len(gh.heap) }

func (gh GameHeap) Less(i, j int) bool {
	return gh.heap[i].TurnsRemaining() < gh.heap[j].TurnsRemaining()
}

func (gh GameHeap) Swap(i, j int) {
	gh.heap[i], gh.heap[j] = gh.heap[j], gh.heap[i]
	gh.heapIndices[gh.heap[i]] = i
	gh.heapIndices[gh.heap[j]] = j
}

func (gh *GameHeap) Push(x interface{}) {
	game := x.(yahtzee.GameState)
	gh.heap = append(gh.heap, game)
	gh.heapIndices[game] = len(gh.heap) - 1
}

func (gh *GameHeap) Pop() interface{} {
	old := gh.heap
	n := len(old)
	game := old[n-1]
	gh.heap = old[:n-1]
	delete(gh.heapIndices, game)
	return game
}

func (gh *GameHeap) Remove(game yahtzee.GameState) {
	if idx, ok := gh.heapIndices[game]; ok {
		heap.Remove(gh, idx)
		delete(gh.heapIndices, game)
	}
}

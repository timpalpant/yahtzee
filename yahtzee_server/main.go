package main

import (
	"encoding/json"
	"net/http"

	"github.com/timpalpant/yahtzee"
)

func computeHistogram(values []int) map[int]int {
	result := make(map[int]int)
	for _, value := range values {
		result[value]++
	}
	return result
}

// Receives a GameState and returns a histogram of possible outcomes.
func yahtzeeHistogramHandler(w http.ResponseWriter, r *http.Request) {
	gameState := yahtzee.GameState{}
	dec = json.NewDecoder(r.Body)
	if err := dec.Decode(&gameState); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if err := gameState.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	outcomes := gameState.GetAllPossibleOutcomes()
	hist := computeHistogram(outcomes)
	enc := json.NewEncoder(w)
	if err := enc.Encode(hist); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/rest/v1/histogram", yahtzeeHistogramHandler)
	http.ListenAndServe(":8123", nil)
}

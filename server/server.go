package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

type YahtzeeServer struct {
	highScoreStrat     *yahtzee.Strategy
	expectedScoreStrat *yahtzee.Strategy
}

func NewYahtzeeServer(highScoreStrat, expectedScoreStrat *yahtzee.Strategy) *YahtzeeServer {
	return &YahtzeeServer{highScoreStrat, expectedScoreStrat}
}

func (ys *YahtzeeServer) Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Execute(w, struct{}{})
}

// OutcomeDistribution returns the probability of achieving a certain score
// given the current game state.
func (ys *YahtzeeServer) OutcomeDistribution(w http.ResponseWriter, r *http.Request) {
	req := &OutcomeDistributionRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := ys.getOutcomes(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		glog.Warning(err)
	}
}

func formatHoldChoices(expectedScores map[yahtzee.Roll]yahtzee.GameResult,
	scoreDistributions map[yahtzee.Roll]yahtzee.GameResult) []HoldChoice {
	holdChoices := make([]HoldChoice, 0, len(expectedScores))
	for roll, es := range expectedScores {
		expectedScore := float64(es.(yahtzee.ExpectedValue))
		distribution := asDistribution(scoreDistributions[roll])
		holdChoice := HoldChoice{roll.Dice(), expectedScore, distribution}
		holdChoices = append(holdChoices, holdChoice)
	}

	return holdChoices
}

func formatFillChoices(expectedScores map[yahtzee.Box]yahtzee.GameResult,
	scoreDistributions map[yahtzee.Box]yahtzee.GameResult) []FillChoice {
	fillChoices := make([]FillChoice, 0, len(expectedScores))
	for box, es := range expectedScores {
		expectedScore := float64(es.(yahtzee.ExpectedValue))
		distribution := asDistribution(scoreDistributions[box])
		fillChoice := FillChoice{int(box), expectedScore, distribution}
		fillChoices = append(fillChoices, fillChoice)
	}

	return fillChoices
}

func (ys *YahtzeeServer) getOutcomes(req *OutcomeDistributionRequest) (*OutcomeDistributionResponse, error) {
	game := req.GameState.ToYahtzeeGameState()
	roll := asRoll(req.TurnState.Dice)
	hsOpt := yahtzee.NewTurnOptimizer(ys.highScoreStrat, game)
	esOpt := yahtzee.NewTurnOptimizer(ys.expectedScoreStrat, game)
	glog.Infof("Computing outcomes for game: %v, roll: %v", game, roll)

	resp := &OutcomeDistributionResponse{}
	switch req.TurnState.Step {
	case Hold1:
		expectedScores := esOpt.GetHold1Outcomes(roll)
		scoreDistributions := hsOpt.GetHold1Outcomes(roll)
		resp.HoldChoices = formatHoldChoices(expectedScores, scoreDistributions)
	case Hold2:
		expectedScores := esOpt.GetHold2Outcomes(roll)
		scoreDistributions := hsOpt.GetHold2Outcomes(roll)
		resp.HoldChoices = formatHoldChoices(expectedScores, scoreDistributions)
	case FillBox:
		expectedScores := esOpt.GetFillOutcomes(roll)
		scoreDistributions := hsOpt.GetFillOutcomes(roll)
		resp.FillChoices = formatFillChoices(expectedScores, scoreDistributions)
	default:
		return nil, fmt.Errorf("Invalid turn state: %v", req.TurnState.Step)
	}

	return resp, nil
}

func asRoll(dice [5]int) yahtzee.Roll {
	r := yahtzee.NewRoll()
	for _, die := range dice {
		r = r.Add(die)
	}

	return r
}

func asDistribution(gr yahtzee.GameResult) []float64 {
	sd := gr.(yahtzee.ScoreDistribution)
	result := make([]float64, yahtzee.MaxScore)
	for score := 0; score < yahtzee.MaxScore; score++ {
		result[score] = sd.GetProbability(score)
	}

	return result
}

package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"

	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/optimization"
)

type YahtzeeServer struct {
	highScoreStrat     *optimization.Strategy
	expectedScoreStrat *optimization.Strategy
	expectedWorkStrat  *optimization.Strategy
}

func NewYahtzeeServer(highScoreStrat, expectedScoreStrat, expectedWorkStrat *optimization.Strategy) *YahtzeeServer {
	return &YahtzeeServer{highScoreStrat, expectedScoreStrat, expectedWorkStrat}
}

func (ys *YahtzeeServer) Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		glog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Execute(w, struct{}{})
}

func (ys *YahtzeeServer) GetScore(w http.ResponseWriter, r *http.Request) {
	req := GetScoreRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		glog.Warning(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	box := yahtzee.Box(req.Box)
	roll := yahtzee.NewRollFromDice(req.Dice)
	score := box.Score(roll)

	resp := GetScoreResponse{Score: score}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		glog.Warning(err)
	}
}

func (ys *YahtzeeServer) OptimalMove(w http.ResponseWriter, r *http.Request) {
	req := &OptimalMoveRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		glog.Warning(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := ys.getOptimalMove(req)
	if err != nil {
		glog.Warning(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		glog.Warning(err)
	}
}

// OutcomeDistribution returns the probability of achieving a certain score
// given the current game state.
func (ys *YahtzeeServer) OutcomeDistribution(w http.ResponseWriter, r *http.Request) {
	req := &OutcomeDistributionRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		glog.Warning(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := ys.getOutcomes(req)
	if err != nil {
		glog.Warning(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		glog.Warning(err)
	}
}

func formatHoldChoices(expectedScores map[yahtzee.Roll]optimization.GameResult,
	scoreDistributions map[yahtzee.Roll]optimization.GameResult) []HoldChoice {
	holdChoices := make([]HoldChoice, 0, len(expectedScores))
	for roll, es := range expectedScores {
		expectedScore := float32(es.(optimization.ExpectedValue))
		distribution := asDistribution(scoreDistributions[roll])
		holdChoice := HoldChoice{roll.Dice(), expectedScore, distribution}
		holdChoices = append(holdChoices, holdChoice)
	}

	return holdChoices
}

func formatFillChoices(expectedScores map[yahtzee.Box]optimization.GameResult,
	scoreDistributions map[yahtzee.Box]optimization.GameResult) []FillChoice {
	fillChoices := make([]FillChoice, 0, len(expectedScores))
	for box, es := range expectedScores {
		expectedScore := float32(es.(optimization.ExpectedValue))
		distribution := asDistribution(scoreDistributions[box])
		fillChoice := FillChoice{int(box), expectedScore, distribution}
		fillChoices = append(fillChoices, fillChoice)
	}

	return fillChoices
}

func (ys *YahtzeeServer) getOptimalMove(req *OptimalMoveRequest) (*OptimalMoveResponse, error) {
	game := req.GameState.ToYahtzeeGameState()
	roll := asRoll(req.TurnState.Dice)
	glog.Infof("Computing optimal move for game: %v, roll: %v", game, roll)

	var opt *optimization.TurnOptimizer
	if req.ScoreToBeat > 0 {
		opt = optimization.NewTurnOptimizer(ys.highScoreStrat, game)
	} else {
		opt = optimization.NewTurnOptimizer(ys.expectedScoreStrat, game)
	}

	resp := &OptimalMoveResponse{}
	switch req.TurnState.Step {
	case yahtzee.Begin:
		outcome := opt.GetOptimalTurnOutcome()
		resp.Value = gameResultValue(outcome, req.ScoreToBeat)
	case yahtzee.Hold1:
		outcomes := opt.GetHold1Outcomes(roll)
		hold, score := bestHold(outcomes, req.ScoreToBeat)
		resp.HeldDice = hold.Dice()
		resp.Value = score
	case yahtzee.Hold2:
		outcomes := opt.GetHold2Outcomes(roll)
		hold, score := bestHold(outcomes, req.ScoreToBeat)
		resp.HeldDice = hold.Dice()
		resp.Value = score
	case yahtzee.FillBox:
		outcomes := opt.GetFillOutcomes(roll)
		fill, score := bestBox(outcomes, req.ScoreToBeat)
		resp.BoxFilled = int(fill)
		resp.Value = score
	default:
		return nil, fmt.Errorf("Invalid turn state: %v", req.TurnState.Step)
	}

	return resp, nil
}

func (ys *YahtzeeServer) getOutcomes(req *OutcomeDistributionRequest) (*OutcomeDistributionResponse, error) {
	game := req.GameState.ToYahtzeeGameState()
	roll := asRoll(req.TurnState.Dice)
	hsOpt := optimization.NewTurnOptimizer(ys.highScoreStrat, game)
	esOpt := optimization.NewTurnOptimizer(ys.expectedScoreStrat, game)
	glog.Infof("Computing outcomes for game: %v, roll: %v", game, roll)

	resp := &OutcomeDistributionResponse{}
	switch req.TurnState.Step {
	case yahtzee.Hold1:
		expectedScores := esOpt.GetHold1Outcomes(roll)
		scoreDistributions := hsOpt.GetHold1Outcomes(roll)
		resp.HoldChoices = formatHoldChoices(expectedScores, scoreDistributions)
	case yahtzee.Hold2:
		expectedScores := esOpt.GetHold2Outcomes(roll)
		scoreDistributions := hsOpt.GetHold2Outcomes(roll)
		resp.HoldChoices = formatHoldChoices(expectedScores, scoreDistributions)
	case yahtzee.FillBox:
	default:
		return nil, fmt.Errorf("Invalid turn state: %v", req.TurnState.Step)
	}

	// Always compute the fill outcomes, since a player may choose to fill
	// a box after only the first or second roll.
	expectedScores := esOpt.GetFillOutcomes(roll)
	scoreDistributions := hsOpt.GetFillOutcomes(roll)
	resp.FillChoices = formatFillChoices(expectedScores, scoreDistributions)

	return resp, nil
}

func asRoll(dice []int) yahtzee.Roll {
	r := yahtzee.NewRoll()
	for _, die := range dice {
		r = r.Add(die)
	}

	return r
}

func asDistribution(gr optimization.GameResult) []float32 {
	sd := gr.(optimization.ScoreDistribution)
	result := make([]float32, yahtzee.MaxScore)
	for score := 0; score < yahtzee.MaxScore; score++ {
		result[score] = sd.GetProbability(score)
	}

	return result
}

func gameResultValue(gr optimization.GameResult, scoreToBeat int) float32 {
	switch gr := gr.(type) {
	case optimization.ExpectedValue:
		return float32(gr)
	case *optimization.ExpectedWork:
		return -float32(gr.Value)
	case optimization.ScoreDistribution:
		return gr.GetProbability(scoreToBeat)
	}

	panic("Unknown game result type")
}

func bestHold(outcomes map[yahtzee.Roll]optimization.GameResult, scoreToBeat int) (yahtzee.Roll, float32) {
	var best yahtzee.Roll
	var bestValue float32 = -math.MaxFloat32
	for hold, gr := range outcomes {
		value := gameResultValue(gr, scoreToBeat)
		if value >= bestValue {
			best = hold
			bestValue = value
		}
	}

	return best, bestValue
}

func bestBox(outcomes map[yahtzee.Box]optimization.GameResult, scoreToBeat int) (yahtzee.Box, float32) {
	var best yahtzee.Box
	var bestValue float32 = -math.MaxFloat32
	for box, gr := range outcomes {
		value := gameResultValue(gr, scoreToBeat)
		if value >= bestValue {
			best = box
			bestValue = value
		}
	}

	return best, bestValue
}

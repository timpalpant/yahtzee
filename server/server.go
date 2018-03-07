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
	roll := yahtzee.NewRollFromDice(req.TurnState.Dice)
	glog.Infof("Computing optimal move for game: %v, roll: %v", game, roll)

	var opt *optimization.TurnOptimizer
	if req.ScoreToBeat > 0 {
		opt = optimization.NewTurnOptimizer(ys.highScoreStrat, game.Unscored())
	} else {
		opt = optimization.NewTurnOptimizer(ys.expectedScoreStrat, game.Unscored())
	}

	workOpt := optimization.NewTurnOptimizer(ys.expectedWorkStrat, game)
	e0 := ys.expectedWorkStrat.Compute(yahtzee.NewGame()).(*optimization.ExpectedWork).Value
	var work float32
	remainingScore := req.ScoreToBeat - game.TotalScore()
	glog.Infof("Score to beat: %v, total score: %v, remaining: %v",
		req.ScoreToBeat, game.TotalScore(), remainingScore)
	resp := &OptimalMoveResponse{}
	switch req.TurnState.Step {
	case yahtzee.Begin:
		outcome := opt.GetOptimalTurnOutcome()
		resp.Value = gameResultValue(outcome, remainingScore)

		work = workOpt.GetOptimalTurnOutcome().(*optimization.ExpectedWork).Value
	case yahtzee.Hold1:
		outcomes := opt.GetHold1Outcomes(roll)
		hold, score := bestHold(outcomes, remainingScore)
		resp.HeldDice = hold.Dice()
		resp.Value = score

		work = workOpt.GetBestHold1(roll).(*optimization.ExpectedWork).Value
	case yahtzee.Hold2:
		outcomes := opt.GetHold2Outcomes(roll)
		hold, score := bestHold(outcomes, remainingScore)
		resp.HeldDice = hold.Dice()
		resp.Value = score

		work = workOpt.GetBestHold2(roll).(*optimization.ExpectedWork).Value
	case yahtzee.FillBox:
		outcomes := opt.GetFillOutcomes(roll)
		fill, score := bestBox(outcomes, remainingScore)
		resp.BoxFilled = int(fill)
		resp.Value = score

		work = workOpt.GetBestFill(roll).(*optimization.ExpectedWork).Value
	default:
		return nil, fmt.Errorf("Invalid turn state: %v", req.TurnState.Step)
	}

	glog.Infof("Work required to win: %v, at start: %v", work, e0)
	resp.StartOver = (work > e0)
	return resp, nil
}

func (ys *YahtzeeServer) getOutcomes(req *OutcomeDistributionRequest) (*OutcomeDistributionResponse, error) {
	game := req.GameState.ToYahtzeeGameState().Unscored()
	roll := yahtzee.NewRollFromDice(req.TurnState.Dice)
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

func asDistribution(gr optimization.GameResult) []float32 {
	return []float32(gr.(optimization.ScoreDistribution))
}

func gameResultValue(gr optimization.GameResult, remainingScore int) float32 {
	switch gr := gr.(type) {
	case optimization.ExpectedValue:
		return float32(gr)
	case *optimization.ExpectedWork:
		return -float32(gr.Value)
	case optimization.ScoreDistribution:
		return gr.GetProbability(remainingScore)
	}

	panic("Unknown game result type")
}

func bestHold(outcomes map[yahtzee.Roll]optimization.GameResult, remainingScore int) (yahtzee.Roll, float32) {
	var best yahtzee.Roll
	var bestValue float32 = -math.MaxFloat32
	for hold, gr := range outcomes {
		value := gameResultValue(gr, remainingScore)
		if value >= bestValue {
			best = hold
			bestValue = value
		}
	}

	return best, bestValue
}

func bestBox(outcomes map[yahtzee.Box]optimization.GameResult, remainingScore int) (yahtzee.Box, float32) {
	var best yahtzee.Box
	var bestValue float32 = -math.MaxFloat32
	for box, gr := range outcomes {
		value := gameResultValue(gr, remainingScore)
		if value >= bestValue {
			best = box
			bestValue = value
		}
	}

	return best, bestValue
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

// GameState represents the current state of the game at the beginning
// of the turn.
type GameState struct {
	Filled               [13]bool
	YahtzeeBonusEligible bool
	UpperHalfScore       int
}

func (gs GameState) ToYahtzeeGameState() yahtzee.GameState {
	game := yahtzee.NewGame()
	for box, filled := range gs.Filled {
		if filled {
			game = game.SetBoxFilled(yahtzee.Box(box))
		}
	}

	if gs.YahtzeeBonusEligible {
		game = game.SetBonusEligible()
	}

	game = game.AddUpperHalfScore(gs.UpperHalfScore)
	return game
}

type TurnStep int

const (
	Hold1 TurnStep = iota
	Hold2
	FillBox
)

// TurnState represents the current progress through a turn.
// In all cases, Dice are the current 5 dice after the previous roll.
type TurnState struct {
	Step TurnStep
	Dice [5]int
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

// OutcomeDistributionRequest gets the range of possible outcomes
// for all possible choices at the current turn state.
type OutcomeDistributionRequest struct {
	GameState GameState
	TurnState TurnState
}

// HoldChoice represents one possible choice of dice to hold,
// and the associated final outcome if that choice is made.
type HoldChoice struct {
	HeldDice               []int
	ExpectedFinalScore     float64
	FinalScoreDistribution []float64
}

// FillChoice represents the outcome of filling a particular box with
// a given roll.
type FillChoice struct {
	BoxFilled              int
	ExpectedFinalScore     float64
	FinalScoreDistribution []float64
}

// HoldChoices are populated if TurnState.Step is Hold1 or Hold2.
// FillChoices are populated if TurnState.Step is FillBox (after third roll).
type OutcomeDistributionResponse struct {
	HoldChoices []HoldChoice
	FillChoices []FillChoice
}

type YahtzeeServer struct {
	highScoreStrat     *yahtzee.Strategy
	expectedScoreStrat *yahtzee.Strategy
}

// Probability returns the probability of achieving a certain score
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

func main() {
	expectedScores := flag.String(
		"expected_scores", "expected-scores.jsonlines",
		"File with expected scores to load")
	scoreDistributions := flag.String(
		"score_distributions", "score-distributions.jsonlines",
		"File with score distributions to load")
	flag.Parse()

	glog.Info("Loading expected scores table")
	expectedScoreStrat := yahtzee.NewStrategy(yahtzee.NewExpectedValue())
	err := expectedScoreStrat.LoadCache(*expectedScores)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Info("Loading score distributions table")
	highScoreStrat := yahtzee.NewStrategy(yahtzee.NewScoreDistribution())
	err = highScoreStrat.LoadCache(*scoreDistributions)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Info("Starting server")
	server := &YahtzeeServer{highScoreStrat, expectedScoreStrat}
	http.Handle("/v1/outcome_distribution",
		gziphandler.GzipHandler(http.HandlerFunc(server.OutcomeDistribution)))
	glog.Fatal(http.ListenAndServe(":8080", nil))
}

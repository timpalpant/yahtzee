/*
 *  Game constants.
 */

let NUM_TURNS = 13;
let NUM_DICE = 5;

let TURN_BEGIN = 0;
let TURN_HOLD1 = 1;
let TURN_HOLD2 = 2;
let TURN_FILL_BOX = 3;

let YAHTZEE_BONUS = 100;

let DIE_SIDE_IMAGES = [
  "/static/images/blank.svg",
  "/static/images/ones.svg",
  "/static/images/twos.svg",
  "/static/images/threes.svg",
  "/static/images/fours.svg",
  "/static/images/fives.svg",
  "/static/images/sixes.svg",
];

/*
 *  Core models for tracking game state.
 */

class Die {
  constructor() {
    this.side = null;
    this.held = false;
  }
}

class GameState {
  constructor() {
    this.boxes = Array(NUM_TURNS).fill(null);
    this.dice = [];
    for (var i = 0; i < NUM_DICE; i++) {
      this.dice.push(new Die());
    }

    this.yahtzeeBonus = 0;

    this.turn = 0;
    this.turnState = TURN_BEGIN;
  }

  get upperHalfScore() {
    var total = 0;
    for (var i = 0; i < 6; i++) {
      if (this.boxes[i] !== null) {
        total += this.boxes[i];
      }
    }

    return total;
  }

  get upperHalfBonus() {
    return game.upperHalfScore >= 63 ? 35 : 0;
  }

  get upperHalfTotal() {
    return this.upperHalfScore + this.upperHalfBonus;
  }

  get lowerHalfTotal() {
    var total = 0;
    for (var i = 6; i < this.boxes.length; i++) {
      if (this.boxes[i] !== null) {
        total += this.boxes[i];
      }
    }

    return total;
  }

  get grandTotal() {
    return this.upperHalfTotal + this.lowerHalfTotal + this.yahtzeeBonus;
  }

  get upperHalfFilled() {
    for (var i = 0; i < 6; i++) {
      if (this.boxes[i] === null) {
        return false;
      }
    }

    return true;
  }

  get yahtzeeBonusEligible() {
    var yahtzeeBox = this.boxes[this.boxes.length - 1];
    return (yahtzeeBox !== null && yahtzeeBox > 0);
  }

  get currentRoll() {
    return this.dice.map((die) => die.side);
  }

  get heldDice() {
    return this.dice.filter((die) => die.held).map((die) => die.side).sort();
  }

  isFilled(box) {
    return this.boxes[box] !== null;
  }

  hold(die) {
    console.log("Holding die " + die);
    this.dice[die].held = !this.dice[die].held;
  }

  fill(box) {
    if (isYahtzee(this.currentRoll) && this.yahtzeeBonusEligible()) {
      console.log("Applying Yahtzee joker bonus");
      this.yahtzeeBonus += YAHTZEE_BONUS;
    }

    console.log("Filling box " + box + " with roll " + this.currentRoll);
    $.ajax({
      type: "POST",
      async: false,
      dataType: "json",
      data: JSON.stringify({"box": box, "dice": this.currentRoll}),
      url: "/rest/v1/score",
      success: (resp) => {
        this.nextTurn(box, resp.Score);
      },
      error: (resp) => {
        console.log(resp);
        window.alert("Error getting score: "+resp);
      }
    });
  }

  nextTurn(box, score) {
    console.log("Scored " + score + " points");
    this.boxes[box] = score;
    this.turn++;
    this.turnState = TURN_BEGIN;
    for (let die of this.dice) {
      die.side = 0;
      die.held = false;
    }
  }

  roll() {
    console.log("Rolling dice");
    for (var i = 0; i < this.dice.length; i++) {
      var die = this.dice[i];
      if (!die.held) {
        console.log("Rolling die " + i);
        die.side = getRandomInt(1, 6);
        console.log("Die " + i + " has new side: " + die.side);
      }
    }

    this.turnState++;
  }
}

class OutcomeCalculator {
  constructor(game, chart) {
    this.game = game;
    this.chart = chart;
    this.initChart();
  }

  get gameStateRequest() {
    return {
      "GameState": {
        "Filled": this.game.boxes.map((box) => box !== null),
        "YahtzeeBonusEligible": this.game.yahtzeeBonusEligible,
        "UpperHalfScore": this.game.upperHalfScore,
      },
      "TurnState": {
        "Step": this.game.turnState,
        "Dice": this.game.currentRoll
      }
    };
  }

  initChart() {
    this.chart.data = {
      labels: Array.from({length: 500}, (x, i) => i),
      datasets: [{
        label: 'Probability of Reaching Score',
        steppedLine: true,
      }]
    }
    this.chart.update();
  }

  update() {
    if (this.game.turnState == TURN_BEGIN) {
      // Can't update outcome distribution until roll.
      return;
    }

    $.ajax({
      type: "POST",
      dataType: "json",
      data: JSON.stringify(this.gameStateRequest),
      url: "/rest/v1/outcome_distribution",
      success: (resp) => {
        this.onSuccess(resp);
      },
      error: (resp) => {
        console.log(resp);
        window.alert("Error fetching outcome distribution: "+resp);
      },
    });
  }

  onSuccess(resp) {
    console.log(resp);
    this.allOptions = resp;
    this.renderHoldChoice();
  }

  get currentHoldChoice() {
    // Get the outcome distribution corresponding to the current held dice.
    let heldDice = this.game.heldDice;
    for (let choice of this.allOptions.HoldChoices) {
      if (isArrayEqual(heldDice, choice.HeldDice)) {
        return choice;
      }
    }
  }

  renderHoldChoice() {
    let current = this.currentHoldChoice;
    this.chart.data.datasets[0].data = current.FinalScoreDistribution;
    this.chart.update();
  }

  renderFillChoice(box) {
    this.chart.data.datasets[0].data = Array.from({length: 500}, (x, i) => i);
    this.chart.update();
  }
}

// Global state variables representing current game state.
var ctx = $("#score-distribution");
var chart = new Chart(ctx, {
  type: 'line',
  data: {},
  options: {
    scales: {
      yAxes: [{
        ticks: {
          beginAtZero: true
        }
      }]
    },
    elements: {
      line: {
        tension: 0, // disables bezier curves
      }
    }
  }
});
var game = new GameState();
var outcomes = new OutcomeCalculator(game, chart);

/*
 *  Rendering functions to render the current game state as display.
 */

var $newGameBtn = $("#new-game-btn");
var $rollBtn = $("#roll-btn");
var $dice = $(".die");
var $boxes = $(".box");

function renderDice() {
  $dice.each(function(index) {
    var die = game.dice[index];

    // Set die image to the correct side.
    var dieImg = DIE_SIDE_IMAGES[die.side];
    $(this).find(".die-img").attr("src", dieImg);

    // Show HELD indicator if die is held.
    if (die.held) {
      $(this).find(".held-indicator").removeClass("invisible");
    } else {
      $(this).find(".held-indicator").addClass("invisible");
    }
  });

  var rollEnabled = (game.turn < NUM_TURNS && game.turnState < TURN_FILL_BOX);
  if (rollEnabled) {
    $rollBtn.removeClass("disabled");
  } else {
    $rollBtn.addClass("disabled");
  }
}

function renderScoreTable() {
  // Show score if already filled, else fill button.
  $boxes.each(function(index) {
    var $box = $(this);
    if (game.isFilled(index)) {
      $box.find("button").addClass("invisible");
      $box.text(game.boxes[index]);
    } else {
      $box.find("button").removeClass("invisible");
    }
  });

  // Fill buttons enabled?
  if (game.turnState == TURN_BEGIN) {
    $boxes.find("button").addClass("disabled");
  } else {
    $boxes.find("button").removeClass("disabled");
  }

  // Upper-half totals.
  $("#upper-half-score").text(game.upperHalfScore);
  $("#upper-half-total").text(game.upperHalfTotal);
  var bonus = game.upperHalfBonus;
  if (bonus === 0 && !game.upperHalfFilled) {
    bonus = "";
  }
  $("#upper-half-bonus").text(bonus);

  // Lower-half totals.
  $("#yahtzee-bonus").text(game.yahtzeeBonus);
  $("#lower-half-total").text(game.lowerHalfTotal);
  $("#grand-total-score").text(game.grandTotal);
}

function render() {
  renderDice();
  renderScoreTable();
}

/*
 * Wiring to hook up all the user interaction to the appropriate
 * game state modifications and re-rendering.
 */

// New game button.
$newGameBtn.click(function() { location.reload() });

// Roll button.
$rollBtn.click(function() {
  if (game.turnState == TURN_FILL_BOX) {
    console.warn("Trying to roll dice at an inappropriate time");
    return;
  }

  game.roll();
  outcomes.update();
});

// Clicking on each of the dice to toggle held state.
$dice.find("a").each(function(index) {
  $(this).click(function() {
    if (game.turnState == TURN_BEGIN) {
      console.warn("Trying to hold dice before they have been rolled");
      return;
    }

    game.hold(index);
    outcomes.update();
  });
});

// Clicking on a box to play the current roll in that box.
$boxes.find("button").each(function(index) {
  $(this).click(function() {
    if (game.turnState == TURN_BEGIN) {
      console.warn("Trying to fill box before dice have been rolled");
      return;
    }

    game.fill(index);
    outcomes.update();
  });

  $(this).mouseenter(function() {
    outcomes.renderFillChoice(index);
  }).mouseleave(function() {
    outcomes.renderHoldChoice();
  });
});

/*
 * Helper functions
 */

/**
 * Returns a random integer between min (inclusive) and max (inclusive)
 * Using Math.round() will give you a non-uniform distribution!
 */
function getRandomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

/**
 * Check whether the given roll of dice is a Yahtzee.
 */
function isYahtzee(roll) {
  if (roll === null || roll.length === 0) {
    return false;
  }

  var first = roll[0];
  for (var i = 1; i < roll.length; i++) {
    if (roll[i] != first) {
      return false;
    }
  }

  return true;
}

function isArrayEqual(arr1, arr2) {
  return arr1.length === arr2.length &&
    arr1.every( function(this_i, i) { return this_i == arr2[i] } )
}

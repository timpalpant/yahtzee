var NUM_TURNS = 13;
var NUM_DICE = 5;

var TURN_BEGIN = 0;
var TURN_HOLD1 = 1;
var TURN_HOLD2 = 2;
var TURN_FILL_BOX = 3;

var DIE_SIDE_IMAGES = [
  "/static/images/blank.svg",
  "/static/images/ones.svg",
  "/static/images/twos.svg",
  "/static/images/threes.svg",
  "/static/images/fours.svg",
  "/static/images/fives.svg",
  "/static/images/sixes.svg",
];

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

    this.turn = 0;
    this.turnState = TURN_BEGIN;
  }

  get yahtzeeBonusEligible() {
    var yahtzeeBox = this.boxes[this.boxes.length - 1];
    return (yahtzeeBox !== null && yahtzeeBox > 0);
  }

  hold(die) {
    console.log("Holding die " + die);
    this.dice[die].held = !this.dice[die].held;
  }

  fill(box) {
    console.log("Filling box " + box);
    this.turn++;
    this.turnState = TURN_BEGIN;
    for (let die of this.dice) {
      die.side = 0;
      die.held = false;
    }
  }

  roll() {
    console.log("Rolling dice");
    if (this.turnState == TURN_FILL_BOX) {
      console.error("Trying to roll dice at an inappropriate time");
      return;
    }

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

/**
 * Returns a random integer between min (inclusive) and max (inclusive)
 * Using Math.round() will give you a non-uniform distribution!
 */
function getRandomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

function getOutcomeDistribution() {
  $.ajax({
    type: "POST",
    dataType: "json",
    data: req,
    url: "/rest/v1/outcome_distribution",
    success: updateOutcomeDistribution,
    error: function(resp) {
      window.alert("Error fetching outcome distribution: "+resp);
    },
  });
}

function updateOutcomeDistribution(resp) {

}

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
      $(this).find(".held-indicator").removeClass("d-none");
    } else {
      $(this).find(".held-indicator").addClass("d-none");
    }
  });

  var rollEnabled = game.turnState < TURN_FILL_BOX;
  if (rollEnabled) {
    $rollBtn.removeClass("disabled");
  } else {
    $rollBtn.addClass("disabled");
  }
}

function renderScoreTable() {

}

// Global representing current game state.
var game = new GameState();

// Hook up all the user interaction to the appropriate
// game state modifications and re-rendering.
$newGameBtn.click(function() { location.reload() });

$rollBtn.click(function() {
  game.roll();
  renderDice();
});

$dice.find("a").each(function(index) {
  $(this).click(function() {
    game.hold(index);
    renderDice();
  });
});

$boxes.find("button").each(function(index) {
  $(this).click(function() {
    game.fill(index);
    renderScoreTable();
    renderDice();
  });
});

var ctx = document.getElementById("score-distribution").getContext('2d');
var myChart = new Chart(ctx, {
  type: 'line',
  data: {
      labels: ["Red", "Blue", "Yellow", "Green", "Purple", "Orange"],
      datasets: [{
          label: '# of Votes',
          data: [12, 19, 3, 5, 2, 3],
          backgroundColor: [
              'rgba(255, 99, 132, 0.2)',
              'rgba(54, 162, 235, 0.2)',
              'rgba(255, 206, 86, 0.2)',
              'rgba(75, 192, 192, 0.2)',
              'rgba(153, 102, 255, 0.2)',
              'rgba(255, 159, 64, 0.2)'
          ],
          borderColor: [
              'rgba(255,99,132,1)',
              'rgba(54, 162, 235, 1)',
              'rgba(255, 206, 86, 1)',
              'rgba(75, 192, 192, 1)',
              'rgba(153, 102, 255, 1)',
              'rgba(255, 159, 64, 1)'
          ],
          borderWidth: 1
      }]
  },
  options: {
      scales: {
          yAxes: [{
              ticks: {
                  beginAtZero:true
              }
          }]
      }
  }
});

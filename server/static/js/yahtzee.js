var NUM_TURNS = 13;

class GameState {
  constructor() {
    this.upperHalfScore = 0;
    this.totalScore = 0;
    this.filled = Array(NUM_TURNS).fill(false);
  }
}

class TurnState {
  constructor(step, dice) {
    this.step = step;
    this.dice = dice;
  }
}

class OutcomeDistributionRequest {
  constructor(gameState, turnState) {
    this.gameState = gameState;
    this.turnState = turnState;
  }
}

var game = GameState();
var rollBtn = $("#roll-btn");
rollBtn.click(function(event) {
    $("#die1").attr("src", "images/twos.svg");
});

var ctx = document.getElementById("score-distribution").getContext('2d');
var myChart = new Chart(ctx, {
  type: 'bar',
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

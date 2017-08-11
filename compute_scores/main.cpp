//
//  main.cpp
//  compute_score_table
//
//  Created by Timothy Palpant on 8/10/17.
//  Copyright © 2017 Timothy Palpant. All rights reserved.
//

#include <iostream>
#include <array>
#include <fstream>
#include <utility>
#include <vector>

#include <gflags/gflags.h>
#include <glog/logging.h>

using namespace std;

const int kUpperHalfBonusThreshold = 63;
const int kUpperHalfBonus = 35;
const int kYahtzeeBonus = 100;

int pow(int n, int k) {
    int result = 1;
    for (int i = 0; i < k; i++) {
        result *= n;
    }
    
    return result;
}


// Each distinct roll is represented by an integer.
// The first digit represents the number of ones, the second the
// number of twos, and so on.
//
// Example: [1, 1, 2, 3, 6] => 100112.
//
// This means that all rolls of five dice are represented by
// an integer <= 500000, and permutations are considered equivalent.
typedef uint Roll;

const int kNDice = 5;
const int kNSides = 6;
const int kMaxRoll = 500000 + 1;

bool is_n_of_a_kind(Roll roll, int n) {
    for (; roll > 0; roll /= 10) {
        int n_dice_of_a_kind = roll % 10;
        if (n_dice_of_a_kind >= n) {
            return true;
        }
        
        roll /= 10;
    }
    
    return false;
}

bool is_full_house(Roll roll) {
    for (; roll > 0; roll /= 10) {
        int n_dice_of_a_kind = roll % 10;
        if (n_dice_of_a_kind != 0 && n_dice_of_a_kind != 2 && n_dice_of_a_kind != 3) {
            return false;
        }
    }
    
    return true;
}

bool has_n_in_a_row(Roll roll, int n) {
    int n_in_a_row = 0;
    for (; roll > 0; roll /= 10) {
        int n_dice_of_a_kind = roll % 10;
        if (n_dice_of_a_kind > 0) {
            n_in_a_row++;
            
            if (n_in_a_row >= n) {
                return true;
            }
        } else {
            n_in_a_row = 0;
        }
    }
    
    return false;
}

bool is_yahtzee(Roll roll) {
    return is_n_of_a_kind(roll, 5);
}

int sum_of_dice(Roll roll) {
    int result = 0;
    
    for (int side = 1; side <= kNSides; side++) {
        int n_dice = roll % 10;
        result += side * n_dice;
        roll /= 10;
    }
    
    return result;
}

int count_dice(Roll roll) {
    int result = 0;
    
    for (; roll > 0; roll /= 10) {
        int n_dice = roll % 10;
        result += n_dice;
    }
    
    return result;
}

int count_dice(Roll roll, int die) {
    return (roll / pow(10, die-1)) % 10;
}


// Each distinct game is represented by an integer as follows:
//
//   1. The lowest 13 bits represent whether a box has been filled.
//      Bits 0-5 are the Upper half (ones, twos, ... sixes).
//      Bits 6-12 are the Lower half (three of a kind ... yahtzee)
//   2. Bit 13 represents whether you are eligible for the bonus,
//      meaning that you have previously filled the Yahtzee for points.
//      Therefore bit 13 can only be set if bit 12 is also set.
//   3. High order bits represent 100,000 x the upper half score in
//      the range [0, 63]. Since for all upper half scores >= 63 you
//      get the upper half bonus, they are equivalent and the upper
//      half score is capped at 63.
//
// This means that all games are represented by an integer < 6.4mm.
typedef uint GameState;

const int kMaxGame = 6400000;
const int kNumBoxes = 13;
const int kUpperHalf = 6;
const int kBoxesMask = (1 << kNumBoxes) - 1;
const int kBonusBit = 13;
const int kUpperHalfScoreMask = ~((1 << 14) - 1);
const int kUpperHalfScoreMultiplier = 100000;

enum Box {
    ones = 0,
    twos,
    threes,
    fours,
    fives,
    sixes,
    three_of_a_kind,
    four_of_a_kind,
    full_house,
    small_straight,
    large_straight,
    chance,
    yahtzee
};

bool is_upper_half(Box box) {
    return box < kUpperHalf;
}

bool box_filled(GameState game, Box box) {
    return game & (1 << box);
}

// The game is over if all of the lowest kNumBoxes bits are set,
// meaning that all boxes have been filled.
bool game_over(GameState game) {
    return (game & kBoxesMask) == kBoxesMask;
}

bool bonus_eligible(GameState game) {
    return game & (1 << kBonusBit);
}

int upper_half_score(GameState game) {
    return (game & kUpperHalfScoreMask) / kUpperHalfScoreMultiplier;
}

vector<Box> available_boxes(GameState game) {
    vector<Box> result;
    for (int boxInt = ones; boxInt <= yahtzee; boxInt++) {
        Box box = static_cast<Box>(boxInt);
        if (!box_filled(game, box)) {
            result.push_back(box);
        }
    }
    
    return result;
}

int box_score(Roll roll, Box box) {
    switch (box) {
        case ones:
            return roll % pow(10, 1);
        case twos:
            return 2 * ((roll / 10) % 10);
        case threes:
            return 3 * ((roll / 100) % 100);
        case fours:
            return 4 * ((roll / 1000) % 1000);
        case fives:
            return 5 * ((roll / 10000) % 10000);
        case sixes:
            return 6 * (roll / 100000);
        case three_of_a_kind:
            if (is_n_of_a_kind(roll, 3)) {
                return sum_of_dice(roll);
            }
        case four_of_a_kind:
            if (is_n_of_a_kind(roll, 4)) {
                return sum_of_dice(roll);
            }
        case full_house:
            if (is_full_house(roll)) {
                return 25;
            }
        case small_straight:
            if (has_n_in_a_row(roll, 4)) {
                return 30;
            }
        case large_straight:
            if (has_n_in_a_row(roll, 5)) {
                return 40;
            }
        case chance:
            return sum_of_dice(roll);
        case yahtzee:
            if (is_yahtzee(roll)) {
                return 50;
            }
    }
    
    return 0;
}

pair<GameState, int> fill_box(GameState game, Roll roll, Box box) {
    GameState new_game = game;
    int value = box_score(roll, box);
    
    new_game |= (1 << box);
    if (box == yahtzee && value != 0) {
        new_game |= (1 << kBonusBit);
    }
    
    if (is_upper_half(box) && upper_half_score(game) < kUpperHalfBonusThreshold) {
        new_game += kUpperHalfScoreMultiplier * value;
        
        // Cap upper half score at bonus threshold since all values > threshold
        // are equivalent in terms of getting the bonus.
        if (upper_half_score(new_game) >= kUpperHalfBonusThreshold) {
            int excess = upper_half_score(new_game) - kUpperHalfBonusThreshold;
            new_game -= kUpperHalfScoreMultiplier * excess;
            value += kUpperHalfBonus;
        }
    }
    
    if (value != 0 && bonus_eligible(game) && is_yahtzee(roll)) {
        // Second Yahtzee bonus.
        value += kYahtzeeBonus;
        
        // Joker rule.
        switch (box) {
            case full_house:
                value += 25;
                break;
            case small_straight:
                value += 30;
                break;
            case large_straight:
                value += 40;
                break;
        }
    }
    
    return {new_game, value};
}

vector<Roll> enumerate_roll_helper(int n, int j, int k) {
    if (n == 0) {
        return {0};
    }
    
    vector<Roll> result;
    for (int die = j; die <= k; die++) {
        for (auto subroll : enumerate_roll_helper(n-1, die, k)) {
            subroll += pow(10, die-1);
            result.push_back(subroll);
        }
    }
    
    return result;
}

vector<Roll> enumerate_rolls(Roll roll, int die) {
    int n_needed = kNDice - count_dice(roll);
    vector<Roll> result = enumerate_roll_helper(n_needed, 1, kNSides);
    for (size_t i = 0; i < result.size(); i++) {
        result[i] += roll;
    }
    
    return result;
}

array<vector<Roll>, kMaxRoll> all_rolls() {
    LOG(INFO) << "Computing rolls table";
    array<vector<Roll>, kMaxRoll> result;
    
    int count = 0;
    for (Roll roll = 0; roll < kMaxRoll; roll++) {
        if (count_dice(roll) > kNDice) {
            continue;  // Not a valid Yahtzee roll.
        }
        
        result[roll] = enumerate_rolls(roll, 1);
        count++;
    }
    
    LOG(INFO) << "Enumerated " << count << " rolls";
    return result;
}

vector<Roll> enumerate_holds(Roll roll, int die) {
    if (die > kNSides) {
        return {0};
    }
    
    vector<Roll> result;
    int n_of_die = count_dice(roll, die);
    int die_value = pow(10, die-1);
    for (int i = 0; i <= n_of_die; i++) {
        Roll kept = i * die_value;
        for (auto remaining : enumerate_holds(kept, die+1)) {
            Roll final_roll = kept + remaining;
            result.push_back(final_roll);
        }
    }
    
    return result;
}

array<vector<Roll>, kMaxRoll> all_holds() {
    LOG(INFO) << "Computing holds table";
    array<vector<Roll>, kMaxRoll> result;
    
    int count = 0;
    for (Roll roll = 0; roll < kMaxRoll; roll++) {
        if (count_dice(roll) > kNDice) {
            continue;  // Not a valid Yahtzee roll.
        }
        
        result[roll] = enumerate_holds(roll, 1);
        count++;
    }
    
    LOG(INFO) << "Enumerated " << count << " holds";
    return result;
}

int factorial(int k) {
    int result = 1;
    for (int i = 2; i <= k; i++) {
        result *= i;
    }
    
    return result;
}

int multinomial(int n, Roll roll) {
    int result = factorial(n);
    for (; roll > 0; roll /= 10) {
        int n_of_die = roll % 10;
        result /= factorial(n_of_die);
    }
    
    return result;
}

double compute_probability(Roll roll) {
    int n_dice = count_dice(roll);
    int n = multinomial(n_dice, roll);
    int d = pow(kNSides, n_dice);
    return static_cast<double>(n) / d;
}

array<double, kMaxRoll> all_probabilities() {
    LOG(INFO) << "Computing roll probabilities table";
    array<double, kMaxRoll> result;
    
    int count = 0;
    for (Roll roll = 0; roll < kMaxRoll; roll++) {
        if (count_dice(roll) > kNDice) {
            continue;  // Not a valid Yahtzee roll.
        }
        
        result[roll] = compute_probability(roll);
        count++;
    }
    
    LOG(INFO) << "Enumerated " << count << " roll probabilities";
    return result;
}

static const array<vector<Roll>, kMaxRoll> rolls = all_rolls();
static const array<vector<Roll>, kMaxRoll> holds = all_holds();
static const array<double, kMaxRoll> probability = all_probabilities();

int n_games_computed = 0;

double compute_expected_score(vector<double>& cache, GameState game) {
    if (game_over(game)) {
        return 0.0;
    }
    
    double expected_score = cache[game];
    if (expected_score != -1) {
        return expected_score;
    }
   
    VLOG(1) << "Computing expected score for game " << game;
    auto remaining_boxes = available_boxes(game);
    int count_iter = 0;
    vector<double> roll2_cache(kMaxRoll, -1);
    vector<double> roll3_cache(kMaxRoll, -1);
    
    for (Roll roll1 : rolls[0]) {
        double max_value1 = 0.0;
        for (Roll held1 : holds[roll1]) {
            double e_value2 = 0;
            for (Roll roll2 : rolls[held1]) {
                double max_value2 = roll2_cache[roll2];
                if (max_value2 == -1) {
                    for (Roll held2 : holds[roll2]) {
                        double e_value3 = 0.0;
                        for (Roll roll3 : rolls[held2]) {
                            double max_value3 = roll3_cache[roll3];
                            if (max_value3 == -1) {
                                for (Box box : remaining_boxes) {
                                    pair<GameState, int> result = fill_box(game, roll3, box);
                                    GameState new_game = result.first;
                                    int added_value = result.second;
                                    
                                    double expected_remaining_score = compute_expected_score(cache, new_game);
                                    double expected_position_value = added_value + expected_remaining_score;
                                    
                                    if (expected_position_value > max_value3) {
                                        max_value3 = expected_position_value;
                                    }

                                    count_iter++;
                                }
                            }
                            
                            roll3_cache[roll3] = max_value3;
                            e_value3 += probability[roll3] * max_value3;
                        }
                        
                        if (e_value3 > max_value2) {
                            max_value2 = e_value3;
                        }
                    }
                }
                
                roll2_cache[roll2] = max_value2;
                e_value2 += probability[roll2] * max_value2;
            }
            
            if (e_value2 > max_value1) {
                max_value1 = e_value2;
            }
        }
        
        expected_score += probability[roll1] * max_value1;
    }
    
    VLOG(1) << "Expected score for game " << game << " = " << expected_score
            << " (" << count_iter << " iterations)";
    n_games_computed++;
    if (n_games_computed % 10000 == 0) {
        LOG(INFO) << "Computed " << n_games_computed << " games\n";
    }
    cache[game] = expected_score;
    return expected_score;
}

int main(int argc, char * argv[]) {
    gflags::ParseCommandLineFlags(&argc, &argv, true);
    google::InitGoogleLogging(argv[0]);

    LOG(INFO) << "Computing expected score table" << endl;
    vector<double> cache(kMaxGame, -1);
    GameState game = 0;
    double result = compute_expected_score(cache, game);
    LOG(INFO) << "Expected score: " << result << endl;
    
    
    string output_filename = "scores.txt";
    LOG(INFO) << "Saving cache table to: " << output_filename << endl;
    ofstream output(output_filename);
    if (!output.is_open()) {
        LOG(ERROR) << "Error opening output file!" << endl;
        return 1;
    }
    
    for (int game = 0; game < kMaxGame; game++) {
        double expected_score = cache[game];
        if (expected_score != 0) {
            output << game << "\t" << expected_score << endl;
        }
    }
    
    output.close();
    return 0;
}

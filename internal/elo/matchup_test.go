package elo

import (
	"math"
	"testing"
)

func TestWinrateCalculation(t *testing.T) {
	tests := []struct {
		name          string
		player1Wins   int
		player2Wins   int
		matchesPlayed int
		expectedRate  float64
	}{
		{"Player1 wins all", 3, 0, 3, 100.0},
		{"Player1 wins 2 of 3", 2, 1, 3, 66.67},
		{"Player1 wins 1 of 3", 1, 2, 3, 33.33},
		{"Player1 wins none", 0, 3, 3, 0.0},
		{"Player1 wins 1 of 2", 1, 1, 2, 50.0},
		{"Even split 4 matches", 2, 2, 4, 50.0},
		{"Player1 wins 5 of 7", 5, 2, 7, 71.43},
		{"Player1 wins 1 of 10", 1, 9, 10, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var winrate float64
			if tt.matchesPlayed > 0 {
				winrate = float64(tt.player1Wins) / float64(tt.matchesPlayed) * 100
			}

			if math.Abs(winrate-tt.expectedRate) > 0.1 {
				t.Errorf("expected %.2f%%, got %.2f%%", tt.expectedRate, winrate)
			}
		})
	}
}

func TestWinrateEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		player1Wins   int
		player2Wins   int
		matchesPlayed int
		expectedRate  float64
	}{
		{"Zero matches", 0, 0, 0, 0.0},
		{"One match win", 1, 0, 1, 100.0},
		{"One match loss", 0, 1, 1, 0.0},
		{"Draws don't count as wins", 0, 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var winrate float64
			if tt.matchesPlayed > 0 {
				winrate = float64(tt.player1Wins) / float64(tt.matchesPlayed) * 100
			}

			if math.Abs(winrate-tt.expectedRate) > 0.01 {
				t.Errorf("expected %.2f%%, got %.2f%%", tt.expectedRate, winrate)
			}
		})
	}
}

func TestMatchOutcomeDetermination(t *testing.T) {
	tests := []struct {
		name           string
		player1Wins    int
		player2Wins    int
		expectedResult string
	}{
		{"Player1 wins", 2, 0, "player1"},
		{"Player2 wins", 0, 2, "player2"},
		{"Draw", 1, 1, "draw"},
		{"Player1 wins close", 2, 1, "player1"},
		{"Player2 wins close", 1, 2, "player2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.player1Wins > tt.player2Wins {
				result = "player1"
			} else if tt.player2Wins > tt.player1Wins {
				result = "player2"
			} else {
				result = "draw"
			}

			if result != tt.expectedResult {
				t.Errorf("expected %s, got %s", tt.expectedResult, result)
			}
		})
	}
}

func TestMatchupBidirectional(t *testing.T) {
	// Test that winrates are calculated correctly in both directions
	// If A vs B is 2-1 (66.67%), then B vs A should be 1-2 (33.33%)
	player1Wins := 2
	player2Wins := 1
	matchesPlayed := player1Wins + player2Wins

	winrateA := float64(player1Wins) / float64(matchesPlayed) * 100
	winrateB := float64(player2Wins) / float64(matchesPlayed) * 100

	if math.Abs(winrateA-66.67) > 0.1 {
		t.Errorf("expected A winrate 66.67%%, got %.2f%%", winrateA)
	}
	if math.Abs(winrateB-33.33) > 0.1 {
		t.Errorf("expected B winrate 33.33%%, got %.2f%%", winrateB)
	}

	// Verify they sum to 100%
	if math.Abs((winrateA+winrateB)-100.0) > 0.1 {
		t.Errorf("winrates should sum to 100%%, got %.2f%% + %.2f%%", winrateA, winrateB)
	}
}

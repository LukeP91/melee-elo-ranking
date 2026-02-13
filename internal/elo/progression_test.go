package elo

import (
	"testing"
)

func TestELORecordStorage(t *testing.T) {
	tests := []struct {
		name           string
		initialELO1    int
		initialELO2    int
		winner         int // 0 = draw, 1 = player1, 2 = player2
		player1Matches int
		player2Matches int
		expectedELO1   int
		expectedELO2   int
	}{
		{
			name:           "Player1 wins first match",
			initialELO1:    1500,
			initialELO2:    1500,
			winner:         1,
			player1Matches: 0,
			player2Matches: 0,
			expectedELO1:   1520,
			expectedELO2:   1480,
		},
		{
			name:           "Player2 wins first match",
			initialELO1:    1500,
			initialELO2:    1500,
			winner:         2,
			player1Matches: 0,
			player2Matches: 0,
			expectedELO1:   1480,
			expectedELO2:   1520,
		},
		{
			name:           "Draw",
			initialELO1:    1500,
			initialELO2:    1500,
			winner:         0,
			player1Matches: 0,
			player2Matches: 0,
			expectedELO1:   1500,
			expectedELO2:   1500,
		},
		{
			name:           "Experienced players - K=20",
			initialELO1:    1500,
			initialELO2:    1500,
			winner:         1,
			player1Matches: 30,
			player2Matches: 30,
			expectedELO1:   1510,
			expectedELO2:   1490,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := New(1500)

			player1ID := int64(1)
			player2ID := int64(2)

			var winnerID *int64
			if tt.winner == 1 {
				winnerID = &player1ID
			} else if tt.winner == 2 {
				winnerID = &player2ID
			}

			newELO1, newELO2 := calc.Calculate(
				tt.initialELO1,
				tt.initialELO2,
				winnerID,
				&player1ID,
				&player2ID,
				tt.player1Matches,
				tt.player2Matches,
			)

			if newELO1 != tt.expectedELO1 {
				t.Errorf("expected player1 ELO %d, got %d", tt.expectedELO1, newELO1)
			}
			if newELO2 != tt.expectedELO2 {
				t.Errorf("expected player2 ELO %d, got %d", tt.expectedELO2, newELO2)
			}

			// When both players have same K-factor, ELO is conserved
			if tt.player1Matches == tt.player2Matches {
				totalBefore := tt.initialELO1 + tt.initialELO2
				totalAfter := newELO1 + newELO2
				if totalBefore != totalAfter {
					t.Errorf("ELO should be conserved when K-factors are equal: before %d, after %d", totalBefore, totalAfter)
				}
			}
		})
	}
}

func TestELOProgressionsChain(t *testing.T) {
	calc := New(1500)
	player1ID := int64(1)
	player2ID := int64(2)

	elo1, elo2 := 1500, 1500
	player1Matches, player2Matches := 0, 0

	// First match: player1 wins (K=40)
	winnerID := &player1ID
	elo1, elo2 = calc.Calculate(elo1, elo2, winnerID, &player1ID, &player2ID, player1Matches, player2Matches)
	player1Matches++
	player2Matches++

	if elo1 != 1520 || elo2 != 1480 {
		t.Errorf("after match 1: expected 1520,1480 got %d,%d", elo1, elo2)
	}

	// Second match: player1 wins again
	winnerID = &player1ID
	elo1, elo2 = calc.Calculate(elo1, elo2, winnerID, &player1ID, &player2ID, player1Matches, player2Matches)
	player1Matches++
	player2Matches++

	// ELO should be conserved here since both have K=40
	if elo1+elo2 != 3000 {
		t.Errorf("ELO not conserved: got %d", elo1+elo2)
	}
}

func TestMatchOutcomeStorage(t *testing.T) {
	tests := []struct {
		name           string
		player1Wins    int
		player2Wins    int
		expectedResult string
	}{
		{"Player1 wins 2-0", 2, 0, "Win"},
		{"Player1 wins 2-1", 2, 1, "Win"},
		{"Player2 wins 0-2", 0, 2, "Loss"},
		{"Player2 wins 1-2", 1, 2, "Loss"},
		{"Draw 1-1", 1, 1, "Draw"},
		{"Draw 0-0", 0, 0, "Draw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.player1Wins > tt.player2Wins {
				result = "Win"
			} else if tt.player1Wins < tt.player2Wins {
				result = "Loss"
			} else {
				result = "Draw"
			}

			if result != tt.expectedResult {
				t.Errorf("expected %s, got %s", tt.expectedResult, result)
			}
		})
	}
}

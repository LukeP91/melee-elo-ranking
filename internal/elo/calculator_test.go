package elo

import (
	"math"
	"testing"
)

func ptr(i int64) *int64 {
	return &i
}

func TestExpectedScore(t *testing.T) {
	calc := New(1500)

	tests := []struct {
		name        string
		playerELO   int
		opponentELO int
		expected    float64
	}{
		{"Same ELO", 1500, 1500, 0.5},
		{"400 points higher", 1900, 1500, 0.909},
		{"400 points lower", 1500, 1900, 0.091},
		{"800 points higher", 2300, 1500, 0.99},
		{"800 points lower", 1500, 2300, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.expectedScore(tt.playerELO, tt.opponentELO)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("expected %.3f, got %.3f", tt.expected, result)
			}
		})
	}
}

func TestDynamicKFactor(t *testing.T) {
	calc := New(1500)

	tests := []struct {
		name          string
		matchesPlayed int
		expectedK     int
	}{
		{"0 matches", 0, 40},
		{"1 match", 1, 40},
		{"10 matches", 10, 40},
		{"29 matches", 29, 40},
		{"30 matches", 30, 20},
		{"31 matches", 31, 20},
		{"100 matches", 100, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.GetDynamicKFactor(tt.matchesPlayed)
			if result != tt.expectedK {
				t.Errorf("expected K=%d, got K=%d", tt.expectedK, result)
			}
		})
	}
}

func TestCalculate_Win(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// Player 1 (1500) beats Player 2 (1500) with K=40 (0 matches played)
	// Expected: Player 1 gains 20, Player 2 loses 20
	newELO1, newELO2 := calc.Calculate(1500, 1500, ptr(player1ID), &player1ID, &player2ID, 0, 0)

	if newELO1 != 1520 {
		t.Errorf("expected winner ELO 1520, got %d", newELO1)
	}
	if newELO2 != 1480 {
		t.Errorf("expected loser ELO 1480, got %d", newELO2)
	}
}

func TestCalculate_Loss(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// Player 1 (1500) loses to Player 2 (1500) with K=40 (0 matches played)
	// Expected: Player 1 loses 20, Player 2 gains 20
	newELO1, newELO2 := calc.Calculate(1500, 1500, ptr(player2ID), &player1ID, &player2ID, 0, 0)

	if newELO1 != 1480 {
		t.Errorf("expected loser ELO 1480, got %d", newELO1)
	}
	if newELO2 != 1520 {
		t.Errorf("expected winner ELO 1520, got %d", newELO2)
	}
}

func TestCalculate_Draw(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// Draw between two 1500 players
	newELO1, newELO2 := calc.Calculate(1500, 1500, nil, &player1ID, &player2ID, 0, 0)

	// In a draw, both should stay at 1500 (or very close)
	if newELO1 != 1500 || newELO2 != 1500 {
		t.Errorf("expected both at 1500, got %d and %d", newELO1, newELO2)
	}
}

func TestCalculate_Upset(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// Low rated (1200) beats high rated (1800)
	// Should gain more points due to K-factor
	newELO1, newELO2 := calc.Calculate(1200, 1800, ptr(player1ID), &player1ID, &player2ID, 0, 0)

	// Winner should gain significant points (more than 10)
	if newELO1 < 1220 {
		t.Errorf("expected winner ELO >= 1220, got %d", newELO1)
	}
	// Loser should lose significant points
	if newELO2 > 1780 {
		t.Errorf("expected loser ELO <= 1780, got %d", newELO2)
	}
}

func TestCalculate_DominantPlayer(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// High rated (1800) beats low rated (1200)
	// Should gain fewer points due to expected win
	newELO1, newELO2 := calc.Calculate(1800, 1200, ptr(player1ID), &player1ID, &player2ID, 0, 0)

	// Winner should gain only a few points
	if newELO1 > 1810 {
		t.Errorf("expected winner ELO <= 1810, got %d", newELO1)
	}
	// Loser should lose only a few points
	if newELO2 < 1190 {
		t.Errorf("expected loser ELO >= 1190, got %d", newELO2)
	}
}

func TestCalculate_DynamicKFactorWin(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// Player with 0 matches (K=40) beats player with 0 matches (K=40)
	newELO1, newELO2 := calc.Calculate(1500, 1500, ptr(player1ID), &player1ID, &player2ID, 0, 0)

	// With K=40, winner should gain 20 points
	if newELO1 != 1520 {
		t.Errorf("expected winner ELO 1520, got %d", newELO1)
	}
	if newELO2 != 1480 {
		t.Errorf("expected loser ELO 1480, got %d", newELO2)
	}
}

func TestCalculate_DynamicKFactorExperienced(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// Player with 30+ matches (K=20) beats player with 30+ matches (K=20)
	newELO1, newELO2 := calc.Calculate(1500, 1500, ptr(player1ID), &player1ID, &player2ID, 30, 30)

	// With K=20, winner should gain 10 points
	if newELO1 != 1510 {
		t.Errorf("expected winner ELO 1510, got %d", newELO1)
	}
	if newELO2 != 1490 {
		t.Errorf("expected loser ELO 1490, got %d", newELO2)
	}
}

func TestCalculate_MixedKFactors(t *testing.T) {
	calc := New(1500)

	player1ID := int64(1)
	player2ID := int64(2)

	// New player (K=40) vs experienced player (K=20)
	newELO1, newELO2 := calc.Calculate(1500, 1500, ptr(player1ID), &player1ID, &player2ID, 0, 30)

	// Winner (new player) should gain 20 points (K=40)
	if newELO1 != 1520 {
		t.Errorf("expected winner ELO 1520, got %d", newELO1)
	}
	// Loser (experienced) should lose 10 points (K=20)
	if newELO2 != 1490 {
		t.Errorf("expected loser ELO 1490, got %d", newELO2)
	}
}

func TestCalculate_InitialRating(t *testing.T) {
	calc := New(1500)

	if calc.GetInitialRating() != 1500 {
		t.Errorf("expected initial rating 1500, got %d", calc.GetInitialRating())
	}
}


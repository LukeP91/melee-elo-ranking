package elo

import "math"

type Calculator struct {
	kFactor           int
	initialRating     int
	dynamicKThreshold int
}

func New(kFactor, initialRating int) *Calculator {
	return &Calculator{
		kFactor:           kFactor,
		initialRating:     initialRating,
		dynamicKThreshold: 30,
	}
}

func (c *Calculator) GetDynamicKFactor(matchesPlayed int) int {
	if matchesPlayed < c.dynamicKThreshold {
		return 40
	}
	return 20
}

// Calculate computes new ELO ratings after a match
// winnerID: nil if draw, otherwise ID of winning player
// player1ID, player2ID: IDs of the two players
// player1Matches, player2Matches: number of matches already played by each player
// Returns: (newELO1, newELO2)
func (c *Calculator) Calculate(player1ELO, player2ELO int, winnerID, player1ID, player2ID *int64, player1Matches, player2Matches int) (int, int) {
	// Calculate expected scores
	expected1 := c.expectedScore(player1ELO, player2ELO)
	expected2 := c.expectedScore(player2ELO, player1ELO)

	// Determine actual scores
	var actual1, actual2 float64
	if winnerID == nil {
		// Draw
		actual1 = 0.5
		actual2 = 0.5
	} else if *winnerID == *player1ID {
		// Player 1 won
		actual1 = 1.0
		actual2 = 0.0
	} else {
		// Player 2 won
		actual1 = 0.0
		actual2 = 1.0
	}

	// Use dynamic K-factor based on matches played
	k1 := float64(c.GetDynamicKFactor(player1Matches))
	k2 := float64(c.GetDynamicKFactor(player2Matches))

	newELO1 := int(math.Round(float64(player1ELO) + k1*(actual1-expected1)))
	newELO2 := int(math.Round(float64(player2ELO) + k2*(actual2-expected2)))

	return newELO1, newELO2
}

// expectedScore calculates the expected score for a player
func (c *Calculator) expectedScore(playerELO, opponentELO int) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, float64(opponentELO-playerELO)/400.0))
}

func (c *Calculator) GetInitialRating() int {
	return c.initialRating
}

func (c *Calculator) GetKFactor() int {
	return c.kFactor
}

func (c *Calculator) SetDynamicKThreshold(threshold int) {
	c.dynamicKThreshold = threshold
}

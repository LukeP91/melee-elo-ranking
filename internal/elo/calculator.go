package elo

import "math"

type Calculator struct {
	kFactor       int
	initialRating int
}

func New(kFactor, initialRating int) *Calculator {
	return &Calculator{
		kFactor:       kFactor,
		initialRating: initialRating,
	}
}

// Calculate computes new ELO ratings after a match
// winnerID: nil if draw, otherwise ID of winning player
// player1ID, player2ID: IDs of the two players
// Returns: (newELO1, newELO2)
func (c *Calculator) Calculate(player1ELO, player2ELO int, winnerID, player1ID, player2ID *int64) (int, int) {
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

	// Calculate new ratings
	k := float64(c.kFactor)
	newELO1 := int(math.Round(float64(player1ELO) + k*(actual1-expected1)))
	newELO2 := int(math.Round(float64(player2ELO) + k*(actual2-expected2)))

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

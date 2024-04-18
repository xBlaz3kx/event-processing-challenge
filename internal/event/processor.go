package event

import (
	"fmt"
	"time"

	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
)

type DescriptionProcessor struct{}

// NewDescriptionProcessor creates a new event processor
func NewDescriptionProcessor() *DescriptionProcessor {
	return &DescriptionProcessor{}
}

// Process processes the event and returns a human-readable description of the event
// Note: Could've just used a string builder, but I wanted to keep it simple
func (p *DescriptionProcessor) Process(event casino.Event) (string, error) {
	// Validate the event type
	isValid := casino.IsValidEventType(event.Type)
	if !isValid {
		return "", fmt.Errorf("event type \"%s\" invalid", event.Type)
	}

	initialString := fmt.Sprintf("Player #%d", event.PlayerID)

	// Add an email if it's available
	if event.Player.Email != "" {
		initialString = fmt.Sprintf("%s (%s)", initialString, event.Player.Email)
	}

	// Based on the event type, add a description
	switch event.Type {
	case "game_start":
		gameDescription, err := p.getGameDescription(event.GameID)
		if err != nil {
			return "", err
		}

		// Example: Player #1 (rick@example.com) started a game "Blackjack" on February 1st, 2021 at 12:00 UTC
		initialString = fmt.Sprintf("%s started a game %s", initialString, gameDescription)
	case "bet":
		gameDescription, err := p.getGameDescription(event.GameID)
		if err != nil {
			return "", err
		}

		// Example: Player #1 (morty@example.com) bet 100 EUR on a game "Blackjack" on February 1st, 2021 at 12:00 UTC
		initialString = fmt.Sprintf("%s bet %d %s on a game %s", initialString, event.Amount, event.Currency, gameDescription)
	case "deposit":
		// Example: Player #1 (ricknmorty@example.com) made a deposit of 5 EUR on February 1st, 2021 at 12:00 UTC
		initialString = fmt.Sprintf("%s made a deposit of %d %s", initialString, event.Amount, event.Currency)

		// Example: Player #1 (rick@example.com) made a deposit of 5 USD (100 EUR) on February 1st, 2021 at 12:00 UTC
		if event.Currency != "EUR" {
			initialString = fmt.Sprintf("%s made a deposit of %d %s (%d EUR)", initialString, event.Amount, event.Currency, event.AmountEUR)
		}
	case "game_stop":
		hasWon := "lost"
		if event.HasWon {
			hasWon = "won"
		}

		gameDescription, err := p.getGameDescription(event.GameID)
		if err != nil {
			return "", err
		}

		// Example: Player #1 (rick@morty.com) has won a game "Blackjack" on February 1st, 2021 at 12:00 UTC
		// Example: Player #2 (rick@rick.com) has lost a game "Blackjack" on February 1st, 2021 at 12:00 UTC
		initialString = fmt.Sprintf("%s has %s a game %s", initialString, hasWon, gameDescription)
	}

	return fmt.Sprintf("%s on %s", initialString, formatTime(event.CreatedAt)), nil
}

// getGameDescription returns the game description based on the game ID or an error if the game is not found.
func (p *DescriptionProcessor) getGameDescription(gameId int) (string, error) {
	gameDescription, isFound := casino.Games[gameId]
	if !isFound {
		return "", fmt.Errorf("game with ID %d not found", gameId)
	}

	return gameDescription.Title, nil
}

func formatTime(t time.Time) string {
	day := t.Day()
	suffix := "th"
	switch day {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}
	return fmt.Sprintf("%s %d%s, %d at %02d:%02d %s", t.Month(), day, suffix, t.Year(), t.Hour(), t.Minute(), t.Location())
}

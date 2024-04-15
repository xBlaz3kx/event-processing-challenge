package generator

import (
	"context"
	"math/rand"
	"time"

	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
)

type Generator struct {
	eventChannel chan casino.Event
}

func New() *Generator {
	return &Generator{
		eventChannel: make(chan casino.Event),
	}
}

// Generate generates random events and sends them to the event channel
func (g *Generator) Generate(ctx context.Context) <-chan casino.Event {
	go func() {
		var id int

		for {
			select {
			case <-ctx.Done():
				return
			default:
				id++
				g.eventChannel <- g.generate(id)
			}

			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}
	}()

	return g.eventChannel
}

// Cleanup closes the event
func (g *Generator) Cleanup() {
	close(g.eventChannel)
}

// generate generates a random event
func (g *Generator) generate(id int) casino.Event {
	amount, currency := randomAmountCurrency()

	return casino.Event{
		ID:        id,
		PlayerID:  10 + rand.Intn(10),
		GameID:    100 + rand.Intn(10),
		Type:      randomType(),
		Amount:    amount,
		Currency:  currency,
		HasWon:    randomHasWon(),
		CreatedAt: time.Now(),
	}
}

func randomType() string {
	return casino.EventTypes[rand.Intn(len(casino.EventTypes))]
}

func randomAmountCurrency() (amount int, currency string) {
	currency = casino.Currencies[rand.Intn(len(casino.Currencies))]

	switch currency {
	case "BTC":
		amount = rand.Intn(1e5)
	default:
		amount = rand.Intn(2000)
	}

	return
}

func randomHasWon() bool {
	return rand.Intn(100) < 5
}

package lib

import (
	"time"
)

// ticker runs the provided function at regular intervals defined by TICKER_TIME.
// It creates a new ticker, and in a loop, it calls the provided function every time the ticker ticks.
// The ticker is stopped when the function exits.
func Ticker(argFunction func() any) {

	ticker := time.NewTicker(TICKER_TIME)
	defer ticker.Stop()

	// Start a loop that will run every 500 milliseconds to get dynamic CPU info
	for range ticker.C {
		argFunction()
	}

}

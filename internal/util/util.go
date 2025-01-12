package util

import "sync/atomic"

// UpdateAndReset updates the latest value with the current value and resets the current value to 0.
func UpdateAndReset(latest, thisSec *atomic.Int64) {
	latest.Store(thisSec.Load())
	thisSec.Store(0)
}

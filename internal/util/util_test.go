package util

import (
	"sync/atomic"
	"testing"
)

func Test_UpdateAndReset(t *testing.T) {
	var latest, thisSec atomic.Int64

	// Test case 1: Initial values
	thisSec.Store(100)
	UpdateAndReset(&latest, &thisSec)
	if latest.Load() != 100 {
		t.Errorf("Expected latest to be 100, got %d", latest.Load())
	}
	if thisSec.Load() != 0 {
		t.Errorf("Expected thisSec to be 0, got %d", thisSec.Load())
	}

	// Test case 2: Update values
	thisSec.Store(200)
	UpdateAndReset(&latest, &thisSec)
	if latest.Load() != 200 {
		t.Errorf("Expected latest to be 200, got %d", latest.Load())
	}
	if thisSec.Load() != 0 {
		t.Errorf("Expected thisSec to be 0, got %d", thisSec.Load())
	}

	// Test case 3: Zero values
	thisSec.Store(0)
	UpdateAndReset(&latest, &thisSec)
	if latest.Load() != 0 {
		t.Errorf("Expected latest to be 0, got %d", latest.Load())
	}
	if thisSec.Load() != 0 {
		t.Errorf("Expected thisSec to be 0, got %d", thisSec.Load())
	}
}

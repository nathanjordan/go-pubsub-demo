package httppubsub

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain tests for leaked goroutines during testing.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

package pubsub

import (
	"testing"

	"go.uber.org/goleak"
)

// tests for leaked goroutines during testing
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

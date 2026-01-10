package mosaic

import (
	"context"
	"testing"
	"time"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

func TestContextCancellation(t *testing.T) {
	// Use RealCommandExecutor to test actual process cancellation
	exec := &executor.RealCommandExecutor{}

	// Create a context that times out quickly
	// Increased to 300ms to avoid flakiness on slower systems
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Run a command that takes longer than the timeout
	// "sleep 2" takes 2 seconds
	start := time.Now()
	_, _, err := exec.Execute(ctx, "sleep", "2")
	duration := time.Since(start)

	// Check if we got an error (expected context deadline exceeded or signal killed)
	if err == nil {
		t.Error("expected error due to context cancellation, got nil")
	}

	// Check if it returned quickly (approx 300ms) instead of 2s
	// Allow up to 1s for slow process startup/teardown
	if duration > 1000*time.Millisecond {
		t.Errorf("expected execution to be cancelled quickly, took %v (timeout was 300ms)", duration)
	}
}

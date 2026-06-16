package aptible

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestWaitForOperationWithPoll_success(t *testing.T) {
	poll := func(_ int64) (bool, error) { return false, nil }
	ctx := context.Background()
	deleted, err := waitForOperationWithPoll(ctx, poll, 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deleted {
		t.Fatal("expected deleted=false")
	}
}

func TestWaitForOperationWithPoll_pollError(t *testing.T) {
	pollErr := errors.New("operation failed")
	poll := func(_ int64) (bool, error) { return false, pollErr }
	ctx := context.Background()
	_, err := waitForOperationWithPoll(ctx, poll, 42)
	if !errors.Is(err, pollErr) {
		t.Fatalf("expected poll error, got %v", err)
	}
}

func TestWaitForOperationWithPoll_deleted(t *testing.T) {
	poll := func(_ int64) (bool, error) { return true, nil }
	ctx := context.Background()
	deleted, err := waitForOperationWithPoll(ctx, poll, 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted=true")
	}
}

func TestWaitForOperationWithPoll_timeout(t *testing.T) {
	// poll blocks until the context is cancelled
	poll := func(_ int64) (bool, error) {
		time.Sleep(10 * time.Second) //lintignore:R018
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := waitForOperationWithPoll(ctx, poll, 99)
	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out waiting for operation 99") {
		t.Fatalf("unexpected error message: %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected error to wrap context.DeadlineExceeded, got %v", err)
	}
}

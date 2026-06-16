package aptible

import (
	"context"
	"fmt"

	"github.com/aptible/go-deploy/aptible"
)

func waitForOperationWithContext(ctx context.Context, legacy *aptible.Client, operationID int64) (bool, error) {
	return waitForOperationWithPoll(ctx, legacy.WaitForOperation, operationID)
}

func waitForOperationWithPoll(ctx context.Context, poll func(int64) (bool, error), operationID int64) (bool, error) {
	type result struct {
		deleted bool
		err     error
	}
	ch := make(chan result, 1)
	go func() {
		deleted, err := poll(operationID)
		ch <- result{deleted, err}
	}()
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("timed out waiting for operation %d: %w", operationID, ctx.Err())
	case r := <-ch:
		return r.deleted, r.err
	}
}

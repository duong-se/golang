package agent

import (
	"context"
	"fmt"
	"time"
)

const (
	MaxRetries        = 3
	MaxConsecutive529 = 2
	BaseDelayMs       = 500
)

type RecoveryState struct {
	HasEscalated         bool
	RecoveryCount        int
	Consecutive529       int
	HasAttemptedRecovery bool
	CurrentModel         string
}

func withRetry(ctx context.Context, fn func() (interface{}, error), state *RecoveryState) (interface{}, error) {
	for attempt := 0; attempt < MaxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			state.Consecutive529 = 0
			return result, nil
		}

		if isRateLimitError(err) {
			state.Consecutive529++
			delay := time.Duration(BaseDelayMs*(1<<attempt)) * time.Millisecond
			time.Sleep(delay)
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func isRateLimitError(err error) bool {
	return err != nil && (err.Error() == "rate_limit" || err.Error() == "service_error")
}

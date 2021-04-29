package requester

import (
	"time"
)

type RequestQualifierFunc func(uint64, time.Time, time.Duration) bool

func MaxAttemptQualifier(maxAttempts uint64) RequestQualifierFunc {
	return func(attempts uint64, _ time.Time, _ time.Duration) bool {
		return attempts < maxAttempts
	}
}

func RetryAfterQualifier() RequestQualifierFunc {
	return func(_ uint64, lastAttempt time.Time, retryAfter time.Duration) bool {
		nextTry := lastAttempt.Add(retryAfter)
		return nextTry.Before(time.Now())
	}
}
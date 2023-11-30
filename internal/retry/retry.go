package retry

import (
	"context"
	"errors"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
)

var RetryConfig retryConfig

func init() {
	RetryConfig = retryConfig{
		waitTime:     time.Duration(1),
		waitTimeDiff: time.Duration(2),
		retryCount:   3,
		retryErrors:  []error{repositories.ErrConnection},
	}
}

type retryConfig struct {
	waitTime     time.Duration
	waitTimeDiff time.Duration
	retryCount   int
	retryErrors  []error
}

func DefaultRetry(
	ctx context.Context,
	callback func(ctx context.Context) error,
) error {
	return Retry(
		ctx,
		callback,
		RetryConfig.waitTime,
		RetryConfig.waitTimeDiff,
		RetryConfig.retryCount,
		RetryConfig.retryErrors...,
	)
}

func Retry(
	ctx context.Context,
	callback func(ctx context.Context) error,
	waitTime time.Duration,
	waitTimeDiff time.Duration,
	retryCount int,
	retryErrors ...error,
) error {
	var err error

	// if len(retryErrors) <= 0 {
	// 	return make([]byte, 0), err
	// }

	for i := 0; i < retryCount+1; i++ {
		select {
		case <-ctx.Done():
			logger.Log.Debug("context in Retry is done")
			if err != nil {
				return err
			}
		default:

			err = callback(ctx)
			logger.Log.Debugw("Retry", "count", i, "error", err)

			if err != nil {
				shouldContinue := false

				for _, retryErr := range retryErrors {
					if errors.Is(err, retryErr) {
						shouldContinue = true
					}
				}

				if shouldContinue {
					time.Sleep(waitTime)
					waitTime += waitTimeDiff
					continue
				}

				return err
			}

			return nil
		}
	}
	return err
}

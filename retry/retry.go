package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"time"
)

const (
	// DefaultRetryTimes times of retry
	DefaultRetryTimes = 5
	// DefaultRetryDuration time duration of two retries
	DefaultRetryLinearInterval = time.Second * 3
)

// RetryConfig is config for retry
type RetryConfig struct {
	context         context.Context
	retryTimes      uint
	backoffStrategy BackoffStrategy
}

// RetryFunc 被重试执行的函数
type RetryFunc func() error

// Option is for adding retry config
type Option func(*RetryConfig)

// RetryTimes 设置重试次数，默认5
func RetryTimes(n uint) Option {
	return func(rc *RetryConfig) {
		rc.retryTimes = n
	}
}

// RetryWithCustomBackoff 设置自定义退避策略
func RetryWithCustomBackoff(backoffStrategy BackoffStrategy) Option {
	if backoffStrategy == nil {
		panic("programming error: backoffStrategy must be not nil")
	}

	return func(rc *RetryConfig) {
		rc.backoffStrategy = backoffStrategy
	}
}

// RetryWithLinearBackoff 设置线性策略退避
func RetryWithLinearBackoff(interval time.Duration) Option {
	if interval <= 0 {
		panic("programming error: retry interval should not be lower or equal to 0")
	}

	return func(rc *RetryConfig) {
		rc.backoffStrategy = &linear{
			interval: interval,
		}
	}
}

// RetryWithExponentialWithJitterBackoff 设置指数策略退避
func RetryWithExponentialWithJitterBackoff(interval time.Duration, base uint64, maxJitter time.Duration) Option {
	if interval <= 0 {
		panic("programming error: retry interval should not be lower or equal to 0")
	}

	if maxJitter < 0 {
		panic("programming error: retry maxJitter should not be lower to 0")
	}

	if base%2 == 0 {
		return func(rc *RetryConfig) {
			rc.backoffStrategy = &shiftExponentialWithJitter{
				interval:  interval,
				maxJitter: maxJitter,
				shifter:   uint64(math.Log2(float64(base))),
			}
		}
	}

	return func(rc *RetryConfig) {
		rc.backoffStrategy = &exponentialWithJitter{
			interval:  interval,
			base:      time.Duration(base),
			maxJitter: maxJitter,
		}
	}
}

// Context 设置重试context参数
func Context(ctx context.Context) Option {
	return func(rc *RetryConfig) {
		rc.context = ctx
	}
}

// Retry 重试执行函数retryFunc，直到函数运行成功，或被context停止
func Retry(retryFunc RetryFunc, opts ...Option) error {
	config := &RetryConfig{
		retryTimes: DefaultRetryTimes,
		context:    context.TODO(),
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.backoffStrategy == nil {
		config.backoffStrategy = &linear{
			interval: DefaultRetryLinearInterval,
		}
	}

	var i uint
	for i < config.retryTimes {
		err := retryFunc()
		if err != nil {
			select {
			case <-time.After(config.backoffStrategy.CalculateInterval()):
			case <-config.context.Done():
				return errors.New("retry is cancelled")
			}
		} else {
			return nil
		}
		i++
	}

	funcPath := runtime.FuncForPC(reflect.ValueOf(retryFunc).Pointer()).Name()
	lastSlash := strings.LastIndex(funcPath, "/")
	funcName := funcPath[lastSlash+1:]

	return fmt.Errorf("function %s run failed after %d times retry", funcName, i)
}

// BackoffStrategy 定义计算退避间隔的方法的接口
type BackoffStrategy interface {
	// CalculateInterval returns the time.Duration after which the next retry attempt should be made.
	CalculateInterval() time.Duration
}

// linear is a struct that implements the BackoffStrategy interface using a linear backoff strategy.
type linear struct {
	// interval specifies the fixed duration to wait between retry attempts.
	interval time.Duration
}

// CalculateInterval calculates the next interval returns a constant.
func (l *linear) CalculateInterval() time.Duration {
	return l.interval
}

// exponentialWithJitter is a struct that implements the BackoffStrategy interface using a exponential backoff strategy.
type exponentialWithJitter struct {
	base      time.Duration // base is the multiplier for the exponential backoff.
	interval  time.Duration // interval is the current backoff interval, which will be adjusted over time.
	maxJitter time.Duration // maxJitter is the maximum amount of jitter to apply to the backoff interval.
}

// CalculateInterval calculates the next backoff interval with jitter and updates the interval.
func (e *exponentialWithJitter) CalculateInterval() time.Duration {
	current := e.interval
	e.interval = e.interval * e.base
	return current + jitter(e.maxJitter)
}

// shiftExponentialWithJitter is a struct that implements the BackoffStrategy interface using a exponential backoff strategy.
type shiftExponentialWithJitter struct {
	interval  time.Duration // interval is the current backoff interval, which will be adjusted over time.
	maxJitter time.Duration // maxJitter is the maximum amount of jitter to apply to the backoff interval.
	shifter   uint64        // shift by n faster than multiplication
}

// CalculateInterval calculates the next backoff interval with jitter and updates the interval.
func (e *shiftExponentialWithJitter) CalculateInterval() time.Duration {
	current := e.interval
	e.interval = e.interval << e.shifter
	return current + jitter(e.maxJitter)
}

// Jitter adds a random duration, up to maxJitter,
func jitter(maxJitter time.Duration) time.Duration {
	if maxJitter == 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(maxJitter)) + 1)
}

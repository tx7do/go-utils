package aggregator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// ParallelFetcher 定义了并发任务的契约
type ParallelFetcher func(ctx context.Context) error

// ExecuteParallel 并行执行多个 Fetch 任务
func ExecuteParallel(ctx context.Context, fetchers []ParallelFetcher, opts ...Option) error {
	if len(fetchers) == 0 {
		return nil
	}

	// 默认配置
	defaultOpts := &options{
		limit:   defaultLimit,
		timeout: defaultTimeout, // 0 表示不限制
		retry:   defaultRetry,
	}
	for _, opt := range opts {
		opt(defaultOpts)
	}

	if defaultOpts.limit <= 0 {
		defaultOpts.limit = defaultLimit
	}

	if defaultOpts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultOpts.timeout)
		defer cancel()
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(defaultOpts.limit)

	for _, fetcher := range fetchers {
		if fetcher == nil {
			continue
		}
		f := fetcher
		g.Go(func() error {
			var lastErr error
			for i := 0; i <= defaultOpts.retry; i++ {
				if err := ctx.Err(); err != nil {
					return err
				}

				lastErr = safeRun(ctx, f)
				if lastErr == nil {
					return nil
				}

				// 如果还没达到最大重试次数，执行指数退避
				if i < defaultOpts.retry {
					// 根据当前重试次数 i 计算等待时间并 Sleep
					if err := backoffWait(ctx, i); err != nil {
						return err
					}
				}
			}
			return lastErr
		})
	}

	return g.Wait()
}

// safeRun 增加 Panic 恢复
func safeRun(ctx context.Context, f ParallelFetcher) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parallel fetcher panic: %v", r)
		}
	}()
	return f(ctx)
}

var (
	jitterRand   = rand.New(rand.NewSource(time.Now().UnixNano()))
	jitterRandMu sync.Mutex
)

// backoffWait 实现指数退避等待
func backoffWait(ctx context.Context, attempt int) error {
	// 基础参数配置
	const (
		baseDelay = 20 * time.Millisecond // 初始等待时间
		maxDelay  = 2 * time.Second       // 最大等待间隔
	)

	// 计算指数延迟: baseDelay * 2^attempt
	delay := baseDelay * (1 << uint(attempt))
	if attempt > 30 { // 2^30 * 20ms ≈ 20,000 秒，远超 maxDelay
		delay = maxDelay
	} else {
		delay = baseDelay * (1 << uint(attempt))
		if delay > maxDelay {
			delay = maxDelay
		}
	}

	// 引入随机抖动 (Jitter)，防止大量请求在同一瞬间重试
	// 实际延迟在 [0.5 * delay, 1.5 * delay] 之间波动
	jitterRandMu.Lock()
	jitter := time.Duration(jitterRand.Int63n(int64(delay)))
	jitterRandMu.Unlock()
	delay = delay/2 + jitter

	// 执行等待
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

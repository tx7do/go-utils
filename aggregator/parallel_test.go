package aggregator

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecuteParallel_Success(t *testing.T) {
	var cnt int32
	makeFetcher := func(d time.Duration) ParallelFetcher {
		return func(ctx context.Context) error {
			time.Sleep(d)
			atomic.AddInt32(&cnt, 1)
			return nil
		}
	}

	err := ExecuteParallel(t.Context(), []ParallelFetcher{
		makeFetcher(10 * time.Millisecond),
		makeFetcher(5 * time.Millisecond),
		makeFetcher(1 * time.Millisecond),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := atomic.LoadInt32(&cnt); got != 3 {
		t.Fatalf("expected 3 fetchers executed, got %d", got)
	}
}

func TestExecuteParallel_ErrorCancelsOthers(t *testing.T) {
	var started int32

	okFetcher := func(wait time.Duration) ParallelFetcher {
		return func(ctx context.Context) error {
			atomic.AddInt32(&started, 1)
			// 长时间等待，但会在 ctx 被取消时返回 ctx.Err()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
				return nil
			}
		}
	}

	errFetcher := func() ParallelFetcher {
		return func(ctx context.Context) error {
			atomic.AddInt32(&started, 1)
			return errors.New("intentional-failure")
		}
	}

	err := ExecuteParallel(t.Context(), []ParallelFetcher{
		okFetcher(200 * time.Millisecond),
		errFetcher(),
		okFetcher(200 * time.Millisecond),
	})
	if err == nil || !strings.Contains(err.Error(), "intentional-failure") {
		t.Fatalf("expected error containing %q, got %v", "intentional-failure", err)
	}

	// 确认至少有多个任务启动（并且其他任务应被取消/尽快返回）
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&started) < 2 {
		t.Fatalf("expected at least 2 fetchers started, got %d", atomic.LoadInt32(&started))
	}
}

func TestExecuteParallel_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // 立即取消上下文

	f := func(d time.Duration) ParallelFetcher {
		return func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d):
				return nil
			}
		}
	}

	err := ExecuteParallel(ctx, []ParallelFetcher{
		f(50 * time.Millisecond),
		f(50 * time.Millisecond),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
}

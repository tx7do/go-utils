package eventloop

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// 简单测试处理器：将处理到的 payload 写入 ch，并返回无错误的 Result。
type testProcessor struct {
	ch chan string
}

func (p *testProcessor) Process(ev Event) Result {
	fmt.Println("Process", ev)
	p.ch <- fmt.Sprint(ev.Data) // 阻塞写入，确保不丢事件
	return Result{Err: nil}
}

// TestSubmitPriorityProcessing 验证 Submit 后高->中->低 的处理顺序。
func TestSubmitPriorityProcessing(t *testing.T) {
	ch := make(chan string, 10)
	proc := &testProcessor{ch: ch}

	el := NewEventLoop(10, proc, false)
	if err := el.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer el.Stop()

	// 先提交低、中、高
	if err := el.Submit(NewEvent("low", "low", WithPriority(PriorityLow), WithContext(context.Background()))); err != nil {
		t.Fatalf("Submit low failed: %v", err)
	}
	if err := el.Submit(NewEvent("medium", "medium", WithPriority(PriorityMedium), WithContext(context.Background()))); err != nil {
		t.Fatalf("Submit medium failed: %v", err)
	}
	if err := el.Submit(NewEvent("high", "high", WithPriority(PriorityHigh), WithContext(context.Background()))); err != nil {
		t.Fatalf("Submit high failed: %v", err)
	}

	// 收集三个处理结果，确保顺序为 high, medium, low
	timeout := time.After(time.Second)
	var got []string
	for i := 0; i < 3; i++ {
		select {
		case s := <-ch:
			got = append(got, s)
		case <-timeout:
			t.Fatal("timeout waiting for processed events")
		}
	}

	if len(got) != 3 || got[0] != "high" || got[1] != "medium" || got[2] != "low" {
		t.Fatalf("unexpected process order: %v, want [high medium low]", got)
	}
}

// TestNilProcessorCallbackError 验证当 processor 为 nil 时，会通过回调返回错误。
func TestNilProcessorCallbackError(t *testing.T) {
	el := NewEventLoop(10, nil, false)

	// 保证回调在事件循环 goroutine 内以可控超时同步投递，避免异步分发带来的不可确定性
	el.SetCallbackInline(true, time.Second)

	if err := el.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer el.Stop()

	cb := make(chan Result, 1)
	ev := NewEvent("high", "high", WithPriority(PriorityHigh), WithContext(context.Background()), WithCallback(cb))

	if err := el.Submit(ev); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	select {
	case res := <-cb:
		if res.Err == nil || res.Err.Error() != "no event processor" {
			t.Fatalf("expected 'no event processor' error, got: %v", res.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for callback result")
	}
}

// TestCallbackDiscardWhenFull 验证当回调通道已满时，结果被非阻塞地丢弃。
func TestCallbackDiscardWhenFull(t *testing.T) {
	procCh := make(chan string, 1)
	proc := &testProcessor{ch: procCh}

	el := NewEventLoop(10, proc, false)
	el.SetCallbackInline(true, 100*time.Millisecond)

	if err := el.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer el.Stop()

	cb := make(chan Result, 1)
	cb <- Result{Err: fmt.Errorf("placeholder")} // fill the buffer

	ev := NewEvent("Medium", "p", WithPriority(PriorityMedium), WithContext(context.Background()), WithCallback(cb))
	if err := el.Submit(ev); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// 等待事件被处理（证明 Process 被调用）
	select {
	case <-procCh:
		// processed
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event to be processed")
	}

	// 此时回调应已尝试投递并因超时丢弃
	if len(cb) != 1 {
		t.Fatalf("expected callback channel length 1, got %d", len(cb))
	}

	// 确认仍是 placeholder
	select {
	case r := <-cb:
		if r.Err == nil || r.Err.Error() != "placeholder" {
			t.Fatalf("unexpected callback content: %v", r)
		}
	default:
		t.Fatal("expected to read placeholder")
	}
}

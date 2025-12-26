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
	// 非阻塞写入以防意外阻塞测试（通道有足够缓冲）
	select {
	case p.ch <- fmt.Sprint(ev.Payload):
	default:
	}
	return Result{Err: nil}
}

// TestSubmitPriorityProcessing 验证 Submit 后高->中->低 的处理顺序。
func TestSubmitPriorityProcessing(t *testing.T) {
	ch := make(chan string, 10)
	proc := &testProcessor{ch: ch}

	el := NewEventLoop(10, proc)
	if err := el.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer el.Stop()

	// 先提交低、中、高
	if err := el.Submit(Event{Priority: PriorityLow, Payload: "low", Ctx: context.Background()}); err != nil {
		t.Fatalf("Submit low failed: %v", err)
	}
	if err := el.Submit(Event{Priority: PriorityMedium, Payload: "medium", Ctx: context.Background()}); err != nil {
		t.Fatalf("Submit medium failed: %v", err)
	}
	if err := el.Submit(Event{Priority: PriorityHigh, Payload: "high", Ctx: context.Background()}); err != nil {
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
	el := NewEventLoop(10, nil)

	// 保证回调在事件循环 goroutine 内以可控超时同步投递，避免异步分发带来的不可确定性
	el.SetCallbackInline(true, time.Second)

	if err := el.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer el.Stop()

	cb := make(chan Result, 1)
	ev := Event{Priority: PriorityHigh, Callback: cb, Ctx: context.Background()}

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

	el := NewEventLoop(10, proc)

	// 使用 inline 模式并设置超时，确保在回调通道已满时主循环会在超时后放弃投递（确定性）
	el.SetCallbackInline(true, 100*time.Millisecond)

	if err := el.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer el.Stop()

	// 回调通道容量为1，先填满一个占位项
	cb := make(chan Result, 1)
	cb <- Result{Err: fmt.Errorf("placeholder")}

	ev := Event{Priority: PriorityMedium, Callback: cb, Payload: "p", Ctx: context.Background()}
	if err := el.Submit(ev); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// 等待一小段时间让事件被处理（若有额外写入会增加长度）
	time.Sleep(200 * time.Millisecond)

	// 通道长度应仍为1（未被追加）
	if l := len(cb); l != 1 {
		t.Fatalf("expected callback channel length 1, got %d", l)
	}

	// 读取并确认仍为占位项
	select {
	case r := <-cb:
		if r.Err == nil || r.Err.Error() != "placeholder" {
			t.Fatalf("unexpected callback content: %v", r)
		}
	default:
		t.Fatal("expected to read placeholder from callback channel")
	}
}

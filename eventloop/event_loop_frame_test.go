package eventloop

import (
	"context"
	"testing"
	"time"
)

// 简单处理器：将处理到的 payload 写入 ch，并可模拟处理耗时。
type frameTestProcessor struct {
	ch       chan string
	workTime time.Duration
}

func (p *frameTestProcessor) Process(ev Event) Result {
	// 模拟工作耗时，确保能被 metrics 记录
	if p.workTime > 0 {
		time.Sleep(p.workTime)
	}
	p.ch <- ev.Data.(string)
	return Result{Err: nil}
}

// TestFrameDrivenPriorityProcessing 验证帧驱动模式下按优先级处理（high->medium->low）。
func TestFrameDrivenPriorityProcessing(t *testing.T) {
	ch := make(chan string, 10)
	proc := &frameTestProcessor{ch: ch, workTime: 1 * time.Millisecond}

	el := NewEventLoop(10, proc, true)
	// 调整帧参数，加快测试速度
	el.frameInterval = 20 * time.Millisecond
	el.frameBudget = 50 * time.Millisecond
	el.maxLowTime = 10 * time.Millisecond

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
	timeout := time.After(1 * time.Second)
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

// TestFrameDrivenMetricsRecording 验证帧驱动下 handleEvent 的耗时被记录到 Metrics。
func TestFrameDrivenMetricsRecording(t *testing.T) {
	ch := make(chan string, 2)
	// 模拟较明显的处理耗时，便于观测
	proc := &frameTestProcessor{ch: ch, workTime: 5 * time.Millisecond}

	el := NewEventLoop(10, proc, true)
	el.frameInterval = 20 * time.Millisecond
	el.frameBudget = 50 * time.Millisecond
	el.maxLowTime = 10 * time.Millisecond

	metrics := NewSimpleMetrics()
	el.SetMetrics(metrics)

	if err := el.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer el.Stop()

	if err := el.Submit(NewEvent("h", "h", WithPriority(PriorityHigh), WithContext(context.Background()))); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// 等待处理完成
	select {
	case <-ch:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for processed event")
	}

	// snapshot 并检查 high 平均耗时大于 0
	snap := metrics.Snapshot()
	if snap.AvgProcessingNsHigh == 0 {
		t.Fatalf("expected AvgProcessingNsHigh > 0, got 0; snapshot: %+v", snap)
	}
}

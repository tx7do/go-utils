package eventloop

import (
	"sync/atomic"
	"time"
)

// QueueLengths 表示当前各队列的缓冲长度快照。
type QueueLengths struct {
	High     int // 高优先级队列长度（channel 存量）
	Medium   int // 中优先级队列长度
	Low      int // 低优先级队列长度（channel 中尚未被取出的项）
	Callback int // 回调分发队列长度（enqueue 使用的 channel）
}

// MetricsSnapshot 是只读快照，包含计数与平均处理耗时（纳秒）。
type MetricsSnapshot struct {
	Timestamp time.Time

	SubmittedHigh   uint64
	SubmittedMedium uint64
	SubmittedLow    uint64

	DroppedHigh   uint64
	DroppedMedium uint64
	DroppedLow    uint64

	ProcessedHigh   uint64
	ProcessedMedium uint64
	ProcessedLow    uint64

	CallbackDiscarded uint64
	InlineTimeout     uint64

	// 平均处理耗时（纳秒），0 表示无数据
	AvgProcessingNsHigh   uint64
	AvgProcessingNsMedium uint64
	AvgProcessingNsLow    uint64
}

// Metrics 定义需要的计数与时长记录操作。
type Metrics interface {
	IncSubmitted(priority Priority)
	IncDropped(priority Priority)
	IncProcessed(priority Priority)
	IncCallbackDiscarded()
	IncInlineTimeout()
	// 记录处理耗时
	ObserveProcessingDuration(priority Priority, d time.Duration)
	Snapshot() MetricsSnapshot
}

// NoopMetrics 不做任何统计，适合默认或测试。
type NoopMetrics struct{}

func (NoopMetrics) IncSubmitted(priority Priority)                               {}
func (NoopMetrics) IncDropped(priority Priority)                                 {}
func (NoopMetrics) IncProcessed(priority Priority)                               {}
func (NoopMetrics) IncCallbackDiscarded()                                        {}
func (NoopMetrics) IncInlineTimeout()                                            {}
func (NoopMetrics) ObserveProcessingDuration(priority Priority, d time.Duration) {}
func (NoopMetrics) Snapshot() MetricsSnapshot                                    { return MetricsSnapshot{Timestamp: time.Now()} }

// SimpleMetrics 基于原子计数，适合内存中统计与测试。
// 对每个优先级分别记录提交/丢弃/处理计数，以及处理总纳秒与处理次数，用于计算平均耗时。
type SimpleMetrics struct {
	submittedHigh   atomic.Uint64
	submittedMedium atomic.Uint64
	submittedLow    atomic.Uint64

	droppedHigh   atomic.Uint64
	droppedMedium atomic.Uint64
	droppedLow    atomic.Uint64

	processedHigh   atomic.Uint64
	processedMedium atomic.Uint64
	processedLow    atomic.Uint64

	callbackDiscarded atomic.Uint64
	inlineTimeout     atomic.Uint64

	// 记录处理耗时：总纳秒与计数（按优先级分别统计）
	processNsHigh    atomic.Uint64
	processCntHigh   atomic.Uint64
	processNsMedium  atomic.Uint64
	processCntMedium atomic.Uint64
	processNsLow     atomic.Uint64
	processCntLow    atomic.Uint64
}

func NewSimpleMetrics() *SimpleMetrics { return &SimpleMetrics{} }

func (s *SimpleMetrics) IncSubmitted(priority Priority) {
	switch priority {
	case PriorityHigh:
		s.submittedHigh.Add(1)
	case PriorityMedium:
		s.submittedMedium.Add(1)
	case PriorityLow:
		s.submittedLow.Add(1)
	}
}

func (s *SimpleMetrics) IncDropped(priority Priority) {
	switch priority {
	case PriorityHigh:
		s.droppedHigh.Add(1)
	case PriorityMedium:
		s.droppedMedium.Add(1)
	case PriorityLow:
		s.droppedLow.Add(1)
	}
}

func (s *SimpleMetrics) IncProcessed(priority Priority) {
	switch priority {
	case PriorityHigh:
		s.processedHigh.Add(1)
	case PriorityMedium:
		s.processedMedium.Add(1)
	case PriorityLow:
		s.processedLow.Add(1)
	}
}

func (s *SimpleMetrics) IncCallbackDiscarded() {
	s.callbackDiscarded.Add(1)
}

func (s *SimpleMetrics) IncInlineTimeout() {
	s.inlineTimeout.Add(1)
}

func (s *SimpleMetrics) ObserveProcessingDuration(priority Priority, d time.Duration) {
	nanos := uint64(d.Nanoseconds())
	switch priority {
	case PriorityHigh:
		s.processNsHigh.Add(nanos)
		s.processCntHigh.Add(1)
	case PriorityMedium:
		s.processNsMedium.Add(nanos)
		s.processCntMedium.Add(1)
	case PriorityLow:
		s.processNsLow.Add(nanos)
		s.processCntLow.Add(1)
	}
}

func (s *SimpleMetrics) Snapshot() MetricsSnapshot {
	// 计算平均纳秒（避免除以 0）
	var avgHigh, avgMed, avgLow uint64
	cntH := s.processCntHigh.Load()
	if cntH > 0 {
		avgHigh = s.processNsHigh.Load() / cntH
	}
	cntM := s.processCntMedium.Load()
	if cntM > 0 {
		avgMed = s.processNsMedium.Load() / cntM
	}
	cntL := s.processCntLow.Load()
	if cntL > 0 {
		avgLow = s.processNsLow.Load() / cntL
	}

	return MetricsSnapshot{
		Timestamp:             time.Now(),
		SubmittedHigh:         s.submittedHigh.Load(),
		SubmittedMedium:       s.submittedMedium.Load(),
		SubmittedLow:          s.submittedLow.Load(),
		DroppedHigh:           s.droppedHigh.Load(),
		DroppedMedium:         s.droppedMedium.Load(),
		DroppedLow:            s.droppedLow.Load(),
		ProcessedHigh:         s.processedHigh.Load(),
		ProcessedMedium:       s.processedMedium.Load(),
		ProcessedLow:          s.processedLow.Load(),
		CallbackDiscarded:     s.callbackDiscarded.Load(),
		InlineTimeout:         s.inlineTimeout.Load(),
		AvgProcessingNsHigh:   avgHigh,
		AvgProcessingNsMedium: avgMed,
		AvgProcessingNsLow:    avgLow,
	}
}

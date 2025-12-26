package eventloop

// EventProcessor 事件处理器接口，由业务层实现
type EventProcessor interface {
	Process(event Event) Result
}

package id

import (
	"errors"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var snowflakeNodeMap = sync.Map{}

type SnowflakeNode struct {
	workerId int64
	node     *snowflake.Node
	sync.Mutex
}

func NewSnowflakeNode(workerId int64) (*SnowflakeNode, error) {
	node, err := snowflake.NewNode(workerId)
	return &SnowflakeNode{
		workerId: workerId,
		node:     node,
		Mutex:    sync.Mutex{},
	}, err
}

func (sfNode *SnowflakeNode) Generate() int64 {
	sfNode.Lock()
	defer sfNode.Unlock()
	return sfNode.node.Generate().Int64()
}

func (sfNode *SnowflakeNode) GenerateString() string {
	sfNode.Lock()
	defer sfNode.Unlock()
	return sfNode.node.Generate().String()
}

func NewSnowflakeID(workerId int64) (int64, error) {
	// 64 位 ID = 41 位时间戳 + 10 位工作节点 ID + 12 位序列号

	var node *SnowflakeNode
	var err error
	find, ok := snowflakeNodeMap.Load(workerId)
	if ok {
		node = find.(*SnowflakeNode)
	} else {
		node, err = NewSnowflakeNode(workerId)
		if err != nil {
			//log.Println(err)
			return 0, err
		}
		snowflakeNodeMap.Store(workerId, node)
	}
	if node == nil {
		//log.Println("snowflake node is nil")
		return 0, errors.New("snowflake node is nil")
	}

	return node.Generate(), nil
}

func GenerateSnowflakeID(workerId int64) int64 {
	id, _ := NewSnowflakeID(workerId)
	return id
}

package qqwry

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient(t *testing.T) {
	g := NewClient()
	assert.NotNil(t, g)

	country, city, isp, err := g.Query("47.108.149.89")
	assert.Nil(t, err)
	fmt.Println("国家：", country, "城市：", city, "服务商：", isp)
}

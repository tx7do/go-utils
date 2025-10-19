package qqwry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	g := NewClient()
	assert.NotNil(t, g)

	res, err := g.Query("47.108.149.89")
	assert.Nil(t, err)
	fmt.Println("国家：", res.Country, "省：", res.Province, "城市：", res.City, "服务商：", res.ISP)
	assert.Equal(t, res.Country, "中国")
	assert.Equal(t, res.Province, "四川省")
	assert.Equal(t, res.City, "成都市")
	assert.Equal(t, res.ISP, "阿里云")

	res, err = g.Query("::1")
	assert.NotNil(t, err)

	res, err = g.Query("127.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, res.City, "本机地址")
}

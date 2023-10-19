package qqwry

import (
	"fmt"
	"testing"
)

func TestSpiltAddress(t *testing.T) {
	names := SpiltAddress("浙江省杭州市西湖区")
	fmt.Println(names)
}

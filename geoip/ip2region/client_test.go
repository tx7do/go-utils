package ip2region

import "testing"

func TestClient_Query(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}
	defer c.Close()

	v4Region, err := c.Query("113.92.157.29") // 进行 IPv4 查询
	t.Logf("v4Region: %+v, err: %v", v4Region, err)
	v6Region, err := c.Query("240e:3b7:3272:d8d0:db09:c067:8d59:539e") // 进行 IPv6 查询
	t.Logf("v6Region: %+v, err: %v", v6Region, err)
}

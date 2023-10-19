package geoip

// Result 归属地信息
type Result struct {
	IP       string `json:"ip"`
	Country  string `json:"country"`  // 国家
	Province string `json:"province"` // 省
	City     string `json:"city"`     // 城市
	ISP      string `json:"isp"`      // 服务提供商
}

// GeoIP 客户端
type GeoIP interface {
	Query(queryIp string) (res Result, err error)
}

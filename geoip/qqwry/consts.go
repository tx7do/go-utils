package qqwry

const (
	ipRecordLength = 7 // IndexLen 索引长度

	redirectMode1 = 0x01 // RedirectMode1 国家的类型, 指向另一个指向

	redirectMode2 = 0x02 // RedirectMode2 国家的类型, 指向一个指向
)

//var unCountry = []byte{"未知国家"}
//var unArea = []byte{"未知地区"}

// Result 归属地信息
type Result struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Area    string `json:"area"`
	ISP     string `json:"isp"`
}

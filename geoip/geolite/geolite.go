package geolite

import (
	"errors"
	"net"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/oschwald/geoip2-golang"

	"github.com/tx7do/go-utils/geoip"
	"github.com/tx7do/go-utils/geoip/geolite/assets"
)

const defaultOutputLanguage = "zh-CN"

// Client 地理位置解析结构体
type Client struct {
	db             *geoip2.Reader
	outputLanguage string
}

// NewClient .
func NewClient() (*Client, error) {
	db, err := geoip2.FromBytes(assets.GeoLite2CityData)
	if err != nil {
		return nil, err
	}
	return &Client{db: db, outputLanguage: defaultOutputLanguage}, nil
}

// Close 关闭客户端
func (g *Client) Close() error {
	if g.db == nil {
		return nil
	}
	return g.db.Close()
}

// SetLanguage 设置输出的语言，默认为：zh-CN
func (g *Client) SetLanguage(code string) {
	g.outputLanguage = code
}

// query 查询城市级别数据
func (g *Client) query(ip net.IP) (city *geoip2.City, err error) {
	return g.db.City(ip)
}

// Query 通过IP获取地区
func (g *Client) Query(rawIP string) (ret geoip.Result, err error) {
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return ret, errors.New("invalid ip address")
	}

	isPrivate := IsPrivateIP(ip)
	if isPrivate {
		ret.Country = "局域网"
		ret.Province = "局域网"
		ret.City = "局域网"
		return ret, nil
	}

	record, err := g.query(ip)
	if err != nil {
		log.Fatal(err)
		return ret, err
	}

	ret.Country = record.Country.Names[g.outputLanguage]
	if len(record.Subdivisions) > 0 {
		ret.Province = record.Subdivisions[0].Names[g.outputLanguage]
	}
	ret.City = record.City.Names[g.outputLanguage]

	return
}

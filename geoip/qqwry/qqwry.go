package qqwry

import (
	"encoding/binary"
	"errors"
	"github.com/tx7do/go-utils/geoip"
	"net"
	"strings"
	"sync"

	"github.com/tx7do/go-utils/geoip/qqwry/assets"
)

type Client struct {
	data    []byte
	dataLen uint32
	ipCache sync.Map

	IPNum int64

	startPos uint32
	endPos   uint32
}

func NewClient() *Client {
	cli := &Client{
		ipCache: sync.Map{},
	}

	cli.init()

	return cli
}

func (c *Client) init() {
	c.startPos, c.endPos = c.readHeader()
	c.IPNum = int64((c.endPos-c.startPos)/ipRecordLength + 1)
}

// parseIp 解析IP
func (c *Client) parseIp(queryIp string) (uint32, error) {
	ip := net.ParseIP(queryIp).To4()
	if ip == nil {
		return 0, errors.New("ip is not ipv4")
	}
	ip32 := binary.BigEndian.Uint32(ip)
	return ip32, nil
}

// readHeader 读取文件头
func (c *Client) readHeader() (uint32, uint32) {
	startPos := binary.LittleEndian.Uint32(assets.QQWryDat[:4])
	endPos := binary.LittleEndian.Uint32(assets.QQWryDat[4:8])
	return startPos, endPos
}

// readMode 获取偏移值类型
func (c *Client) readMode(offset uint32) byte {
	return assets.QQWryDat[offset]
}

// readIpRecord 读取IP记录 前4字节：起始IP，后3字节：偏移量
func (c *Client) readIpRecord(offset uint32) (ip32 uint32, ipOffset uint32) {
	buf := assets.QQWryDat[offset : offset+ipRecordLength]
	ip32 = binary.LittleEndian.Uint32(buf[:4])
	ipOffset = byte3ToUInt32(buf[4:])
	return ip32, ipOffset
}

// locateIP 定位IP
func (c *Client) locateIP(ip32 uint32) int32 {
	var _ip32 uint32
	var _ipOffset uint32
	var offset uint32

	var mid uint32
	i := c.startPos
	j := c.endPos
	for {
		mid = getMiddleOffset(i, j)
		_ip32, _ipOffset = c.readIpRecord(mid)

		if j-i == ipRecordLength {
			offset = _ipOffset
			_ip32, _ipOffset = c.readIpRecord(mid + ipRecordLength)
			if ip32 < _ip32 {
				break
			} else {
				offset = 0
				break
			}
		}

		if _ip32 > ip32 {
			j = mid
		} else if _ip32 < ip32 {
			i = mid
		} else if _ip32 == ip32 {
			offset = _ipOffset
			break
		}
	}

	return int32(offset)
}

// readArea 读取区域
func (c *Client) readArea(offset uint32) []byte {
	mode := c.readMode(offset)
	if mode == redirectMode1 || mode == redirectMode2 {
		areaOffset := c.readUInt24(int32(offset) + 1)
		if areaOffset == 0 {
			return []byte{}
		}
		return c.readString(areaOffset)
	}

	return c.readString(offset)
}

// readString 获取字符串
func (c *Client) readString(offset uint32) []byte {
	data := make([]byte, 0, 30)
	for i := offset; i < uint32(len(assets.QQWryDat)); i++ {
		if assets.QQWryDat[i] == 0 {
			data = assets.QQWryDat[offset:i]
			break
		}
	}
	return data
}

func (c *Client) readUInt24(offset int32) uint32 {
	i := uint32(assets.QQWryDat[offset+0]) & 0xFF
	i |= (uint32(assets.QQWryDat[offset+1]) << 8) & 0xFF00
	i |= (uint32(assets.QQWryDat[offset+2]) << 16) & 0xFF0000
	return i
}

func (c *Client) Query(queryIp string) (res geoip.Result, err error) {
	res.IP = queryIp
	res.Country = "中国"

	ip32, err := c.parseIp(queryIp)
	if err != nil {
		return
	}

	offset := c.locateIP(ip32)
	if offset <= 0 {
		err = errors.New("ip not found")
		return
	}

	//读取第一个字节判断是否是标志字节
	offset += 4
	mode := c.readMode(uint32(offset))

	var _area []byte
	var area string

	var ispPos uint32
	switch mode {
	case redirectMode1:
		posC := c.readUInt24(offset + 1)
		mode = c.readMode(posC)
		posCA := posC
		if mode == redirectMode2 {
			posCA = c.readUInt24(int32(posC) + 1)
			posC += 4
		}
		_area = c.readString(posCA)
		if mode != redirectMode2 {
			posC += uint32(len(area) + 1)
		}
		ispPos = posC

	case redirectMode2:
		posCA := c.readUInt24(offset + 1)
		_area = c.readString(posCA)
		ispPos = uint32(offset) + 4

	default:
		posCA := offset + 0
		_area = c.readString(uint32(posCA))
		ispPos = uint32(offset) + uint32(len(area)) + 1
	}

	if len(_area) != 0 {
		area = strings.TrimSpace(gb18030Decode(_area))

		areas := SpiltAddress(area)
		if len(areas) == 2 {
			res.Province = areas[0]
			res.City = areas[1]
		} else if len(areas) == 1 {
			res.City = areas[0]
		} else {
			res.City = area
		}
	}

	ispMode := assets.QQWryDat[ispPos]
	if ispMode == redirectMode1 || ispMode == redirectMode2 {
		ispPos = c.readUInt24(int32(ispPos + 1))
	}
	if ispPos > 0 {
		var _isp []byte
		_isp = c.readString(ispPos)
		res.ISP = strings.TrimSpace(gb18030Decode(_isp))
		if res.ISP != "" {
			if strings.Contains(res.ISP, "CZ88.NET") {
				res.ISP = ""
			}
		}
	}

	return
}

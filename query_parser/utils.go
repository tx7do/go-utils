package query_parser

import "net/url"

// EncodeSpecialCharacters 对字符串进行编码
func EncodeSpecialCharacters(input string) string {
	return url.QueryEscape(input)
}

// DecodeSpecialCharacters 对字符串进行解码
func DecodeSpecialCharacters(input string) (string, error) {
	return url.QueryUnescape(input)
}

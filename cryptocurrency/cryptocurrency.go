package cryptocurrency

import (
	"errors"
	"regexp"
	"strings"
)

var cryptoRegexMap = map[string]*regexp.Regexp{
	"btc":   regexp.MustCompile("^(bc1|[13])[a-zA-HJ-NP-Z0-9]{25,39}$"),
	"btg":   regexp.MustCompile("^([GA])[a-zA-HJ-NP-Z0-9]{24,34}$"),
	"dash":  regexp.MustCompile("^([X7])[a-zA-Z0-9]{33}$"),
	"dgb":   regexp.MustCompile("^(D)[a-zA-Z0-9]{24,33}$"),
	"eth":   regexp.MustCompile("^(0x)[a-zA-Z0-9]{40}$"),
	"smart": regexp.MustCompile("^(S)[a-zA-Z0-9]{33}$"),
	"xrp":   regexp.MustCompile("^(r)[a-zA-Z0-9]{33}$"),
	"zcr":   regexp.MustCompile("^(Z)[a-zA-Z0-9]{33}$"),
	"zec":   regexp.MustCompile("^(t)[a-zA-Z0-9]{34}$"),
	"xmr":   regexp.MustCompile("/4[0-9AB][1-9A-HJ-NP-Za-km-z]{93}$"),
	"trc":   regexp.MustCompile("T[A-Za-z1-9]{33}"),
}

// DetermineWalletType 判断钱包地址的类型
func DetermineWalletType(wallet string) (string, error) {
	if strings.HasPrefix(wallet, "0x") {
		if len(wallet) != 42 {
			return "", errors.New("无效的ETH地址")
		}
		return "eth", nil
	} else if strings.HasPrefix(wallet, "T") {
		if len(wallet) != 34 {
			return "", errors.New("无效的TRC地址")
		}
		return "trc", nil
	} else {
		if len(wallet) != 34 {
			return "", errors.New("无效的OMINI地址")
		}
		return "omini", nil
	}
}

// IsValidBTCAddress 是否有效的BTC地址
// 简称：OMNI
// 使用的比特币地址的正确格式：
// 1. BTC 地址是包含 26-35 个字母数字字符的标识符。
// 2. BTC 地址以数字 1、3 或 bc1 开头。
// 3. 它包含 0 到 9 范围内的数字。
// 4. 它允许使用大写和小写字母字符。
// 5. 有一点需要注意：没有使用大写字母 O、大写字母 I、小写字母 l 和数字 0，以避免视觉上的歧义。
// 6. 它不应包含空格和其他特殊字符。
func IsValidBTCAddress(address string) bool {
	return isValidCryptocurrencyAddress("btc", address)
}

// IsValidETHAddress 是否有效的ETH地址
// 简称：ERC20
func IsValidETHAddress(address string) bool {
	return isValidCryptocurrencyAddress("eth", address)
}

// IsValidTRONAddress 是否有效的TRON地址
// 简称：TRC20
func IsValidTRONAddress(address string) bool {
	return isValidCryptocurrencyAddress("trc", address)
}

func isValidCryptocurrencyAddress(crypto, address string) bool {
	if len(address) == 0 {
		return false
	}

	item, ok := cryptoRegexMap[crypto]
	if !ok {
		return false
	}

	if item.MatchString(address) {
		return true
	}

	return false
}

// IsValidCryptocurrencyAddress 校验加密货币钱包地址
func IsValidCryptocurrencyAddress(address string) string {
	if len(address) == 0 {
		return ""
	}

	for k, re := range cryptoRegexMap {
		if re.MatchString(address) {
			return k
		}
	}

	return ""
}

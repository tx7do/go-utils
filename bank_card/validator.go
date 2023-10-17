package bank_card

// IsValidLuhn 使用Luhn算法校验银行卡号码
// @see https://en.wikipedia.org/wiki/Luhn_algorithm
// @see https://www.geeksforgeeks.org/luhn-algorithm/
// @see https://medium.com/@akshaymohite/luhns-algorithm-to-validate-credit-debit-card-numbers-1952e6c7a9d0
// @see https://www.woshipm.com/pd/371041.html
func IsValidLuhn(cardNo string) bool {
	length := len(cardNo)
	if length == 0 {
		return false
	}

	if !isNumberString(cardNo) {
		return false
	}

	sum := 0
	second := false
	for i := length - 1; i >= 0; i-- {
		d := cardNo[i] - '0'

		if second == true {
			d = d * 2
		}

		sum += int(d) / 10
		sum += int(d) % 10

		second = !second
	}

	return sum%10 == 0
}

// IsValidBankCardNo 是否合法的银行卡号
func IsValidBankCardNo(cardNo string) bool {
	length := len(cardNo)
	if length < 12 || length > 19 {
		return false
	}
	return IsValidLuhn(cardNo)
}

// isNumberString 验证字符是数字
func isNumberString(s string) bool {
	length := len(s)
	for i := 0; i < length; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

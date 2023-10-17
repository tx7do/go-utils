package bank_card

type CardType string

const (
	CardTypeDC  CardType = "DC"  // 储蓄卡
	CardTypeCC  CardType = "CC"  // 信用卡
	CardTypeSCC CardType = "SCC" // 准贷记卡
	CardTypePC  CardType = "PC"  // 预付费卡
)

// Bank 银行信息
type Bank struct {
	Id       uint32 `gorm:"primarykey,column:id"`         // ID
	BankCode string `gorm:"uniqueIndex,column:bank_code"` // 银行简称
	BankName string `gorm:"column:bank_name"`             // 银行名称
}

// BankCard 银行卡信息
type BankCard struct {
	BIN        uint32 `gorm:"primarykey,column:bin"` // 银行识别码
	BankCode   string `gorm:"column:bank_code"`      // 银行代码
	BankName   string // 银行名称
	CardType   string `gorm:"column:card_type"`   // 银行卡类型
	CardName   string `gorm:"column:card_name"`   // 银行卡名称
	CardLength uint32 `gorm:"column:card_length"` // 银行卡号长度
}

// CardTypeName 将卡类型转为类型名
func (b *BankCard) CardTypeName() string {
	switch CardType(b.CardType) {
	case CardTypeDC:
		return "储蓄卡"
	case CardTypeCC:
		return "信用卡"
	case CardTypeSCC:
		return "准贷记卡"
	case CardTypePC:
		return "预付费卡"
	}
	return ""
}

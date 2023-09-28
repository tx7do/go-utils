package cryptocurrency

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidCryptocurrencyAddress(t *testing.T) {
	type walletAddressInput struct {
		Crypto  string
		Address string
	}

	var validWalletAddresses = []walletAddressInput{
		{Crypto: "btc", Address: "1CFNjwLjZdSKB8nZopxhLaR8vvqaQKD3Bi"}, //old btc type
		{Crypto: "BTC", Address: "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq"},
		{Crypto: "BTC", Address: "1RAHUEYstWetqabcFn5Au4m4GFg7xJaNVN2"},
		{Crypto: "BTC", Address: "3J98t1RHT73CNmQwertyyWrnqRhWNLy"},
		{Crypto: "BTC", Address: "bc1qarsrrr7ASHy5643ydab9re59gtzzwfrah"},
		// {Crypto: "bch", Address: "qq7ujnfl6tqx7xcdsdsrsqlqgqz8rm5stsvgx2kcvu"},                     // cash address
		// {Crypto: "bch", Address: "bitcoincash:qq7ujnfl6tqx7xcdsdsrsqlqgqz8rm5stsvgx2kcvu"},         // cash address
		// {Crypto: "bch", Address: "16dhNPnPp346wzrRTkArKhqPM1ELeJDvRr"},
		{Crypto: "BTG", Address: "GakMJVF7Du16VK9dpN6nhJyLUPLXkTfqSY"},
		{Crypto: "DGB", Address: "D59P8MiMXkjs7HPn31zAnUSvRNwvNZUBYa"},
		{Crypto: "DASH", Address: "XiHMBEic8q8wX5aKqVv6zRFec7cAuYGjBV"},
		{Crypto: "ETH", Address: "0x15cc4bf4fe84fea178d2b10f89f1a6c914dfc8c2"},
		{Crypto: "ETH", Address: "0x323b5d4c32345ced77393b3530b1eed0f346429d"},
		{Crypto: "ETH", Address: "0xZYXb5d4c32345ced77393b3530b1eed0f346429d"},
		{Crypto: "ETH", Address: "0xe41d2489571d322189246dafa5ebde1f4699f498"},
		{Crypto: "ETH", Address: "0x8e215d06ea7ec1fdb4fc5fd21768f4b34ee92ef4"},
		{Crypto: "SMART", Address: "SbsLb8eM583oraW89qhbkcqZmuR4aYKkea"},
		{Crypto: "XRP", Address: "rMkfgicNKuCfXojDhcX4W2LnGoHFqhFrr6"},
		{Crypto: "ZEC", Address: "t1SBt3V8MfG4ZJ2ZDTuWfDshn4PuyvqjJV3"},
		{Crypto: "ZCR", Address: "ZXvpr2M6wvKoFcTJ57WCjT9Wkd38xkL8Fo"},
		{Crypto: "trc", Address: "TC74QG8tbtixG5Raa4fEifywgjrFs45fNz"},
		{Crypto: "trc", Address: "TFUD8x3iAZ9dF7NDCGBtSjznemEomE5rP9"},
		{Crypto: "trc", Address: "TPcKtz5TRfP4xUZSos81RmXB9K2DBqj2iu"},
	}

	var invalidWalletAddresses = []walletAddressInput{
		{Crypto: "btc", Address: "2CFNjwLjZdSKB8nZopxhLaR8vvqaQKD3Bi"},
		{Crypto: "BTC", Address: "bc2qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq"},
		{Crypto: "BTC", Address: "b1qarsrrr7ASHy5643ydab9re59gtzzwfrah"},
		{Crypto: "BTC", Address: "0J98t1RHT73CNmQwertyyWrnqRhWNLy"},
		{Crypto: "BTG", Address: "DakMJVF7Du16VK9dpN6nhJyLUPLXkTfqSY"},
		{Crypto: "DGB", Address: "G59P8MiMXkjs7HPn31zAnUSvRNwvNZUBYa"},
		{Crypto: "DASH", Address: "QiHMBEic8q8wX5aKqVv6zRFec7cAuYGjBV"},
		{Crypto: "ETH", Address: "1x15cc4bf4fe84fea178d2b10f89f1a6c914dfc8c2"},
		{Crypto: "SMART", Address: "sbsLb8eM583oraW89qhbkcqZmuR4aYKkea"},
		{Crypto: "XRP", Address: "RMkfgicNKuCfXojDhcX4W2LnGoHFqhFrr6"},
		{Crypto: "ZEC", Address: "z1SBt3V8MfG4ZJ2ZDTuWfDshn4PuyvqjJV3"},
		{Crypto: "ZCR", Address: "zXvpr2M6wvKoFcTJ57WCjT9Wkd38xkL8Fo"},
	}

	for _, w := range validWalletAddresses {
		t.Run("valid address "+w.Crypto, func(t *testing.T) {
			result := IsValidCryptocurrencyAddress(w.Address)
			assert.True(t, strings.Compare(strings.ToLower(w.Crypto), result) == 0)
		})
	}

	for _, w := range invalidWalletAddresses {
		t.Run("invalid address "+w.Crypto, func(t *testing.T) {
			result := IsValidCryptocurrencyAddress(w.Address)
			assert.False(t, strings.Compare(strings.ToLower(w.Crypto), result) == 0)
		})
	}
}

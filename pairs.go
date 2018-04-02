package bitcoinaverage

// Pair is a trading pair
type Pair string

const (
	// BTCUSD is the price of 1 BTC in USD
	BTCUSD = Pair("BTCUSD")
	// BTCCNY is the price of 1 BTC in CNY
	BTCCNY = Pair("BTCCNY")
)

// GetBase gets the first currency in the pair
func (p Pair) GetBase() string {
	if len(p) < 3 {
		return ""
	}
	return string(p[0:3])
}

// GetCounter gets the second currency in the pair
func (p Pair) GetCounter() string {
	if len(p) < 6 {
		return ""
	}
	return string(p[3:6])
}

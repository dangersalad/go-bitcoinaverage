package bitcoinaverage

import (
	"os"
	"testing"
)

const btcAveragePublicKey = "BTC_AVERAGE_PUBLIC_KEY"
const btcAverageSecretKey = "BTC_AVERAGE_SECRET_KEY"

func getTestingClient(t *testing.T) *Client {
	pubkey := os.Getenv(btcAveragePublicKey)
	if pubkey == "" {
		t.Fatalf("Missing %s from env", btcAveragePublicKey)
	}
	secret := os.Getenv(btcAverageSecretKey)
	if secret == "" {
		t.Fatalf("Missing %s from env", btcAverageSecretKey)
	}
	return NewClient(pubkey, secret)

}

func TestTickers(t *testing.T) {
	c := getTestingClient(t)
	data, err := c.Tickers([]string{"BTC", "LTC"}, []string{"USD", "GBP"})
	if err != nil {
		t.Error(err)
	}
	for p, d := range data {
		t.Logf("%s: %#v", p, d)
	}
}

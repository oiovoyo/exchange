package exchange

import (
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestBinanceUpdateExchangeInfo(t *testing.T) {
	err := BinanceUpdateExchangeInfo()
	t.Log(err)

	BinanceExchangeInfo.Range(func(key, value interface{}) bool {
		t.Log(key, value)
		return true
	})

}
func TestTruncatePrice(t *testing.T) {

	pair := "GASBTC"

	testA := []struct {
		Before float64
		Trunc  PriceTrunc
		After  string
	}{
		{0.01, PRICE_DOWN, "0.010000"},
		{0.01, PRICE_UP, "0.010000"},

		{0.000001, PRICE_UP, "0.000001"},
		{0.000001, PRICE_DOWN, "0.000001"},

		{0.0000009, PRICE_DOWN, "0.000001"},
		{0.0000009, PRICE_UP, "0.000001"},

		{0.0000112, PRICE_UP, "0.000012"},
		{0.0000112, PRICE_DOWN, "0.000011"},

		{0.0009992, PRICE_UP, "0.001000"},
		{0.0009992, PRICE_DOWN, "0.000999"},
	}

	for _, v := range testA {
		r := TruncatePrice(pair, v.Before, v.Trunc)
		if r != v.After {
			t.Logf("TruncatePrice(\"%s\",%0.8f,%v) = [\"%s\"] expect [\"%s\"]",
				pair, v.Before, v.Trunc, r, v.After,
			)
		}
	}

}

func TestTruncateAmount(t *testing.T) {

	pair := "GASBTC"
	// min 0.01
	testA := []struct {
		Before float64
		After  string
	}{
		{0.01, "0.01"},
		{0.02, "0.02"},

		{0.000001, "0.00"},

		{10000.0000112, "10000.00"},

		{10000.019, "10000.01"},
	}

	for _, v := range testA {
		r := TruncateAmount(pair, v.Before)
		if r != v.After {
			t.Logf("TruncateAmount(\"%s\",%0.8f) = [\"%s\"] expect [\"%s\"]",
				pair, v.Before, r, v.After,
			)
		}
	}
}

var (
	apiKey    = ""
	secretKey = ""

	proxyString = "https://127.0.0.1:1087"
	proxyUrl, _ = url.Parse(proxyString)
	t           = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		// We use ABSURDLY large keys, and should probably not.
		TLSHandshakeTimeout: 60 * time.Second,
		Proxy:               http.ProxyURL(proxyUrl),
	}
	c0 = &http.Client{
		Transport: t,
		Timeout:   15 * time.Second,
	}

	c = NewCustomBinance(apiKey, secretKey, c0)
)

func TestBinance_tradeOneTime(t *testing.T) {
	//v,e := c.tradeOneTime("buy","BTC_USDT",0.01,6400.0004)
	//v,e := c.tradeOneTime("buy","GAS_BTC",0.5,0.00282455)
	v, e := c.tradeOneTime("buy", "GAS_BTC", 0.4995, 0.0022055456)
	t.Log(v, e)
}

func TestBinance_CancelOpenOrders(t *testing.T) {
	e := c.CancelOpenOrders("GAS_BTC")
	t.Log(e)
}

func TestBinance_GetOrderBook(t *testing.T) {
	v, _ := c.GetOrderBook("GAS_BTC", 50)
	t.Log(v.Buy)
	t.Log(v.Sell)
}

func TestBinance_Withdraw(t *testing.T) {
	v, e := c.Withdraw("BTC", 0.0098, "", "")
	t.Log(v, e)
}

func TestBinance_GetWithdrawStatus(t *testing.T) {
	v, e := c.GetWithdrawStatus("")
	t.Log(v, e)
}

func TestBinance_GetDepositList(t *testing.T) {
	v, e := c.GetDepositList("USDT", time.Now().AddDate(0, 0, -5).UTC())
	t.Log(v, e)
}

func TestBinance_CheckWalletValid(t *testing.T) {
	v, e := c.CheckWalletValid("USDT")
	t.Log(v, e)
}

func TestBinance_GetTradingBalance(t *testing.T) {
	v, e := c.GetTradingBalance("RCN")
	t.Log(v, e)
}

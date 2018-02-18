package exchange

import (
	"testing"
)

func TestExxGetOrderBook(t *testing.T) {
	b, e := NewExx("", "").GetOrderBook("BTC_USDT", 50)
	if e != nil {
		t.Error(e)
	}
	t.Logf("%+v", b)
}
func TestExxMarkets(t *testing.T) {
	b, e := NewExx("", "").Markets()
	if e != nil {
		t.Error(e)
	}
	t.Log(b)
}

package exchange

import (
	"testing"
)

func TestZBGetOrderBook(t *testing.T) {
	b, e := NewZB("", "").GetOrderBook("QTUM_BTC", 50)
	if e != nil {
		t.Error(e)
	}
	t.Logf("%+v", b)
}
func TestZBMarkets(t *testing.T) {
	b, e := NewZB("", "").Markets()
	if e != nil {
		t.Error(e)
	}
	t.Log(b)
}

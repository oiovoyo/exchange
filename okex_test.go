package exchange

import (
	"testing"
)

func TestOkExGetOrderBook(t *testing.T) {
	b, e := NewOkEx("", "").GetOrderBook("QTUM_BTC", 50)
	if e != nil {
		t.Error(e)
	}
	t.Logf("%+v", b)
}
func TestOkExMarkets(t *testing.T) {
	b, e := NewOkEx("", "").Markets()
	if e != nil {
		t.Error(e)
	}
	t.Log(b)
}

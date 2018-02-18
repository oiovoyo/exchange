package exchange

import (
	"testing"
)

func TestHuoBiGetOrderBook(t *testing.T) {
	b, e := NewHuoBi("", "").GetOrderBook("QTUM_BTC", 50)
	if e != nil {
		t.Error(e)
	}
	t.Logf("%+v", b)
}
func TestHuoBiMarkets(t *testing.T) {
	b, e := NewHuoBi("", "").Markets()
	if e != nil {
		t.Error(e)
	}
	t.Log(b)
}

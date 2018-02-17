package exchange

import (
	"testing"
)

func TestExxGetOrderBook(t *testing.T) {
	b, e := NewExx("", "").GetOrderBook("BTC_QTUM", 50)
	if e != nil {
		t.Error(e)
	}
	t.Log(b)
}
func TestExxMarkets(t *testing.T) {
	b, e := NewExx("", "").Markets()
	if e != nil {
		t.Error(e)
	}
	t.Log(b)
}

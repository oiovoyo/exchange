package exchange

import (
	"testing"
)

func TestPoloniexMarkets(t *testing.T) {
	m, e := NewPoloniex("", "").Markets()
	t.Log(e)
	t.Log(m)

}

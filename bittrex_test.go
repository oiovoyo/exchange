package exchange

import (
	"testing"
)

func TestBittrexMarkets(t *testing.T) {
	m, e := NewBittrex("", "").Markets()
	t.Log(e)
	t.Log(m)

}

package exchange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

/*
{
  "zb_qc": {
    "amountScale": 2,
    "priceScale": 4
  },
  "zb_usdt": {
    "amountScale": 2,
    "priceScale": 4
  }
}
*/
type ZBInfoItem struct {
	AmountScale int    `json:"amountScale"`
	PriceScale  int    `json:"priceScale"`
	Base        string `json:"base"`
	Quot        string `json:"quot"`
}

var (
	ZBInfo sync.Map
)

type ZB struct {
}

func NewZB(access, secret string) *ZB {
	return &ZB{}
}

func init() {
	err := NewZB("", "").ZBUpdateExchangeInfo()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func (z *ZB) Markets() ([]Market, error) {
	m := make([]Market, 0)
	ZBInfo.Range(func(key, value interface{}) bool {

		v := value.(ZBInfoItem)
		m = append(m, Market{Base: v.Base, Quot: v.Quot})
		return true
	})
	return m, nil
}
func (z *ZB) ZBUpdateExchangeInfo() error {
	url := fmt.Sprintf("http://api.zb.com/data/v1/markets")
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return err // handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	zbInfo := make(map[string]ZBInfoItem)
	err = json.Unmarshal(body, &zbInfo)
	if err != nil {
		return err
	}
	//fmt.Println(ZBInfo)
	for k, v := range zbInfo {
		pair := strings.Split(k, "_")
		base := strings.ToUpper(pair[1])
		quot := strings.ToUpper(pair[0])
		v.Base = base
		v.Quot = quot
		ZBInfo.Store(quot+"_"+base, v)
	}
	return nil
}

func (z *ZB) MakeLocalPair(pair string) string {

	s := strings.Split(pair, "_")
	if len(s) != 2 {
		return "err-pair"
	}
	return strings.ToLower(s[0] + "_" + s[1])

}

func (z *ZB) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
	pair = z.MakeLocalPair(pair)
	return 0.0, fmt.Errorf("not impl")
}

func (z *ZB) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
	return z.tradeOneTime("buy", pair, buyAmount, price)
}
func (z *ZB) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
	return z.tradeOneTime("sell", pair, sellAmount, price)
}
func (z *ZB) CancelOpenOrders(pair string) error {
	pair = z.MakeLocalPair(pair)
	return fmt.Errorf("not impl")
}
func (z *ZB) GetOrderBook(pairString string, depthSize int) (*OrderBook, error) {
	pairString = z.MakeLocalPair(pairString)
	url := fmt.Sprintf("http://api.zb.com/data/v1/depth?market=%s&size=%d", pairString, depthSize)
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err // handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	depth_ := struct {
		Timestamp int64       `json:"timestamp"`
		Asks      [][]float64 `json:"asks"`
		Bids      [][]float64 `json:"bids"`
	}{}
	if err != nil {
		return nil, err // handle error
	}
	err = json.Unmarshal(body, &depth_)
	if err != nil {
		return nil, err // handle error
	}
	//fmt.Println(depth_)
	orderBook := OrderBook{Buy: make([]Orderb, 0), Sell: make([]Orderb, 0)}
	for _, buy := range depth_.Bids {
		rate := buy[0]
		quan := buy[1]
		oneOrder := Orderb{Rate: rate, Quantity: quan}
		orderBook.Buy = append(orderBook.Buy, oneOrder)
	}
	len_ := len(depth_.Asks)
	for i := len_ - 1; i > 0; i-- {
		sell := depth_.Asks[i]
		rate := sell[0]
		quan := sell[1]
		oneOrder := Orderb{Rate: rate, Quantity: quan}
		orderBook.Sell = append(orderBook.Sell, oneOrder)
	}
	/*if err != nil {
		return nil, fmt.Errorf("ZB.GetOrderBook(\"%s\",%d) error %v", pairString, depthSize, err)
	}
	return &orderBook, err
	*/
	return &orderBook, nil
}

/*set "" if coin has no paymentid*/
func (z *ZB) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {
	return "000", fmt.Errorf("not impl")
}
func (z *ZB) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
	return WITHDRAW_ERROR, fmt.Errorf("not impl")
}
func (z *ZB) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
	return nil, fmt.Errorf("not impl")
}
func (z *ZB) CheckWalletValid(coin string) (bool, error) {
	return true, nil
	return false, fmt.Errorf("not impl")
}
func (z *ZB) GetTradingBalance(currency string) (float64, error) {
	return 0.0, fmt.Errorf("not impl")
}
func (z *ZB) GetPaymentBalance(currency string) (float64, error) {
	return z.GetTradingBalance(currency)
}
func (z *ZB) TransferToPayment(currency string, amount float64) error {
	return nil
}
func (z *ZB) TransferToTrading(currency string, amount float64) error {
	return nil
}
func (z *ZB) GetDepositAddress(currency string) (string, string, error) {
	return "", "", fmt.Errorf("not impl")
}

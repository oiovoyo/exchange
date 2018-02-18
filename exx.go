package exchange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
{
  "etc_usdt": {
    "minAmount": "0.00010",
    "amountScale": 2,
    "priceScale": 2,
    "maxLevels": 0,
    "isOpen": true
  },
  "ubtc_usdt": {
    "minAmount": "0.001",
    "amountScale": 3,
    "priceScale": 2,
    "maxLevels": 0,
    "isOpen": true
  }
}
*/
type ExxInfoItem struct {
	MinAmount   float64 `json:"minAmount,string"`
	AmountScale int     `json:"amountScale"`
	PriceScale  int     `json:"priceScale"`
	MaxLevels   int     `json:"maxLevels"`
	IsOpen      bool    `json:"isOpen"`
	Base        string  `json:"base"`
	Quot        string  `json:"quot"`
}

var (
	ExxInfo sync.Map
)

type Exx struct {
}

func NewExx(access, secret string) *Exx {
	return &Exx{}
}

func init() {
	err := NewExx("", "").ExxUpdateExchangeInfo()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func (e *Exx) Markets() ([]Market, error) {
	m := make([]Market, 0)
	ExxInfo.Range(func(key, value interface{}) bool {

		v := value.(ExxInfoItem)
		m = append(m, Market{Base: v.Base, Quot: v.Quot})
		return true
	})
	return m, nil
}
func (e *Exx) ExxUpdateExchangeInfo() error {
	url := fmt.Sprintf("https://api.exx.com/data/v1/markets")
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
	exxInfo := make(map[string]ExxInfoItem)
	err = json.Unmarshal(body, &exxInfo)
	if err != nil {
		return err
	}
	//fmt.Println(exxInfo)
	for k, v := range exxInfo {
		pair := strings.Split(k, "_")
		base := strings.ToUpper(pair[1])
		quot := strings.ToUpper(pair[0])
		v.Base = base
		v.Quot = quot
		ExxInfo.Store(quot+"_"+base, v)
	}
	return nil
}

func (e *Exx) MakeLocalPair(pair string) string {

	s := strings.Split(pair, "_")
	if len(s) != 2 {
		return "err-pair"
	}
	return strings.ToLower(s[0] + "_" + s[1])

}

func (e *Exx) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
	pair = e.MakeLocalPair(pair)
	return 0.0, fmt.Errorf("not impl")
}

func (e *Exx) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
	return e.tradeOneTime("buy", pair, buyAmount, price)
}
func (e *Exx) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
	return e.tradeOneTime("sell", pair, sellAmount, price)
}
func (e *Exx) CancelOpenOrders(pair string) error {
	pair = e.MakeLocalPair(pair)
	return fmt.Errorf("not impl")
}
func (e *Exx) GetOrderBook(pairString string, depthSize int) (*OrderBook, error) {
	pairString = e.MakeLocalPair(pairString)
	url := fmt.Sprintf("https://api.exx.com/data/v1/depth?currency=%s", pairString)
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err // handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	depth_ := struct {
		Timestamp int64      `json:"timestamp"`
		Asks      [][]string `json:"asks"`
		Bids      [][]string `json:"bids"`
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
		rate, _ := strconv.ParseFloat(buy[0], 64)
		quan, _ := strconv.ParseFloat(buy[1], 64)
		oneOrder := Orderb{Rate: rate, Quantity: quan}
		orderBook.Buy = append(orderBook.Buy, oneOrder)
	}
	len_ := len(depth_.Asks)
	for i := len_ - 1; i > 0; i-- {
		sell := depth_.Asks[i]
		rate, _ := strconv.ParseFloat(sell[0], 64)
		quan, _ := strconv.ParseFloat(sell[1], 64)
		oneOrder := Orderb{Rate: rate, Quantity: quan}
		orderBook.Sell = append(orderBook.Sell, oneOrder)
	}
	/*if err != nil {
		return nil, fmt.Errorf("Exx.GetOrderBook(\"%s\",%d) error %v", pairString, depthSize, err)
	}
	return &orderBook, err
	*/
	return &orderBook, nil
}

/*set "" if coin has no paymentid*/
func (e *Exx) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {
	return "000", fmt.Errorf("not impl")
}
func (e *Exx) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
	return WITHDRAW_ERROR, fmt.Errorf("not impl")
}
func (e *Exx) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
	return nil, fmt.Errorf("not impl")
}
func (e *Exx) CheckWalletValid(coin string) (bool, error) {
	return true, nil
	return false, fmt.Errorf("not impl")
}
func (e *Exx) GetTradingBalance(currency string) (float64, error) {
	return 0.0, fmt.Errorf("not impl")
}
func (e *Exx) GetPaymentBalance(currency string) (float64, error) {
	return e.GetTradingBalance(currency)
}
func (e *Exx) TransferToPayment(currency string, amount float64) error {
	return nil
}
func (e *Exx) TransferToTrading(currency string, amount float64) error {
	return nil
}
func (e *Exx) GetDepositAddress(currency string) (string, string, error) {
	return "", "", fmt.Errorf("not impl")
}

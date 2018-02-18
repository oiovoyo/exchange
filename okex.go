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
  "code": 0,
  "data": [
    {
      "baseCurrency": 1,
      "collect": "0",
      "isMarginOpen": false,
      "marginRiskPreRatio": 0,
      "marginRiskRatio": 0,
      "marketFrom": 103,
      "maxMarginLeverage": 0,
      "maxPriceDigit": 8,
      "maxSizeDigit": 8,
      "minTradeSize": 0.01,
      "online": 1,
      "productId": 12,
      "quoteCurrency": 0,
      "sort": 10013,
      "symbol": "ltc_btc"
    },
    {
      "baseCurrency": 2,
      "collect": "0",
      "isMarginOpen": true,
      "marginRiskPreRatio": 1.3,
      "marginRiskRatio": 1.1,
      "marketFrom": 104,
      "maxMarginLeverage": 3,
      "maxPriceDigit": 8,
      "maxSizeDigit": 8,
      "minTradeSize": 0.01,
      "online": 1,
      "productId": 13,
      "quoteCurrency": 0,
      "sort": 10014,
      "symbol": "eth_btc"
    }]
}
*/
type OkExInfoItem struct {
	MinTradeSize float64 `json:"minTradeSize"`
	Symbol       string  `json:"symbol"`
	Base         string  `json:"base"`
	Quot         string  `json:"quot"`
}

var (
	OkExInfo sync.Map
)

type OkEx struct {
}

func NewOkEx(access, secret string) *OkEx {
	return &OkEx{}
}

func init() {
	err := NewOkEx("", "").OkExUpdateExchangeInfo()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func (o *OkEx) Markets() ([]Market, error) {
	m := make([]Market, 0)
	OkExInfo.Range(func(key, value interface{}) bool {

		v := value.(OkExInfoItem)
		m = append(m, Market{Base: v.Base, Quot: v.Quot})
		return true
	})
	return m, nil
}
func (o *OkEx) OkExUpdateExchangeInfo() error {
	url := fmt.Sprintf("https://www.okex.com/v2/markets/products")
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
	okExInfo := struct {
		Code int            `json:"code"`
		Data []OkExInfoItem `json:"data"`
	}{}
	err = json.Unmarshal(body, &okExInfo)
	if err != nil {
		return err
	}
	//fmt.Println(OkExInfo)
	for _, v := range okExInfo.Data {
		pair := strings.Split(v.Symbol, "_")
		base := strings.ToUpper(pair[1])
		quot := strings.ToUpper(pair[0])
		v.Base = base
		v.Quot = quot
		OkExInfo.Store(quot+"_"+base, v)
	}
	return nil
}

func (o *OkEx) MakeLocalPair(pair string) string {

	s := strings.Split(pair, "_")
	if len(s) != 2 {
		return "err-pair"
	}
	return strings.ToLower(s[0] + "_" + s[1])

}

func (o *OkEx) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
	pair = o.MakeLocalPair(pair)
	return 0.0, fmt.Errorf("not impl")
}

func (o *OkEx) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
	return o.tradeOneTime("buy", pair, buyAmount, price)
}
func (o *OkEx) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
	return o.tradeOneTime("sell", pair, sellAmount, price)
}
func (o *OkEx) CancelOpenOrders(pair string) error {
	pair = o.MakeLocalPair(pair)
	return fmt.Errorf("not impl")
}
func (o *OkEx) GetOrderBook(pairString string, depthSize int) (*OrderBook, error) {
	pairString = o.MakeLocalPair(pairString)
	url := fmt.Sprintf("https://www.okex.com/api/v1/depth.do?symbol=%s&size=%d", pairString, depthSize)
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err // handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	depth_ := struct {
		Asks [][]float64 `json:"asks"`
		Bids [][]float64 `json:"bids"`
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
		return nil, fmt.Errorf("OkEx.GetOrderBook(\"%s\",%d) error %v", pairString, depthSize, err)
	}
	return &orderBook, err
	*/
	return &orderBook, nil
}

/*set "" if coin has no paymentid*/
func (o *OkEx) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {
	return "000", fmt.Errorf("not impl")
}
func (o *OkEx) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
	return WITHDRAW_ERROR, fmt.Errorf("not impl")
}
func (o *OkEx) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
	return nil, fmt.Errorf("not impl")
}
func (o *OkEx) CheckWalletValid(coin string) (bool, error) {
	return true, nil
	return false, fmt.Errorf("not impl")
}
func (o *OkEx) GetTradingBalance(currency string) (float64, error) {
	return 0.0, fmt.Errorf("not impl")
}
func (o *OkEx) GetPaymentBalance(currency string) (float64, error) {
	return o.GetTradingBalance(currency)
}
func (o *OkEx) TransferToPayment(currency string, amount float64) error {
	return nil
}
func (o *OkEx) TransferToTrading(currency string, amount float64) error {
	return nil
}
func (o *OkEx) GetDepositAddress(currency string) (string, string, error) {
	return "", "", fmt.Errorf("not impl")
}

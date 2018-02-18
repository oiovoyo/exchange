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
  "status": "ok",
  "data": [
    {
      "base-currency": "nas",
      "quote-currency": "eth",
      "price-precision": 6,
      "amount-precision": 4,
      "symbol-partition": "innovation"
    },
    {
      "base-currency": "eos",
      "quote-currency": "eth",
      "price-precision": 8,
      "amount-precision": 2,
      "symbol-partition": "main"
    }
  ]
}
*/
// base is quot, quot is base, wtf
type HuoBiInfoItem struct {
	AmountPrecision int    `json:"amount-precision"`
	PricePrecision  int    `json:"price-precision"`
	Quot            string `json:"base-currency"`
	Base            string `json:"quote-currency"`
}

var (
	HuoBiInfo sync.Map
)

type HuoBi struct {
}

func NewHuoBi(access, secret string) *HuoBi {
	return &HuoBi{}
}

func init() {
	err := NewHuoBi("", "").HuoBiUpdateExchangeInfo()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func (b *HuoBi) Markets() ([]Market, error) {
	m := make([]Market, 0)
	HuoBiInfo.Range(func(key, value interface{}) bool {

		v := value.(HuoBiInfoItem)
		m = append(m, Market{Base: v.Base, Quot: v.Quot})
		return true
	})
	return m, nil
}
func (b *HuoBi) HuoBiUpdateExchangeInfo() error {
	url := fmt.Sprintf("https://api.huobi.pro/v1/common/symbols")
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
	//fmt.Printf(string(body))
	huoBiInfo := struct {
		Status string          `json:"status"`
		Data   []HuoBiInfoItem `json:"data"`
	}{}
	err = json.Unmarshal(body, &huoBiInfo)
	if err != nil {
		return err
	}
	//fmt.Println(HuoBiInfo)
	for _, v := range huoBiInfo.Data {
		v.Base = strings.ToUpper(v.Base)
		v.Quot = strings.ToUpper(v.Quot)
		v.Base = v.Base
		v.Quot = v.Quot
		HuoBiInfo.Store(v.Quot+"_"+v.Base, v)
	}
	return nil
}

func (b *HuoBi) MakeLocalPair(pair string) string {

	s := strings.Split(pair, "_")
	if len(s) != 2 {
		return "err-pair"
	}
	return strings.ToLower(s[0] + s[1])

}

func (b *HuoBi) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
	pair = b.MakeLocalPair(pair)
	return 0.0, fmt.Errorf("not impl")
}

func (b *HuoBi) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
	return b.tradeOneTime("buy", pair, buyAmount, price)
}
func (b *HuoBi) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
	return b.tradeOneTime("sell", pair, sellAmount, price)
}
func (b *HuoBi) CancelOpenOrders(pair string) error {
	pair = b.MakeLocalPair(pair)
	return fmt.Errorf("not impl")
}
func (b *HuoBi) GetOrderBook(pairString string, depthSize int) (*OrderBook, error) {
	pairString = b.MakeLocalPair(pairString)
	url := fmt.Sprintf("https://api.huobi.pro/market/depth?symbol=%s&type=step0", pairString)
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err // handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	depth_ := struct {
		Status string `json:"status"`
		Ch     string `json:"ch"`
		Ts     int64  `json:"ts"`
		Tick   struct {
			Asks [][]float64 `json:"asks"`
			Bids [][]float64 `json:"bids"`
		} `json:"tick"`
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
	for _, buy := range depth_.Tick.Bids {
		rate := buy[0]
		quan := buy[1]
		oneOrder := Orderb{Rate: rate, Quantity: quan}
		orderBook.Buy = append(orderBook.Buy, oneOrder)
	}
	for _, sell := range depth_.Tick.Asks {
		rate := sell[0]
		quan := sell[1]
		oneOrder := Orderb{Rate: rate, Quantity: quan}
		orderBook.Sell = append(orderBook.Sell, oneOrder)
	}
	/*if err != nil {
		return nil, fmt.Errorf("HuoBi.GetOrderBook(\"%s\",%d) error %v", pairString, depthSize, err)
	}
	return &orderBook, err
	*/
	return &orderBook, nil
}

/*set "" if coin has no paymentid*/
func (b *HuoBi) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {
	return "000", fmt.Errorf("not impl")
}
func (b *HuoBi) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
	return WITHDRAW_ERROR, fmt.Errorf("not impl")
}
func (b *HuoBi) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
	return nil, fmt.Errorf("not impl")
}
func (b *HuoBi) CheckWalletValid(coin string) (bool, error) {
	return true, nil
	return false, fmt.Errorf("not impl")
}
func (b *HuoBi) GetTradingBalance(currency string) (float64, error) {
	return 0.0, fmt.Errorf("not impl")
}
func (b *HuoBi) GetPaymentBalance(currency string) (float64, error) {
	return b.GetTradingBalance(currency)
}
func (b *HuoBi) TransferToPayment(currency string, amount float64) error {
	return nil
}
func (b *HuoBi) TransferToTrading(currency string, amount float64) error {
	return nil
}
func (b *HuoBi) GetDepositAddress(currency string) (string, string, error) {
	return "", "", fmt.Errorf("not impl")
}

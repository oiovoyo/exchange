package exchange

import (
	"fmt"
	"github.com/qct/crypto_coin_api"
	"github.com/qct/crypto_coin_api/poloniex"
	"net/http"
	"strconv"
	"strings"
	"time"
	"log"
)

type Poloniex struct {
	*poloniex.Poloniex
}

func NewPoloniex(access, secret string) *Poloniex {
	return &Poloniex{
		poloniex.New(http.DefaultClient, access, secret),
	}
}

func (p *Poloniex) MakeLocalPair(pair string) string {
	return pair
}

func (p *Poloniex) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
	pair = p.MakeLocalPair(pair)
	currPair := coinapi.CurrencyPairMap[pair]
	priceString := fmt.Sprintf("%.8f", price)
	amountString := fmt.Sprintf("%.8f", amount)
	var order *coinapi.Order
	switch tradeType {
	case "buy":
		order, err = p.LimitBuy(amountString, priceString, currPair)
	case "sell":
		order, err = p.LimitSell(amountString, priceString, currPair)
	default:
		return 0.0, fmt.Errorf("uknown trade type %s", tradeType)
	}

	if err != nil {
		log.Printf("tradeOneTime error %v", err)
		if err := p.CancelOpenOrders(pair); err != nil {
			log.Printf("CancelOpenOrders error %v", err)
		}
		return 0.0, fmt.Errorf("Poloniex.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
			tradeType, pair, amount, price, err)
	}
	time.Sleep(time.Millisecond * time.Duration(1000))

	//
	for i := 0; i < 10; i++ {
		ok, err := p.CancelOrder(strconv.Itoa(order.OrderID), currPair)
		if err == nil && ok {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(100))
	}
	var getOrder *coinapi.Order
	for i := 0; i < 10; i++ {
		getOrder, err = p.GetOneOrder(strconv.Itoa(order.OrderID), currPair)
		if err != nil {
			log.Printf("GetOneOrder error %v", err)
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		break
	}
	if err != nil {
		// fuck me here
		return 0.0, fmt.Errorf("Poloniex.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
			tradeType, pair, amount, price, err)
	}
	return getOrder.DealAmount, nil
}
func (p *Poloniex) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
	return p.tradeOneTime("buy", pair, buyAmount, price)

}
func (p *Poloniex) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
	return p.tradeOneTime("sell", pair, sellAmount, price)
}
func (p *Poloniex) CancelOpenOrders(pair string) error {
	pair = p.MakeLocalPair(pair)
	currPair := coinapi.CurrencyPairMap[pair]
	orders, err := p.GetUnfinishOrders(currPair)
	if err != nil {
		return fmt.Errorf("Poloniex.CancelOpenOrders(\"%s\") error %v", pair, err)
	}

	for _, o := range orders {
		if _, err := p.CancelOrder(strconv.Itoa(o.OrderID), currPair); err != nil {
			log.Printf("CancelOrder(%d,%s) error : %v", o.OrderID, pair, err)
		}
	}
	return nil
}
func (p *Poloniex) GetOrderBook(pairString string, depthSize int) (*OrderBook, error) {
	pairString = p.MakeLocalPair(pairString)
	pair := coinapi.CurrencyPairMap[pairString]
	ob, err := p.GetDepth(depthSize, pair)
	if err != nil {
		return nil, fmt.Errorf("Poloniex.GetOrderBook(\"%s\",%d) error %v", pairString, depthSize, err)
	}
	orderBook := &OrderBook{Buy: make([]Orderb, 0), Sell: make([]Orderb, 0)}

	for _, v := range ob.AskList {
		var o Orderb
		o.Quantity, o.Rate = v.Amount, v.Price
		orderBook.Sell = append(orderBook.Sell, o)
	}
	for _, v := range ob.BidList {
		var o Orderb
		o.Quantity, o.Rate = v.Amount, v.Price
		orderBook.Buy = append(orderBook.Buy, o)
	}
	return orderBook, nil
}

/*set "" if coin has no paymentid*/
func (p *Poloniex) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {
	curr := coinapi.SymbolCurrency[currency]
	amountString := fmt.Sprintf("%.4f", amount)
	if paymentID != "" {
		_, err = p.WithdrawWithMemo(amountString, curr, paymentID, address, "safepass")
	} else {
		_, err = p.Poloniex.Withdraw(amountString, curr, "0.0000001", address, "safepass")
	}
	if err != nil {
		return "", fmt.Errorf("Poloniex.Withdraw(\"%s\",%.8f,\"%s\",\"%s\") error %v",
			currency, amount, address, paymentID, err)
	}
	return p.getWithdrawLastID(currency, address, time.Now().UTC().Add(-1*time.Minute).Unix())
}
func (p *Poloniex) getWithdrawLastID(currency string, address string, afterTimeUTC int64) (withrawID string, err error) {
	// from 5 days ago
	start := strconv.FormatInt(time.Now().UTC().Add(-120*time.Hour).Unix(), 10)
	//end := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	w, err := p.GetDepositsWithdrawals(start, "")
	if err != nil {
		return "", fmt.Errorf("Poloniex.getWithdrawLastID(\"%s\",\"%s\",%d) error %v",
			currency, address, afterTimeUTC, err)
	}

	for _, o := range w.Withdrawals {

		if o.Currency == currency && o.Address == address {
			if int64(o.Timestamp) >= time.Unix(afterTimeUTC, 0).UTC().Unix() {
				log.Printf("found one %+v", o)
				return strconv.FormatInt(o.WithdrawalNumber, 10), nil
			}

		}
	}
	return "", fmt.Errorf("Poloniex.getWithdrawLastID(\"%s\",\"%s\",%d) error not found",
		currency, address, afterTimeUTC)
}
func (p *Poloniex) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
	start := strconv.FormatInt(time.Now().UTC().Add(-120*time.Hour).Unix(), 10)
	//end := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	w, err := p.GetDepositsWithdrawals(start, "")
	if err != nil {
		return WITHDRAW_ERROR, fmt.Errorf("Poloniex.GetWithdrawStatus(\"%s\") error %v", withdrawID, err)
	}

	for _, o := range w.Withdrawals {
		if strconv.FormatInt(o.WithdrawalNumber, 10) == withdrawID {
			if strings.HasPrefix(o.Status, "COMPLETE") {
				return WITHDRAW_COMPLETE, nil
			} else {
				return WITHDRAW_PENDING, nil
			}
		}
	}

	return WITHDRAW_ERROR, fmt.Errorf("Poloniex.GetWithdrawStatus(\"%s\") not found", withdrawID)
}
func (p *Poloniex) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
	//end := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	w, err := p.GetDepositsWithdrawals(strconv.FormatInt(utcTime.Unix(), 10), "")
	if err != nil {
		return nil, fmt.Errorf("Poloniex.GetDepositList(\"%s\",%d) error %v", coin, utcTime, err)
	}

	ret := make([]DepositItem, 0)
	for _, o := range w.Deposits {
		if strings.ToUpper(o.Currency) == strings.ToUpper(coin) {
			if int64(o.Timestamp) >= utcTime.Unix() {
				d := DepositItem{coin: o.Currency,
					address: o.Address,
					amount:  o.Amount,
					txid:    o.TransactionID,
					time:    time.Unix(int64(o.Timestamp), 0).UTC()}
				if strings.HasPrefix(o.Status, "COMPLETE") {
					d.Status = DEPOSIT_COMPLETE
				} else {
					d.Status = DEPOSIT_PENDING
				}
				ret = append(ret, d)
			}
		}
	}
	return ret, nil
}
func (p *Poloniex) CheckWalletValid(coin string) (bool, error) {
	curr := coinapi.SymbolCurrency[coin]
	c, err := p.GetCurrency(curr)
	if err != nil {
		return false, fmt.Errorf("Poloniex.CheckWalletValid(\"%s\") error %v", coin, err)
	}
	return c.Disabled != 1, nil
}
func (p *Poloniex) GetTradingBalance(currency string) (float64, error) {
	curr := coinapi.SymbolCurrency[currency]
	a, err := p.GetAccount()
	if err != nil {
		return 0.0, fmt.Errorf("Poloniex.GetBalance(\"%s\") error %v", currency, err)
	}
	if a == nil {
		return 0.0, fmt.Errorf("Poloniex.GetBalance(\"%s\") error account is nil", currency)
	}
	return a.SubAccounts[curr].Amount, nil
}
func (p *Poloniex) GetPaymentBalance(currency string) (float64, error) {
	return p.GetTradingBalance(currency)
}
func (p *Poloniex) TransferToPayment(currency string, amount float64) error {
	return nil
}
func (p *Poloniex) TransferToTrading(currency string, amount float64) error {
	return nil
}
func (p *Poloniex) GetDepositAddress(currency string) (string, error) {
	a, err := p.Poloniex.GetDepositAdresses()
	if err != nil {
		return "", fmt.Errorf("Poloniex.GetDepositAddress(\"%s\") error %v", currency, err)
	}
	if v, ok := a[currency]; ok {
		return v, nil
	}
	return "", fmt.Errorf("Poloniex.GetDepositAddress(\"%s\") error not found", currency)
}

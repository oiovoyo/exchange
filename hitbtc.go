package exchange

import (
	"fmt"
	"github.com/oiovoyo/hitbtc"
	"log"
	"strings"
	"time"
)

type Hitbtc struct {
	*hitbtc.Hitbtc
}

func NewHitbtc(access, secret string) *Hitbtc {
	return &Hitbtc{
		hitbtc.NewHitbtc(access, secret),
	}
}

func (h *Hitbtc) Markets() ([]Market, error) {
	return nil, fmt.Errorf("hitbtc not impl this")
}
func (h *Hitbtc) MakeLocalPair(pair string) string {

	s := strings.Split(pair, "_")
	if len(s) != 2 {
		return "err-pair"
	}
	return s[0] + s[1]

}

func (h *Hitbtc) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
	pair = h.MakeLocalPair(pair)

	var order string
	switch tradeType {
	case "buy":
		order, err = h.BuyLimit(pair, amount, price)
	case "sell":
		order, err = h.SellLimit(pair, amount, price)
	default:
		return 0.0, fmt.Errorf("uknown trade type %s", tradeType)
	}
	log.Printf("tradeOneTime(\"%s\",%.8f,%.8f)", pair, amount, price)
	log.Printf("order:%s error:%v", order, err)
	if err != nil {
		log.Printf("Hitbtc.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
			tradeType, pair, amount, price, err)
		if err := h.CancelOpenOrders(pair); err != nil {
			log.Printf("CancelOpenOrders error %v", err)
		}
		return 0.0, fmt.Errorf("Hitbtc.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
			tradeType, pair, amount, price, err)
	}
	time.Sleep(time.Millisecond * time.Duration(1000))
	var getOrder *hitbtc.QueryOrder
	for i := 0; i < 10; i++ {
		getOrder, err = h.GetOrder(order)
		log.Printf("getOrder:%+v error:%+v", getOrder, err)
		if err != nil {
			log.Printf("Buy error %v", err)
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		if len(getOrder.Orders) > 0 {
			if getOrder.Orders[0].DealAmount/amount > 0.95 {
				break
			}
		}
		time.Sleep(time.Millisecond * time.Duration(1000))
	}
	//
	for i := 0; i < 10; i++ {
		err := h.CancelOrder(order)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(100))
	}
	for i := 0; i < 10; i++ {
		getOrder, err = h.GetOrder(order)
		if err != nil {
			log.Printf("Buy error %v", err)
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		break
	}
	if err != nil {
		// fuck me here
		log.Printf("Buy error %v", err)
		return 0.0, fmt.Errorf("Hitbtc.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
			tradeType, pair, amount, price, err)
	}
	if len(getOrder.Orders) < 1 {
		return 0.0, fmt.Errorf("Hitbtc.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) get order length error",
			tradeType, pair, amount, price)
	}
	log.Printf("getOrder:%+v", getOrder)
	return getOrder.Orders[0].DealAmount, nil
}
func (h *Hitbtc) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
	return h.tradeOneTime("buy", pair, buyAmount, price)

}
func (h *Hitbtc) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
	return h.tradeOneTime("sell", pair, sellAmount, price)
}
func (h *Hitbtc) CancelOpenOrders(pair string) error {
	pair = h.MakeLocalPair(pair)
	return h.Hitbtc.CancelOrders(pair)
}
func (h *Hitbtc) GetOrderBook(pair string, depthSize int) (*OrderBook, error) {
	pair = h.MakeLocalPair(pair)
	return h.Hitbtc.GetOrderBook(pair)
}

func (h *Hitbtc) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {
	return h.Hitbtc.Withdraw(currency, amount, address, paymentID)
}
func (h *Hitbtc) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
	t, err := h.Hitbtc.GetTransaction(withdrawID)
	if err != nil {
		return WITHDRAW_ERROR, err
	}
	if t.IsDone {
		return WITHDRAW_COMPLETE, nil
	}
	return WITHDRAW_PENDING, nil
}
func (h *Hitbtc) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
	offset, limit := 0, 100
	ret := make([]DepositItem, 0)
	for {
		dw, count, err := h.Hitbtc.GetDepositWithdrawal(coin, offset, limit)
		if err != nil {
			break
		}
		if count == 0 {
			break
		}
		beforeLen := len(ret)
		for _, d := range dw.Deposit {
			if d.Created.After(utcTime) {
				ret = append(ret,
					DepositItem{
						amount:  d.Amount,
						address: d.Address,
						time:    d.Created,
						id:      d.Id,
						coin:    d.Coin,
						Status: func() DepositStatus {
							if d.IsDone {
								return DEPOSIT_COMPLETE
							}
							return DEPOSIT_PENDING
						}(),
					},
				)
			}
		}
		if len(ret) == beforeLen {
			break
		}
		offset += count
	}
	return ret, nil
}
func (h *Hitbtc) CheckWalletValid(coin string) (bool, error) {
	return true, nil
	return h.Hitbtc.WalletValid(coin), nil
}
func (h *Hitbtc) GetTradingBalance(currency string) (float64, error) {
	b, err := h.Hitbtc.GetTradeBalance()

	if err != nil {
		return 0.0, err
	}
	for _, o := range b.Balance {
		if o.Currency == currency {
			return o.Cash, nil
		}
	}
	return 0.0, nil
}
func (h *Hitbtc) GetPaymentBalance(currency string) (float64, error) {
	b, err := h.Hitbtc.GetPaymentBalance()

	if err != nil {
		return 0.0, err
	}
	for _, o := range b.Balance {
		if o.Currency == currency {
			return o.Balance, nil
		}
	}
	return 0.0, nil
}
func (h *Hitbtc) GetDepositAddress(currency string) (string, string, error) {
	a, err := h.Hitbtc.GetDepositAddress(currency)
	return a, "", err
}

func (h *Hitbtc) TransferToPayment(currency string, amount float64) error {
	return h.Hitbtc.TransferToMain(currency, amount)
}
func (h *Hitbtc) TransferToTrading(currency string, amount float64) error {
	return h.Hitbtc.TransferToTrading(currency, amount)
}

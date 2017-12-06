package exchange

import (
	"github.com/oiovoyo/go-bittrex"
	"time"
)

type WithdrawStatus int
type DepositStatus int
type DepositItem struct {
	amount  float64
	coin    string
	address string
	txid    string
	time    time.Time
	Status  DepositStatus
	id      string
}

const (
	WITHDRAW_ERROR WithdrawStatus = iota
	WITHDRAW_PENDING
	WITHDRAW_COMPLETE
)
const (
	DEPOSIT_ERROR DepositStatus = iota
	DEPOSIT_PENDING
	DEPOSIT_COMPLETE
)

type Orderb = bittrex.Orderb
type OrderBook = bittrex.OrderBook

type IExchange interface {
	/*pair format is countercoin_basecoin example XMR_BTC same as followed api*/
	BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error)
	SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error)
	CancelOpenOrders(pair string) error
	GetOrderBook(pair string, dethSize int) (*OrderBook, error)
	/*set "" if coin has no paymentid*/
	Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error)
	GetWithdrawStatus(withdrawID string) (WithdrawStatus, error)
	GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error)
	CheckWalletValid(coin string) (bool, error)
	GetPaymentBalance(currency string) (float64, error)
	GetTradingBalance(currency string) (float64, error)
	TransferToPayment(currency string, amount float64) error
	TransferToTrading(currency string, amount float64) error
	GetDepositAddress(currency string) (string, error)
}

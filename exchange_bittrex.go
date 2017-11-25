package exchange

import (
    "fmt"
    "github.com/toorop/go-bittrex"
    "strconv"
    "strings"
    "time"
    "log"
    "net/http"
)

type Bittrex struct {
    *bittrex.Bittrex
}

func NewBittrex(access, secret string) *Bittrex {
    return &Bittrex{
        bittrex.New(access, secret),
    }
}

func NewCustomBittrex(access, secret string, client *http.Client) *Bittrex {
    return &Bittrex{
        bittrex.NewWithCustomHttpClient(access, secret, client),
    }
}
func (b *Bittrex) MakeLocalPair(pair string) string {

    s := strings.Split(pair, "_")
    if len(s) != 2 {
        return "err-pair"
    }
    return s[1] + "-" + s[0]

}

func (b *Bittrex) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
    pair = b.MakeLocalPair(pair)

    var order string
    switch tradeType {
    case "buy":
        order, err = b.BuyLimit(pair, amount, price)
    case "sell":
        order, err = b.SellLimit(pair, amount, price)
    default:
        return 0.0, fmt.Errorf("uknown trade type %s", tradeType)
    }

    if err != nil {
        log.Printf("Buy error %v", err)
        if err := b.CancelOpenOrders(pair); err != nil {
            log.Printf("CancelOpenOrders error %v", err)
        }
        return 0.0, fmt.Errorf("Bittrex.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
            tradeType, pair, amount, price, err)
    }
    time.Sleep(time.Millisecond * time.Duration(1000))

    //
    for i := 0; i < 10; i++ {
        err := b.CancelOrder(order)
        if err == nil {
            break
        }
        time.Sleep(time.Millisecond * time.Duration(100))
    }
    var getOrder bittrex.Order2
    for i := 0; i < 10; i++ {
        getOrder, err = b.GetOrder(order)
        if err != nil {
            log.Printf("trade error %v", err)
            time.Sleep(time.Millisecond * time.Duration(100))
            continue
        }
        break
    }
    if err != nil {
        // fuck me here
        log.Printf("trade error %v", err)
        return 0.0, fmt.Errorf("Bittrex.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
            tradeType, pair, amount, price, err)
    }
    return getOrder.Quantity - getOrder.QuantityRemaining, nil
}

func (b *Bittrex) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
    return b.tradeOneTime("buy", pair, buyAmount, price)
}
func (b *Bittrex) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
    return b.tradeOneTime("sell", pair, sellAmount, price)
}
func (b *Bittrex) CancelOpenOrders(pair string) error {
    pair = b.MakeLocalPair(pair)
    orders, err := b.GetOpenOrders(pair)
    if err != nil {
        return fmt.Errorf("Bittrex.CancelOpenOrders(\"%s\") error %v", pair, err)
    }
    for _, o := range orders {
        if err := b.CancelOrder(o.OrderUuid); err != nil {
            log.Printf("CancelOrder(%s,%s) error : %v", o.OrderUuid, pair, err)
        }
    }
    return nil
}
func (b *Bittrex) GetOrderBook(pairString string, depthSize int) (*OrderBook, error) {
    pairString = b.MakeLocalPair(pairString)
    orderBook, err := b.Bittrex.GetOrderBook(pairString, "both", depthSize)
    if err != nil {
        return nil, fmt.Errorf("Bittrex.GetOrderBook(\"%s\",%d) error %v", pairString, depthSize, err)
    }
    return &orderBook, err

}

/*set "" if coin has no paymentid*/
func (b *Bittrex) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {
    /*if paymentID != "" {
        return "", fmt.Errorf("Bittrex.Withdraw(\"%s\",%.8f,\"%s\",\"%s\") error not support paymentid",
            currency, amount, address, paymentID)

    }*/
    if paymentID != "" {
        withrawID, err = b.Bittrex.WithdrawPaymentID(address, paymentID, currency, amount)

    } else {
        withrawID, err = b.Bittrex.Withdraw(address, currency, amount)
    }

    if err != nil {
        return "", fmt.Errorf("Bittrex.Withdraw(\"%s\",%.8f,\"%s\",\"%s\") error %v",
            currency, amount, address, paymentID, err)
    }
    return withrawID, nil
}
func (b *Bittrex) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
    w, err := b.GetWithdrawalHistory("all")
    if err != nil {
        return WITHDRAW_ERROR, fmt.Errorf("Bittrex.GetWithdrawStatus(\"%s\") error %v", withdrawID, err)
    }
    for _, o := range w {
        if o.PaymentUuid == withdrawID {
            if o.PendingPayment == false && o.Authorized == true {
                return WITHDRAW_COMPLETE, nil
            } else {
                return WITHDRAW_PENDING, nil
            }
        }
    }
    return WITHDRAW_PENDING, nil
    //return WITHDRAW_ERROR, fmt.Errorf("Bittrex.GetWithdrawStatus(\"%s\") error not found", withdrawID)
}
func (b *Bittrex) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
    d, err := b.GetDepositHistory(coin)
    if err != nil {
        return nil, fmt.Errorf("Bittrex.GetDepositList(\"%s\",%d) error %v", coin, utcTime, err)
    }
    ret := make([]DepositItem, 0)
    for _, o := range d {
        if o.Currency == coin {
            if o.LastUpdated.Unix() >= utcTime.Unix() {
                d := DepositItem{coin: o.Currency,
                    address: o.CryptoAddress,
                    amount: o.Amount,
                    txid: o.TxId,
                    id: strconv.FormatInt(o.Id, 10),
                    time: o.LastUpdated.Time,
                    Status: DEPOSIT_COMPLETE, // bittrex can only query complete deposit
                }
                ret = append(ret, d)
            }
        }
    }
    return ret, nil
}
func (b *Bittrex) CheckWalletValid(coin string) (bool, error) {
    //return true, nil
    c, err := b.GetCurrencies()
    if err != nil {
        return false, fmt.Errorf("Bittrex.CheckWalletValid(\"%s\") error %v", coin, err)
    }
    for _, v := range c {
        if strings.ToUpper(v.Currency) == strings.ToUpper(coin) {
            return v.IsActive, nil
        }
    }
    return false, fmt.Errorf("Bittrex.CheckWalletValid(\"%s\") error not found", coin)
}
func (b *Bittrex) GetTradingBalance(currency string) (float64, error) {
    a, err := b.Bittrex.GetBalance(currency)
    if err != nil {
        return 0.0, fmt.Errorf("Bittrex.GetBalance(\"%s\") error %v", currency, err)
    }
    return a.Available, nil
}
func (b *Bittrex) GetPaymentBalance(currency string) (float64, error) {
    return b.GetTradingBalance(currency)
}
func (b *Bittrex) TransferToPayment(currency string, amount float64) error {
    return nil
}
func (b *Bittrex) TransferToTrading(currency string, amount float64) error {
    return nil
}
func (b *Bittrex) GetDepositAddress(currency string) (string, error) {
    a, err := b.Bittrex.GetDepositAddress(currency)
    if err != nil {
        return "", fmt.Errorf("Bittrex.GetDepositAddress(\"%s\") error not found", currency)
    }
    return a.Address, nil
}

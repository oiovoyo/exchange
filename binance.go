package exchange

import (
	"context"
	"fmt"
	"github.com/oiovoyo/go-binance"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TickSizeMinTrade struct {
	TickSize   float64
	MinTrade   float64
	BaseAsset  string
	QuoteAsset string
}

var (
	ProductMap          map[string]binance.Product
	BinanceExchangeInfo sync.Map
)

func BinanceUpdateExchangeInfo() error {

	exchangeinfo, err := binance.NewClient("", "").NewExchangeInfoService().Do(context.Background())

	if err != nil {
		return err
	}

	for _, ex := range exchangeinfo.Symbols {

		ex_min_trade := 0.0
		ex_tick_size := 0.0
		for _, fi := range ex.Filters {
			if fi["filterType"] == "LOT_SIZE" {
				ex_min_trade, _ = strconv.ParseFloat(fi["minQty"], 64)
			} else if fi["filterType"] == "PRICE_FILTER" {
				ex_tick_size, _ = strconv.ParseFloat(fi["tickSize"], 64)
			}
		}
		if ex_min_trade == 0.0 {
			return fmt.Errorf("%s min trade not found", ex.Symbol)
		}
		if ex_tick_size == 0.0 {
			return fmt.Errorf("%s tick size not found", ex.Symbol)
		}
		BinanceExchangeInfo.Store(ex.Symbol,
			TickSizeMinTrade{
				TickSize:   ex_tick_size,
				MinTrade:   ex_min_trade,
				BaseAsset:  ex.BaseAsset,
				QuoteAsset: ex.QuoteAsset,
			},
		)
	}
	return nil
}
func binanceGetTickSize(pair string) (float64, error) {
	t_m, ok := BinanceExchangeInfo.Load(pair)
	if !ok {
		return 0.0, fmt.Errorf("binanceGetTickSize(\"%s\") load false\n", pair)
	}

	return t_m.(TickSizeMinTrade).TickSize, nil
}
func binanceGetMinTrade(pair string) (float64, error) {
	t_m, ok := BinanceExchangeInfo.Load(pair)
	if !ok {
		return 0.0, fmt.Errorf("binanceGetMinTrade(\"%s\") load false \n", pair)
	}

	return t_m.(TickSizeMinTrade).MinTrade, nil
}
func init() {
	/*
		p, err := binance.NewClient("", "").NewListProductService().Do(context.Background())

		if err != nil {
			fmt.Printf("exchange_binance init error %v", err)
			os.Exit(1)
		}

		ProductMap = make(map[string]binance.Product)

		for _, v := range p.Data {
			ProductMap[v.Symbol] = v
		}
	*/
	err := BinanceUpdateExchangeInfo()
	if err != nil {
		fmt.Printf("exchange_binance init error %v", err)
		os.Exit(1)
	}
}

type PriceTrunc int

const (
	PRICE_UP = PriceTrunc(iota)
	PRICE_DOWN
)

func TruncatePrice(pair string, price float64, priceTrunc PriceTrunc) string {
	tickSize, err := binanceGetTickSize(pair) //ProductMap[pair].TickSize
	if err != nil {
		fmt.Printf("TruncatePrice(\"%s\") err %v\n", pair, err)
		os.Exit(1)
	}
	var retPrice float64
	if price < tickSize {
		retPrice = tickSize
	} else {
		count := int64(price / tickSize)

		if float64(count)*tickSize == price {
			retPrice = price
		} else {
			switch priceTrunc {
			case PRICE_UP:
				retPrice = float64(count+1) * tickSize
			case PRICE_DOWN:
				retPrice = float64(count) * tickSize
			}
		}

	}
	fmtCount := 0
	for tickSize < 1.0 {
		fmtCount++
		tickSize *= 10.0
	}
	fmtStr := fmt.Sprintf("%%.%df", fmtCount)
	return fmt.Sprintf(fmtStr, retPrice)
}
func TruncateAmount(pair string, amount float64) string {
	//minAmount := ProductMap[pair].MinTrade
	minAmount, err := binanceGetMinTrade(pair) //ProductMap[pair].MinTrade
	//fmt.Printf("min %.8f amount %.8f\n",minAmount, amount)
	if err != nil {
		fmt.Printf("TruncateAmount(\"%s\") err %v\n", pair, err)
		os.Exit(1)
	}
	var retAmount float64
	if amount < minAmount {
		retAmount = 0.0
	} else {
		count := int64(amount / minAmount)
		retAmount = float64(count) * minAmount
	}
	fmtCount := 0
	for minAmount < 1.0 {
		fmtCount++
		minAmount *= 10.0
	}
	fmtStr := fmt.Sprintf("%%.%df", fmtCount)
	return fmt.Sprintf(fmtStr, retAmount)

}

type Binance struct {
	*binance.Client
}

func NewBinance(access, secret string) *Binance {
	return &Binance{
		binance.NewClient(access, secret),
	}
}
func NewCustomBinance(access, secret string, client *http.Client) *Binance {
	return &Binance{
		binance.NewClientCustomHttp(access, secret, client),
	}
}
func (b *Binance) Markets() ([]Market, error) {
	m := make([]Market, 0)
	BinanceExchangeInfo.Range(func(key, value interface{}) bool {
		v := value.(TickSizeMinTrade)
		m = append(m, Market{Base: v.BaseAsset, Quot: v.QuoteAsset})
		return true
	})
	return m, nil
}
func (b *Binance) MakeLocalPair(pair string) string {

	s := strings.Split(pair, "_")
	if len(s) != 2 {
		return "err-pair"
	}
	return s[0] + s[1]

}

func (b *Binance) tradeOneTime(tradeType, pair string, amount, price float64) (dealAmount float64, err error) {
	pair = b.MakeLocalPair(pair)

	amountStr := TruncateAmount(pair, amount)

	var order *binance.CreateOrderResponse
	switch tradeType {
	case "buy":
		priceStr := TruncatePrice(pair, price, PRICE_UP)
		fmt.Printf("amount %s price %s\n", amountStr, priceStr)
		order, err = b.NewCreateOrderService().Symbol(pair).
			Price(priceStr).
			Quantity(amountStr).
			Side(binance.SideTypeBuy).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceGTC).Do(context.Background())
	case "sell":
		priceStr := TruncatePrice(pair, price, PRICE_DOWN)
		order, err = b.NewCreateOrderService().Symbol(pair).
			Price(priceStr).
			Quantity(amountStr).
			Side(binance.SideTypeSell).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceGTC).Do(context.Background())
	default:
		return 0.0, fmt.Errorf("uknown trade type %s", tradeType)
	}

	time.Sleep(time.Millisecond * time.Duration(1000))

	//
	for i := 0; i < 10; i++ {
		orders, err := b.NewListOpenOrdersService().Symbol(pair).Do(context.Background())
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		for _, o := range orders {
			b.NewCancelOrderService().Symbol(pair).OrigClientOrderID(o.ClientOrderID).Do(context.Background())
		}
		break
	}

	if err != nil {
		log.Printf("%s error %v", tradeType, err)
		return 0.0, fmt.Errorf("Binance.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
			tradeType, pair, amount, price, err)
	}

	var getOrder *binance.Order
	for i := 0; i < 10; i++ {
		getOrder, err = b.NewGetOrderService().
			Symbol(pair).
			OrigClientOrderID(order.ClientOrderID).
			Do(context.Background())
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
		return 0.0, fmt.Errorf("Binance.tradeOneTime(\"%s\",\"%s\",%.8f,%.8f) error %v",
			tradeType, pair, amount, price, err)
	}
	return strconv.ParseFloat(getOrder.ExecutedQuantity, 64)
}

func (b *Binance) BuyOneTime(pair string, buyAmount, price float64) (dealAmount float64, err error) {
	return b.tradeOneTime("buy", pair, buyAmount, price)
}
func (b *Binance) SellOneTime(pair string, sellAmount, price float64) (dealAmount float64, err error) {
	return b.tradeOneTime("sell", pair, sellAmount, price)
}

func (b *Binance) CancelOpenOrders(pair string) error {
	pair = b.MakeLocalPair(pair)
	orders, err := b.NewListOpenOrdersService().Symbol(pair).Do(context.Background())
	if err != nil {
		return err
	}
	for _, o := range orders {
		b.NewCancelOrderService().Symbol(pair).OrigClientOrderID(o.ClientOrderID).Do(context.Background())
	}
	return nil
}

func (b *Binance) GetOrderBook(pairString string, depthSize int) (*OrderBook, error) {
	pairString = b.MakeLocalPair(pairString)
	orderBookBinance, err := b.NewDepthService().
		Symbol(pairString).
		Limit(depthSize).
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Binance.GetOrderBook(\"%s\",%d) error %v", pairString, depthSize, err)
	}
	var orderBook OrderBook
	orderBook.Sell = make([]Orderb, 0)
	orderBook.Buy = make([]Orderb, 0)

	for _, v := range orderBookBinance.Asks {
		q, _ := strconv.ParseFloat(v.Quantity, 64)
		r, _ := strconv.ParseFloat(v.Price, 64)
		order := Orderb{Quantity: q, Rate: r}
		orderBook.Sell = append(orderBook.Sell, order)
	}
	for _, v := range orderBookBinance.Bids {
		q, _ := strconv.ParseFloat(v.Quantity, 64)
		r, _ := strconv.ParseFloat(v.Price, 64)
		order := Orderb{Quantity: q, Rate: r}
		orderBook.Buy = append(orderBook.Buy, order)
	}
	return &orderBook, err

}

func (b *Binance) Withdraw(currency string, amount float64, address, paymentID string) (withrawID string, err error) {

	amountStr := ""
	if currency == "BTC" {
		amountStr = fmt.Sprintf("%.8f", amount)
	} else {
		pair := b.MakeLocalPair(currency + "_BTC")
		amountStr = TruncateAmount(pair, amount)
	}

	var withdrawService = b.NewCreateWithdrawService().
		Asset(currency).
		Name("bc").
		Address(address).
		Amount(amountStr)

	if paymentID != "" {
		withdrawService.AddressTag(paymentID)
	}
	id, err := withdrawService.Do(context.Background())

	if err != nil {
		return "", fmt.Errorf("Binance.Withdraw(\"%s\",%.8f,\"%s\",\"%s\") error %v",
			currency, amount, address, paymentID, err)
	}

	return id, nil
}

func (b *Binance) GetWithdrawStatus(withdrawID string) (WithdrawStatus, error) {
	w, err := b.NewListWithdrawsService().
		StartTime(time.Now().AddDate(0, 0, -10).UTC().UnixNano() / int64(time.Millisecond)).
		Do(context.Background())
	if err != nil {
		return WITHDRAW_ERROR, fmt.Errorf("Binance.GetWithdrawStatus(\"%s\") error %v", withdrawID, err)
	}
	for _, o := range w {
		id := o.Address + fmt.Sprintf("_%v", o.ApplyTime)
		//fmt.Printf("%+v\n",o)
		if o.Id == withdrawID {
			if o.TxID != "" || o.Status == 6 {
				return WITHDRAW_COMPLETE, nil
			} else {
				return WITHDRAW_PENDING, nil
			}
		} else if id == withdrawID {
			if o.TxID != "" || o.Status == 6 {
				return WITHDRAW_COMPLETE, nil
			} else {
				return WITHDRAW_PENDING, nil
			}
		}
	}
	return WITHDRAW_ERROR, fmt.Errorf("Binance.GetWithdrawStatus(\"%s\") error not found", withdrawID)
}

func (b *Binance) GetDepositList(coin string, utcTime time.Time) ([]DepositItem, error) {
	d, err := b.NewListDepositsService().Asset(coin).
		StartTime(utcTime.UnixNano() / int64(time.Millisecond)).
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Binance.GetDepositList(\"%s\",%d) error %v", coin, utcTime, err)
	}
	ret := make([]DepositItem, 0)
	fmt.Println(d)
	for _, o := range d {
		if o.Asset == coin {
			depositStatus := DEPOSIT_PENDING
			if o.Status == 1 {
				depositStatus = DEPOSIT_COMPLETE
			}
			if o.InsertTime >= utcTime.UnixNano()/int64(time.Millisecond) {
				d := DepositItem{coin: o.Asset,
					address: "",
					amount:  o.Amount,
					txid:    "",
					id:      "0",
					time:    time.Unix(utcTime.UnixNano()/int64(time.Millisecond), 0),
					Status:  depositStatus, // Binance can only query complete deposit
				}
				ret = append(ret, d)
			}
		}
	}
	return ret, nil
}

func (b *Binance) CheckWalletValid(coin string) (bool, error) {
	// avoid ban
	return true, nil
	c, err := b.NewGetAccountService().Do(context.Background())
	if err != nil {
		return false, fmt.Errorf("Binance.CheckWalletValid(\"%s\") error %v", coin, err)
	}
	if c.CanDeposit && c.CanWithdraw {
		return true, nil
	}
	return false, fmt.Errorf("Binance.CheckWalletValid(\"%s\") error invalid", coin)
}

func (b *Binance) GetTradingBalance(currency string) (float64, error) {
	c, err := b.NewGetAccountService().Do(context.Background())
	if err != nil {
		return 0.0, fmt.Errorf("Binance.GetBalance(\"%s\") error %v", currency, err)
	}
	for _, v := range c.Balances {
		if v.Asset == currency {
			f, _ := strconv.ParseFloat(v.Free, 64)
			if currency != "BTC" {
				pair := b.MakeLocalPair(currency + "_BTC")
				amountStr := TruncateAmount(pair, f)
				f, _ = strconv.ParseFloat(amountStr, 64)
			}

			return f, nil
		}
	}
	return 0.0, nil
}
func (b *Binance) GetPaymentBalance(currency string) (float64, error) {
	return b.GetTradingBalance(currency)
}
func (b *Binance) TransferToPayment(currency string, amount float64) error {
	return nil
}
func (b *Binance) TransferToTrading(currency string, amount float64) error {
	return nil
}

func (b *Binance) GetDepositAddress(currency string) (string, string, error) {
	c, err := b.NewDepositAddressService().Asset(currency).Do(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("Binance.GetBalance(\"%s\") error %v", currency, err)
	}
	return c.Address, c.AddressTag, nil
	//return "",errors.New("binance not supported")
}

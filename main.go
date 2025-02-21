package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/asadhayat1068/go_exchange/orderbook"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

const (
	MarketETH          Market    = "ETH"
	MarketOrder        OrderType = "MARKET"
	LimitOrder         OrderType = "LIMIT"
	exchangePrivateKey           = "2ec39631df9d2fb4969ab7e31fde7378e0db1802147838e35bedbcb8b79aeae9"
)

type (
	OrderType string
	Market    string

	PlaceOrderRequest struct {
		Type   OrderType // Limit or Market
		UserID int64
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}
	Order struct {
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}
	OrderbookData struct {
		AskTotalVolume float64
		BidTotalVolume float64
		Asks           []*Order
		Bids           []*Order
	}
)

func main() {
	e := echo.New()

	e.HTTPErrorHandler = httpErrorHandler
	url := "http://localhost:8545"
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Fatal(err)
	}

	ex := NewExchange(exchangePrivateKey, client)
	e.GET("/books/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.handleCancelOrder)

	// ctx := context.Background()
	// address := common.HexToAddress("0xc9E8B0d061e610A02882F67Cb5daFCfd61Bb7253")

	// balance, err := client.BalanceAt(ctx, address, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("Balance: ", balance)

	fmt.Println("Working!!")
	e.Start(":3000")
}

type User struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(privateKey string) *User {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		PrivateKey: pk,
	}
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

type Exchange struct {
	Client     *ethclient.Client
	users      map[int64]*User
	orders     map[int64]int64
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange(privateKey string, client *ethclient.Client) *Exchange {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	pk, err := crypto.HexToECDSA(exchangePrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	address, err := getAddress(pk)
	if err != nil {
		log.Fatal(err)
	}

	return &Exchange{
		Client:     client,
		users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
		PrivateKey: pk,
		Address:    address,
		orderbooks: orderbooks,
	}
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))

	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "Market not Found"})
	}

	orderbookData := OrderbookData{
		AskTotalVolume: ob.AskTotalVolume(),
		BidTotalVolume: ob.BidTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)

		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)

		}
	}

	return c.JSON(http.StatusOK, orderbookData)
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*Order) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchedOrders := make([]*Order, len(matches))
	for i := 0; i < len(matchedOrders); i++ {
		match := matches[i]
		id := match.Bid.ID
		if order.Bid {
			id = match.Ask.ID
		}
		matchedOrders[i] = &Order{
			Size:  match.SizeFilled,
			Price: match.Price,
			ID:    id,
		}
	}
	return matches, matchedOrders
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	user := ex.users[order.UserID]

	//TODO: Work on this conversion from ETH to Wei
	amount := big.NewInt(int64(order.Size))
	return transferETH(ex.Client, user.PrivateKey, ex.Address, amount)
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)

	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	if placeOrderData.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}
		return c.JSON(200, map[string]any{"msg": "Limit Order Placed!"})
	}
	if placeOrderData.Type == MarketOrder {
		matches, matchedOrders := ex.handlePlaceMarketOrder(market, order)
		if err := ex.handleMatches(matches); err != nil {
			return err
		}
		return c.JSON(200, map[string]any{"matches": matchedOrders})
	}
	return nil
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	ob := ex.orderbooks[MarketETH]
	order := ob.Orders[int64(id)]
	ob.CancelOrder(order)

	return c.JSON(200, map[string]any{"msg": "Order Deleted!"})
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	// for _, match := range matches {
	// 	// transfer from
	// }
	return nil
}

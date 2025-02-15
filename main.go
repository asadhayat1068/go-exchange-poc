package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
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
	ex := NewExchange(exchangePrivateKey)
	e.GET("/books/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.handleCancelOrder)

	url := "http://localhost:8545"
	client, err := ethclient.Dial(url)

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	address := common.HexToAddress("0xc9E8B0d061e610A02882F67Cb5daFCfd61Bb7253")

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Balance: ", balance)

	fmt.Println("Working!!")
	e.Start(":3000")
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

type Exchange struct {
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange(privateKey string) *Exchange {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	pk, err := crypto.HexToECDSA(exchangePrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	return &Exchange{
		PrivateKey: pk,
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

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	ob := ex.orderbooks[market]
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size)

	if placeOrderData.Type == LimitOrder {
		ob.PlaceLimitOrder(placeOrderData.Price, order)
		return c.JSON(200, map[string]any{"msg": "Limit Order Placed!"})
	} else if placeOrderData.Type == MarketOrder {
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

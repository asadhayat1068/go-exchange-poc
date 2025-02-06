package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asadhayat1068/go_exchange/orderbook"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	ex := NewExchange()
	e.GET("/books/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	// e.DELETE("/order/:id", ex.handleCancelOrder)

	e.Start(":3000")

	fmt.Println("Working!!")
}

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

type Market string

const (
	MarketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type Order struct {
	ID        int64
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type OrderbookData struct {
	AskTotalVolume float64
	BidTotalVolume float64
	Asks           []*Order
	Bids           []*Order
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

type PlaceOrderRequest struct {
	Type   OrderType // Limit or Market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
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
		return c.JSON(200, map[string]any{"matches": len(matches)})
	}
	return nil
}

// func (ex *Exchange) handleCancelOrder(c echo.Context) error {
// 	id := c.Param("id")
// 	return nil
// }

package orderbook

import (
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}
func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5, 0)
	buyOrderB := NewOrder(true, 10, 0)
	buyOrderC := NewOrder(true, 15, 0)
	buyOrderD := NewOrder(true, 20, 0)
	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)
	l.AddOrder(buyOrderD)
	// fmt.Println(l)
	// l.DeleteOrder(buyOrderB)
	// fmt.Println(l)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()
	sellOrderA := NewOrder(false, 10, 0)
	sellOrderB := NewOrder(false, 5, 0)
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(9_000, sellOrderB)

	assert(t, ob.Orders[sellOrderA.ID], sellOrderA)
	assert(t, ob.Orders[sellOrderB.ID], sellOrderB)
	assert(t, len(ob.Orders), 2)
	assert(t, len(ob.Asks()), 2)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20, 0)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(ob.Asks()), 1)
	assert(t, len(matches), 1)
	assert(t, ob.AskTotalVolume(), 10.0)

	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)

	assert(t, buyOrder.IsFilled(), true)

}

func TestPlaceMarketOrderMultiMatch(t *testing.T) {
	ob := NewOrderbook()

	buyOrderC := NewOrder(true, 10, 0)
	buyOrderD := NewOrder(true, 1, 0)
	buyOrderB := NewOrder(true, 8, 0)
	buyOrderA := NewOrder(true, 5, 0)

	ob.PlaceLimitOrder(5_000, buyOrderD)
	ob.PlaceLimitOrder(5_000, buyOrderA)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(10_000, buyOrderC)

	sellOrder := NewOrder(false, 20, 0)
	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, ob.BidTotalVolume(), 4.0)
	assert(t, len(matches), 4)
	assert(t, len(ob.Bids()), 1)
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()

	buyOrder := NewOrder(true, 4, 0)
	ob.PlaceLimitOrder(10_000, buyOrder)

	assert(t, len(ob.Bids()), 1)
	assert(t, ob.BidTotalVolume(), 4.0)

	ob.CancelOrder(buyOrder)

	assert(t, len(ob.Bids()), 0)
	assert(t, ob.BidTotalVolume(), 0.0)
	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)
}

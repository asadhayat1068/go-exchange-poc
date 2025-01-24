package main

import (
	"fmt"
	"sort"
	"time"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

type Orders []*Order

func (o Orders) Len() int {
	return len(o)
}
func (o Orders) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
func (o Orders) Less(i, j int) bool {
	return o[i].Timestamp < o[j].Timestamp
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("Order{\n\tSize: %2f,\n\t Bid: %t,\n\t Timestamp: %d\n}\n", o.Size, o.Bid, o.Timestamp)
}

type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

type Limits []*Limit

type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int {
	return len(a.Limits)
}
func (a ByBestAsk) Swap(i, j int) {
	a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i]
}
func (a ByBestAsk) Less(i, j int) bool {
	return a.Limits[i].Price < a.Limits[j].Price
}

type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int {
	return len(b.Limits)
}
func (b ByBestBid) Swap(i, j int) {
	b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i]
}
func (b ByBestBid) Less(i, j int) bool {
	return b.Limits[i].Price > b.Limits[j].Price
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

func (l *Limit) AddOrder(order *Order) {
	order.Limit = l
	l.Orders = append(l.Orders, order)
	l.TotalVolume += order.Size
}

func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
		}
	}
	o.Limit = nil
	l.TotalVolume -= o.Size
	// TODO: Resort the whole resting  orders
	sort.Sort(l.Orders)
}

func (l *Limit) String() string {
	return fmt.Sprintf("Limit{\n\tPrice: %2f, \n\tOrders: %v, \n\tTotalVolume: %2f\n}\n", l.Price, l.Orders, l.TotalVolume)
}

type Orderbook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func (ob *Orderbook) String() string {
	return fmt.Sprintf("ORDERBOOK::\n\n__ASKS__: %s\n__BIDS__: %s\n\n__ASK_LIMITS__: %v\n__BID_LIMITS__: %v\n", ob.Asks, ob.Bids, ob.AskLimits, ob.BidLimits)
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		Asks:      []*Limit{},
		Bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

func (ob *Orderbook) PlaceOrder(price float64, o *Order) []Match {
	//1. try to match the orders
	//matching logic
	//2. add the rest of the orders to the book
	if o.Size > 0.0 {
		ob.add(price, o)
	}

	return []Match{}

}

func (ob *Orderbook) add(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		limit.AddOrder(o)
		if o.Bid {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit
		}
	} else {
		limit.AddOrder(o)
	}

}

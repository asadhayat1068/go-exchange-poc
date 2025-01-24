package main

import (
	"fmt"
	"testing"
)

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 10)
	buyOrderC := NewOrder(true, 15)
	buyOrderD := NewOrder(true, 20)
	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)
	l.AddOrder(buyOrderD)
	// fmt.Println(l)
	// l.DeleteOrder(buyOrderB)
	// fmt.Println(l)
}

func TestOrderbook(t *testing.T) {
	ob := NewOrderbook()
	orderA := NewOrder(true, 12)
	orderB := NewOrder(true, 3)
	ob.PlaceOrder(13000, orderA)
	ob.PlaceOrder(13000, orderB)

	fmt.Println(ob)
}

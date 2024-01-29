package models

type PlaceOrderRequest struct {
	UserID int64
	Type   OrderType // limit or market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

type Market string

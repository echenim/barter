package models

type Order struct {
	UserID    int64
	ID        int64
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type OrderType string

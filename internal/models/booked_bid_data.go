package models

type BookedBidData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Asks           []*Order
	Bids           []*Order
}

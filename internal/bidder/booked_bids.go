package bidder

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Note
// here buy and sell orders are matched in real-time, and efficient,
// thread-safe operations are crucial. The use of mutexes and careful structuring of
// the methods ensures that the order book can be accessed and modified safely in a
// concurrent environment.

type BookBid struct {
	asks []*Limit
	bids []*Limit

	Trades []*Trade

	mu        sync.RWMutex
	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
	Orders    map[int64]*Bid
}

// NewBookBid creates and returns a new instance of BookBid.
// This function initializes slices for asks, bids, and trades,
// and creates maps for AskLimits, BidLimits, and Orders.
func NewBookBid() *BookBid {
	return &BookBid{
		asks:      []*Limit{},
		bids:      []*Limit{},
		Trades:    []*Trade{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Bid),
	}
}

// PlaceMarketOrder places a market bid into the book.
// It takes a Bid object as a parameter and returns a slice of Match objects.
// This method locks the BookBid for concurrent access, calculates matches for the given bid,
// and records the trades. It panics if there is insufficient volume for the bid.
func (ob *BookBid) PlaceMarketOrder(o *Bid) []Match {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	matches := []Match{}

	if o.Bid {
		if o.Size > ob.AskTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.AskTotalVolume(), o.Size))
		}

		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(false, limit)
			}
		}
	} else {
		if o.Size > ob.BidTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.BidTotalVolume(), o.Size))
		}

		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}
		}
	}

	for _, match := range matches {
		trade := &Trade{
			Price:     match.Price,
			Size:      match.SizeFilled,
			Timestamp: time.Now().UnixNano(),
			Bid:       o.Bid,
		}
		ob.Trades = append(ob.Trades, trade)
	}

	logrus.WithFields(logrus.Fields{
		"currentPrice": ob.Trades[len(ob.Trades)-1].Price,
	}).Info()

	return matches
}

// PlaceLimitOrder places a limit order in the book.
// It locks the current state, checks or creates the necessary limit,
// logs the new order information, and adds the order to the limit and book.
func (ob *BookBid) PlaceLimitOrder(price float64, o *Bid) {
	var limit *Limit

	ob.mu.Lock()
	defer ob.mu.Unlock()

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)

		if o.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	logrus.WithFields(logrus.Fields{
		"price":  limit.Price,
		"type":   o.Type(),
		"size":   o.Size,
		"userID": o.UserID,
	}).Info("new limit order")

	ob.Orders[o.ID] = o
	limit.AddOrder(o)
}

// clearLimit removes a limit from the book.
// It is used when all orders at a limit have been fulfilled.
// The function updates the limits map and the bids or asks slice depending on the type of limit.
func (ob *BookBid) clearLimit(bid bool, l *Limit) {
	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}
		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}
		}
	}

	fmt.Printf("clearing limit price level [%.2f]\n", l.Price)
}

// CancelOrder handles the cancellation of an order.
// It removes the order from its limit and the book, and clears the limit if it becomes empty.
func (ob *BookBid) CancelOrder(o *Bid) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.ID)

	if len(limit.Orders) == 0 {
		ob.clearLimit(o.Bid, limit)
	}
}

// BidTotalVolume calculates the total volume of all bid orders.
// It iterates over all bids and sums up their total volumes.
func (ob *BookBid) BidTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}

	return totalVolume
}

// AskTotalVolume calculates the total volume of all ask orders.
// It iterates over all asks and sums up their total volumes.
func (ob *BookBid) AskTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}

	return totalVolume
}

// Asks returns a sorted slice of all ask limits.
// The limits are sorted based on the criteria defined in ByBestAsk.
func (ob *BookBid) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

// Bids returns a sorted slice of all bid limits.
// The limits are sorted based on the criteria defined in ByBestBid.
func (ob *BookBid) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}

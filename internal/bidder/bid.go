package bidder

import (
	"fmt"
	"math/rand"
	"time"
)

type Bid struct {
	ID        int64
	UserID    int64
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

type Bids []*Bid

func (o Bids) Len() int           { return len(o) }
func (o Bids) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Bids) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

func NewBid(bid bool, size float64, userID int64) *Bid {
	return &Bid{
		UserID:    userID,
		ID:        int64(rand.Intn(10000000)),
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Bid) String() string {
	return fmt.Sprintf("[size: %.2f] | [id: %d]", o.Size, o.ID)
}

func (o *Bid) Type() string {
	if o.Bid {
		return "BID"
	}
	return "ASK"
}

func (o *Bid) IsFilled() bool {
	return o.Size == 0.0
}

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

// Len returns the number of elements in the Bids slice.
func (o Bids) Len() int { return len(o) }

// Swap exchanges the elements with indexes i and j in the Bids slice.
func (o Bids) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

// Less reports whether the element with index i should sort before the element with index j.
// It sorts based on the Timestamp, making it useful for time-based sorting.
func (o Bids) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

// NewBid creates and returns a new Bid instance.
// It initializes a Bid with the provided bid status, size, and userID,
// assigns a random ID, and sets the current time as the Timestamp.
func NewBid(bid bool, size float64, userID int64) *Bid {
	return &Bid{
		UserID:    userID,
		ID:        int64(rand.Intn(10000000)),
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// String provides a string representation of the Bid instance.
// It formats the Bid's size and ID in a readable format.
func (o *Bid) String() string {
	return fmt.Sprintf("[size: %.2f] | [id: %d]", o.Size, o.ID)
}

// Type returns a string indicating whether the Bid is a "BID" or an "ASK".
// This is determined based on the Bid's boolean 'Bid' field.
func (o *Bid) Type() string {
	if o.Bid {
		return "BID"
	}
	return "ASK"
}

// IsFilled checks if the Bid's size is zero.
// Returns true if the size is zero, indicating that the Bid is filled.
func (o *Bid) IsFilled() bool {
	return o.Size == 0.0
}

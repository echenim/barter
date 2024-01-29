package bidder

import "sort"

// Note
// This code involves creating and managing limit orders in a financial trading system.
// Each Limit object represents a specific price point and contains orders (Bid objects)
// at that price. The code allows adding and removing orders, and processing (filling)
// incoming orders against existing ones. The filling process involves matching orders
// based on their size and updating them accordingly.

// Limit represents a single limit with a price.
// Assuming there's a predefined struct 'Limit' with a field 'Price'.
type Limit struct {
	Price       float64
	Orders      Bids
	TotalVolume float64
}

// Limits is a slice of pointers to Limit objects.
type Limits []*Limit

// NewLimit creates a new Limit object with the specified price.
// It initializes an empty slice of Bids for the Orders field.
func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Bid{},
	}
}

// AddOrder adds a new Bid to the Limit.
// It sets the Limit of the Bid to itself and updates the total volume of the Limit.
func (l *Limit) AddOrder(o *Bid) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

// DeleteOrder removes a Bid from the Limit's orders.
// It adjusts the slice to remove the Bid and updates the total volume.
// Finally, it sorts the Orders slice.
func (l *Limit) DeleteOrder(o *Bid) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders[i] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}

	o.Limit = nil
	l.TotalVolume -= o.Size

	sort.Sort(l.Orders)
}

// Fill processes an incoming Bid and attempts to fill it with existing Orders.
// It returns a slice of Match objects representing the filled orders.
func (l *Limit) Fill(o *Bid) []Match {
	var (
		matches        []Match
		ordersToDelete []*Bid
	)

	for _, order := range l.Orders {
		if o.IsFilled() {
			break
		}

		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.IsFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}
	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

// fillOrder calculates the match between two Bids (a and b).
// It updates the size of the Bids based on the match and returns a Match object.
func (l *Limit) fillOrder(a, b *Bid) Match {
	var (
		bid        *Bid
		ask        *Bid
		sizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.Size >= b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0.0
	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}

	return Match{
		Bid:        bid,
		Ask:        ask,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}

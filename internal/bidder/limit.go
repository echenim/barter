package bidder

import "sort"

type Limit struct {
	Price       float64
	Orders      Bids
	TotalVolume float64
}

type Limits []*Limit

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Bid{},
	}
}

func (l *Limit) AddOrder(o *Bid) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

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

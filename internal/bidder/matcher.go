package bidder

type Match struct {
	Ask        *Bid
	Bid        *Bid
	SizeFilled float64
	Price      float64
}

package bidder

// Note
// These two types, ByBestAsk and ByBestBid, are tailored for sorting slices of a Limit struct 
// (presumably containing a Price field) in ascending and descending order of their prices,
//  respectively. They implement the sort.Interface, which requires the methods Len, Swap,
//   and Less. This allows the use of Go's built-in sort package to sort slices of these types.


// ByBestAsk is a struct that embeds the Limits slice.
// It is used for sorting Limit structures based on the Price in ascending order (i.e., from the lowest to highest price).
type ByBestAsk struct{ Limits }

// Len returns the number of elements in the Limits slice.
// This method is required by the sort.Interface in Go.
func (a ByBestAsk) Len() int {
	return len(a.Limits)
}

// Swap exchanges the elements with indices i and j in the Limits slice.
// This method is also required by the sort.Interface.
func (a ByBestAsk) Swap(i, j int) {
	a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i]
}

// Less reports whether the element with index i should sort before the element with index j.
// In ByBestAsk, it compares the Prices and returns true if the price at index i is less than the price at index j.
// This is used to sort the Limits in ascending order of their Prices.
func (a ByBestAsk) Less(i, j int) bool {
	return a.Limits[i].Price < a.Limits[j].Price
}

// ByBestBid is a struct that embeds the Limits slice.
// It is used for sorting Limit structures based on the Price in descending order (i.e., from the highest to lowest price).
type ByBestBid struct{ Limits }

// Len returns the number of elements in the Limits slice.
// This method is consistent with the sort.Interface requirements.
func (b ByBestBid) Len() int {
	return len(b.Limits)
}

// Swap exchanges the elements with indices i and j in the Limits slice.
// This method functions identically to Swap in ByBestAsk, adhering to the sort.Interface.
func (b ByBestBid) Swap(i, j int) {
	b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i]
}

// Less reports whether the element with index i should sort before the element with index j.
// In ByBestBid, it compares the Prices and returns true if the price at index i is greater than the price at index j.
// This is used to sort the Limits in descending order of their Prices.
func (b ByBestBid) Less(i, j int) bool {
	return b.Limits[i].Price > b.Limits[j].Price
}

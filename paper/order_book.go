package paper

import (
	"sort"
	"strconv"
)

// OrderBookLevel is a single price level. Price and size are kept as strings
// (as the exchange delivers them) and parsed on demand; the original string is
// preserved so round-tripping is exact.
type OrderBookLevel struct {
	Price string
	Size  string
}

// LevelFromMap builds a level from a {"price":..,"size":..} map, accepting
// "remaining_base_amount" as an alias for size (matching the Python
// OrderBookLevel.from_any behaviour). Returns an error if either is missing.
func LevelFromMap(m map[string]string) (OrderBookLevel, error) {
	price, okP := m["price"]
	size, okS := m["size"]
	if !okS {
		size, okS = m["remaining_base_amount"]
	}
	if !okP || !okS {
		return OrderBookLevel{}, &ParamError{Msg: "order book levels must include price and size data"}
	}
	return OrderBookLevel{Price: price, Size: size}, nil
}

func (l OrderBookLevel) priceFloat() (float64, bool) {
	v, err := strconv.ParseFloat(l.Price, 64)
	return v, err == nil
}

func (l OrderBookLevel) sizeFloat() (float64, bool) {
	v, err := strconv.ParseFloat(l.Size, 64)
	return v, err == nil
}

// BookData is a snapshot or delta payload: replacement (snapshot) or
// incremental (delta) ask/bid levels plus an optional sequence offset.
type BookData struct {
	Asks   []OrderBookLevel
	Bids   []OrderBookLevel
	Offset *int
}

// InMemoryOrderBook keeps asks ascending and bids descending so the best
// prices are always at index 0.
type InMemoryOrderBook struct {
	Asks   []OrderBookLevel
	Bids   []OrderBookLevel
	Offset *int
}

// NewOrderBook builds a book from initial levels, sorting and filtering them.
func NewOrderBook(asks, bids []OrderBookLevel) *InMemoryOrderBook {
	return &InMemoryOrderBook{
		Asks: sortedLevels(asks, true),
		Bids: sortedLevels(bids, false),
	}
}

// ApplySnapshot replaces the book wholesale.
func (b *InMemoryOrderBook) ApplySnapshot(data BookData) {
	b.Asks = sortedLevels(data.Asks, true)
	b.Bids = sortedLevels(data.Bids, false)
	b.Offset = data.Offset
}

// ApplyDelta merges incremental updates: a level with size 0 removes the
// matching price, any other size inserts or replaces it.
func (b *InMemoryOrderBook) ApplyDelta(data BookData) {
	b.Asks = mergeLevels(b.Asks, data.Asks, true)
	b.Bids = mergeLevels(b.Bids, data.Bids, false)
	b.Offset = data.Offset
}

// BestAsk returns the lowest ask, or nil if there are none.
func (b *InMemoryOrderBook) BestAsk() *OrderBookLevel {
	if len(b.Asks) == 0 {
		return nil
	}
	return &b.Asks[0]
}

// BestBid returns the highest bid, or nil if there are none.
func (b *InMemoryOrderBook) BestBid() *OrderBookLevel {
	if len(b.Bids) == 0 {
		return nil
	}
	return &b.Bids[0]
}

// MidPrice returns the mid price; ok is false if either side is empty.
func (b *InMemoryOrderBook) MidPrice() (float64, bool) {
	ask, bid := b.BestAsk(), b.BestBid()
	if ask == nil || bid == nil {
		return 0, false
	}
	a, okA := ask.priceFloat()
	d, okB := bid.priceFloat()
	if !okA || !okB {
		return 0, false
	}
	return (a + d) / 2, true
}

// sortedLevels normalizes (drops unparseable/zero/negative sizes) then stably
// sorts by price: ascending for asks, descending for bids.
func sortedLevels(levels []OrderBookLevel, isAsks bool) []OrderBookLevel {
	out := make([]OrderBookLevel, 0, len(levels))
	for _, l := range levels {
		size, ok := l.sizeFloat()
		if !ok || size <= 0 {
			continue
		}
		if _, ok := l.priceFloat(); !ok {
			continue
		}
		out = append(out, l)
	}
	sort.SliceStable(out, func(i, j int) bool {
		pi, _ := out[i].priceFloat()
		pj, _ := out[j].priceFloat()
		if isAsks {
			return pi < pj
		}
		return pi > pj
	})
	return out
}

// mergeLevels applies new levels onto existing ones, keyed by the price string,
// preserving insertion order before the final stable sort (matching Python's
// dict-based merge).
func mergeLevels(existing, updates []OrderBookLevel, isAsks bool) []OrderBookLevel {
	order := make([]string, 0, len(existing)+len(updates))
	byPrice := make(map[string]OrderBookLevel, len(existing)+len(updates))
	put := func(l OrderBookLevel) {
		if _, seen := byPrice[l.Price]; !seen {
			order = append(order, l.Price)
		}
		byPrice[l.Price] = l
	}

	for _, l := range sortedLevels(existing, isAsks) {
		put(l)
	}
	for _, l := range updates {
		size, ok := l.sizeFloat()
		if ok && size == 0 {
			if _, present := byPrice[l.Price]; present {
				delete(byPrice, l.Price)
				order = removeString(order, l.Price)
			}
			continue
		}
		put(l)
	}

	merged := make([]OrderBookLevel, 0, len(order))
	for _, price := range order {
		merged = append(merged, byPrice[price])
	}
	return sortedLevels(merged, isAsks)
}

func removeString(s []string, target string) []string {
	for i, v := range s {
		if v == target {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

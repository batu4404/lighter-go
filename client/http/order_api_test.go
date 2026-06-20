package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core "github.com/0xJord4n/lighter-go/client"
)

// orderBookOrdersGolden is shaped like a real /api/v1/orderBookOrders response
// (verified against mainnet). The critical fields are price and
// remaining_base_amount -- earlier versions of the Go SDK mapped these to
// the wrong field names (size/side), causing silent data loss.
const orderBookOrdersGolden = `{
  "code": 200,
  "asks": [
    {
      "initial_base_amount": "3.5754",
      "order_expiry": 0,
      "order_id": "0x1234",
      "order_index": 101,
      "owner_account_index": 7,
      "price": "1691.51",
      "remaining_base_amount": "3.5754",
      "transaction_time": 1718800000
    }
  ],
  "bids": [
    {
      "initial_base_amount": "2.0000",
      "order_expiry": 0,
      "order_id": "0x5678",
      "order_index": 99,
      "owner_account_index": 8,
      "price": "1690.00",
      "remaining_base_amount": "2.0000",
      "transaction_time": 1718799999
    }
  ]
}`

func TestGetOrderBookOrders_GoldenResponse(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderBookOrdersGolden))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	ob, err := c.Order().GetOrderBookOrders(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("GetOrderBookOrders error: %v", err)
	}

	if gotPath != "/api/v1/orderBookOrders" {
		t.Errorf("path = %q, want /api/v1/orderBookOrders", gotPath)
	}

	if len(ob.Asks) != 1 {
		t.Fatalf("asks len = %d, want 1", len(ob.Asks))
	}
	if ob.Asks[0].Price != "1691.51" {
		t.Errorf("ask price = %q, want 1691.51", ob.Asks[0].Price)
	}
	if ob.Asks[0].RemainingBaseAmount != "3.5754" {
		t.Errorf("ask remaining_base_amount = %q, want 3.5754", ob.Asks[0].RemainingBaseAmount)
	}
	if len(ob.Bids) != 1 {
		t.Fatalf("bids len = %d, want 1", len(ob.Bids))
	}
	if ob.Bids[0].Price != "1690.00" {
		t.Errorf("bid price = %q, want 1690.00", ob.Bids[0].Price)
	}
}

// exchangeStatsGolden is shaped like a real /api/v1/exchangeStats response.
const exchangeStatsGolden = `{
  "code": 200,
  "total": 4,
  "order_book_stats": [],
  "total_volume_24h": "123456.78",
  "total_trades_24h": 9999,
  "total_open_interest": "5000.00",
  "total_users": 42,
  "total_markets": 4,
  "timestamp": 1718800000
}`

func TestGetExchangeStats_GoldenResponse(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(exchangeStatsGolden))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	stats, err := c.Order().GetExchangeStats(context.Background())
	if err != nil {
		t.Fatalf("GetExchangeStats error: %v", err)
	}
	if gotPath != "/api/v1/exchangeStats" {
		t.Errorf("path = %q, want /api/v1/exchangeStats", gotPath)
	}
	if stats.TotalTrades24h != 9999 {
		t.Errorf("total_trades_24h = %d, want 9999", stats.TotalTrades24h)
	}
	if stats.TotalVolume24h != "123456.78" {
		t.Errorf("total_volume_24h = %q, want 123456.78", stats.TotalVolume24h)
	}
}

// TestGetTrades_UsesCorrectSortDirParam verifies the bug-fix: lighter-python
// names this param "sort_dir", not "sort_order".
func TestGetTrades_UsesCorrectSortDirParam(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trades":[]}`))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	_, _ = c.Order().GetTrades(context.Background(), 0, nil, &core.TradesOpts{SortOrder: "desc"})

	if strings.Contains(gotQuery, "sort_order") {
		t.Errorf("query contains sort_order; should be sort_dir. Got: %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "sort_dir=desc") {
		t.Errorf("query missing sort_dir=desc. Got: %s", gotQuery)
	}
}

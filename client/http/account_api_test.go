package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core "github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/types/api"
)

// accountGolden is shaped like a real /api/v1/account response.
const accountGolden = `{
  "code": 200,
  "accounts": [
    {
      "index": 42,
      "l1_address": "0xDeAdBeEf",
      "nonce": 7,
      "collateral_value": "1000.00",
      "position_value": "500.00",
      "portfolio_value": "1500.00",
      "available_balance": "900.00",
      "max_withdrawable": "800.00",
      "initial_margin": "50.00",
      "maintenance_margin": "25.00",
      "unrealized_pnl": "10.00",
      "is_liquidatable": false
    }
  ]
}`

func TestGetAccount_GoldenResponse(t *testing.T) {
	var gotPath string
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(accountGolden))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	accs, err := c.Account().GetAccount(context.Background(), api.QueryByIndex, "42")
	if err != nil {
		t.Fatalf("GetAccount error: %v", err)
	}

	if gotPath != "/api/v1/account" {
		t.Errorf("path = %q, want /api/v1/account", gotPath)
	}
	// GetAccount maps QueryByIndex -> by=index&value=<val>
	if !strings.Contains(gotQuery, "by=index") {
		t.Errorf("query missing by=index, got: %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "value=42") {
		t.Errorf("query missing value=42, got: %s", gotQuery)
	}

	if len(accs.Accounts) != 1 {
		t.Fatalf("accounts len = %d, want 1", len(accs.Accounts))
	}
	acc := accs.Accounts[0]
	if acc.Index != 42 {
		t.Errorf("index = %d, want 42", acc.Index)
	}
	if acc.L1Address != "0xDeAdBeEf" {
		t.Errorf("l1_address = %q, want 0xDeAdBeEf", acc.L1Address)
	}
	if acc.PortfolioValue != "1500.00" {
		t.Errorf("portfolio_value = %q, want 1500.00", acc.PortfolioValue)
	}
	if acc.IsLiquidatable {
		t.Error("is_liquidatable = true, want false")
	}
}

// TestChangeAccountTier_UsesURLEncoded verifies the bug-fix: this endpoint
// must use application/x-www-form-urlencoded (not multipart).
func TestChangeAccountTier_UsesURLEncoded(t *testing.T) {
	var gotContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	_, _ = c.Account().ChangeAccountTier(context.Background(), 1, "2", "auth123")

	if !strings.HasPrefix(gotContentType, "application/x-www-form-urlencoded") {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", gotContentType)
	}
}

// apiKeysGolden is shaped like a real /api/v1/apiKeys response.
const apiKeysGolden = `{
  "code": 200,
  "api_keys": [
    {
      "key_index": 0,
      "public_key": "0xabc123"
    }
  ]
}`

func TestGetApiKeys_GoldenResponse(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(apiKeysGolden))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	// nil apiKeyIndex = fetch all keys
	keys, err := c.Account().GetApiKeys(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("GetApiKeys error: %v", err)
	}
	if gotPath != "/api/v1/apikeys" {
		t.Errorf("path = %q, want /api/v1/apikeys", gotPath)
	}
	if len(keys.ApiKeys) != 1 {
		t.Fatalf("api_keys len = %d, want 1", len(keys.ApiKeys))
	}
	if keys.ApiKeys[0].PublicKey != "0xabc123" {
		t.Errorf("public_key = %q, want 0xabc123", keys.ApiKeys[0].PublicKey)
	}
}

// TestGetApiKeys_WithKeyIndex verifies the api_key_index param is sent when provided.
func TestGetApiKeys_WithKeyIndex(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"api_keys":[]}`))
	}))
	defer ts.Close()

	idx := uint8(2)
	c := NewFullClient(ts.URL)
	_, _ = c.Account().GetApiKeys(context.Background(), 5, &idx)
	if !strings.Contains(gotQuery, "api_key_index=2") {
		t.Errorf("query missing api_key_index=2, got: %s", gotQuery)
	}
}

var _ core.AccountAPI = nil // suppress unused import

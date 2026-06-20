package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSendTx_UsesURLEncoded verifies the critical bug-fix: SendTx must use
// application/x-www-form-urlencoded (lighter-python uses form encoding, not
// multipart). Using multipart causes the Lighter API to reject the request.
func TestSendTx_UsesURLEncoded(t *testing.T) {
	var gotContentType string
	var gotFormTxType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_ = r.ParseForm()
		gotFormTxType = r.PostFormValue("tx_type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	_, _ = c.Transaction().SendTx(context.Background(), 3, "txdata", nil)

	if !strings.HasPrefix(gotContentType, "application/x-www-form-urlencoded") {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", gotContentType)
	}
	if gotFormTxType != "3" {
		t.Errorf("tx_type form field = %q, want 3", gotFormTxType)
	}
}

// TestSendTx_ForwardsTxInfo verifies tx_info is sent as a form field.
func TestSendTx_ForwardsTxInfo(t *testing.T) {
	var gotTxInfo string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotTxInfo = r.PostFormValue("tx_info")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	_, _ = c.Transaction().SendTx(context.Background(), 1, "encoded_tx_payload", nil)

	if gotTxInfo != "encoded_tx_payload" {
		t.Errorf("tx_info = %q, want encoded_tx_payload", gotTxInfo)
	}
}

// depositHistoryGolden is shaped like a real /api/v1/deposit/history response.
const depositHistoryGolden = `{
  "code": 200,
  "deposits": [
    {
      "deposit_index": 1,
      "account_index": 7,
      "l1_address": "0xDEAD",
      "l1_tx_hash": "0xdeadbeef",
      "asset_index": 0,
      "amount": "100.00",
      "status": "confirmed",
      "created_at": 1718800000
    }
  ]
}`

func TestGetDepositHistory_GoldenResponse(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(depositHistoryGolden))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	hist, err := c.Transaction().GetDepositHistory(context.Background(), 1, "", "", "")
	if err != nil {
		t.Fatalf("GetDepositHistory error: %v", err)
	}
	if gotPath != "/api/v1/deposit/history" {
		t.Errorf("path = %q, want /api/v1/deposit/history", gotPath)
	}
	if len(hist.Deposits) != 1 {
		t.Fatalf("deposits len = %d, want 1", len(hist.Deposits))
	}
	if hist.Deposits[0].Amount != "100.00" {
		t.Errorf("amount = %q, want 100.00", hist.Deposits[0].Amount)
	}
	if hist.Deposits[0].L1TxHash != "0xdeadbeef" {
		t.Errorf("l1_tx_hash = %q, want 0xdeadbeef", hist.Deposits[0].L1TxHash)
	}
}

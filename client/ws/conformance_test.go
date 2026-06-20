//go:build wsconformance

// Package ws conformance test: verifies our WebSocket channels against the
// LIVE Lighter server. This is the WS equivalent of tools/parity_check.py for
// the signer -- there is no machine-readable WS schema (openapi.json describes
// zero stream endpoints, and lighter-python's ws_client.py only implements 2 of
// our channels), so the running server is the only source of truth.
//
// It is build-tagged so it never runs under a normal `go test ./...` (it needs
// network). Run it explicitly:
//
//	make ws-conformance              # public channels only
//	LIGHTER_WS_AUTH=<token> LIGHTER_WS_ACCOUNT=<index> make ws-conformance
//
// A failure means Lighter changed their WS protocol (channel renamed/removed,
// or the subscribe handshake changed) and our client has drifted -- exactly the
// drift we otherwise could not detect.
package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/coder/websocket"
)

const conformanceURL = "wss://mainnet.zklighter.elliot.ai/stream"

// wsFrame is the minimal envelope every Lighter WS message shares.
type wsFrame struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

// dialAndAwaitConnected opens the socket and waits for the server's initial
// "connected" frame, mirroring what wsClient.Connect does.
func dialAndAwaitConnected(t *testing.T, ctx context.Context) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.Dial(ctx, conformanceURL, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", conformanceURL, err)
	}
	// Match the production client: order-book snapshots exceed the 32KB
	// coder/websocket default (see client.go SetReadLimit).
	conn.SetReadLimit(10 * 1024 * 1024)
	for i := 0; i < 5; i++ {
		f := readFrame(t, ctx, conn)
		if f.Type == "connected" {
			return conn
		}
	}
	t.Fatal(`never received "connected" handshake frame`)
	return nil
}

func readFrame(t *testing.T, ctx context.Context, conn *websocket.Conn) wsFrame {
	t.Helper()
	rctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, data, err := conn.Read(rctx)
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	var f wsFrame
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("unmarshal frame %q: %v", string(data), err)
	}
	// Server pings must be answered or it drops us mid-test.
	if f.Type == "ping" {
		_ = conn.Write(rctx, websocket.MessageText, []byte(`{"type":"pong"}`))
		return readFrame(t, ctx, conn)
	}
	return f
}

// assertSubscribe sends a subscribe for channel and asserts the server acks
// with the expected "subscribed/<wantType>" message (not an error frame).
func assertSubscribe(t *testing.T, ctx context.Context, conn *websocket.Conn, channel, wantType string) {
	t.Helper()
	sub := fmt.Sprintf(`{"type":"subscribe","channel":%q}`, channel)
	wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := conn.Write(wctx, websocket.MessageText, []byte(sub)); err != nil {
		t.Fatalf("subscribe %s: %v", channel, err)
	}
	want := "subscribed/" + wantType
	for i := 0; i < 6; i++ {
		f := readFrame(t, ctx, conn)
		switch {
		case f.Type == want:
			return
		case f.Type == "error":
			t.Fatalf("channel %q: server returned error frame instead of %q", channel, want)
			// ignore unrelated update/* frames that may arrive before the ack
		}
	}
	t.Fatalf("channel %q: never received %q ack", channel, want)
}

// TestWSPublicChannelsConformance exercises every no-auth channel against the
// live server. The channel strings here MUST match what client.go builds.
func TestWSPublicChannelsConformance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cases := []struct{ channel, wantType string }{
		{"order_book/0", "order_book"},
		{"trade/0", "trade"},
		{"market_stats/0", "market_stats"},
		{"market_stats/all", "market_stats"},
		{"height", "height"},
	}
	for _, tc := range cases {
		t.Run(tc.channel, func(t *testing.T) {
			conn := dialAndAwaitConnected(t, ctx)
			defer conn.Close(websocket.StatusNormalClosure, "")
			assertSubscribe(t, ctx, conn, tc.channel, tc.wantType)
		})
	}
}

// TestWSAuthedChannelsConformance exercises the account/pool channels. It runs
// only when LIGHTER_WS_AUTH and LIGHTER_WS_ACCOUNT are set, since these require
// a valid auth token + real account index. The channel strings MUST match
// what client.go builds for the corresponding Subscribe* methods.
func TestWSAuthedChannelsConformance(t *testing.T) {
	auth := os.Getenv("LIGHTER_WS_AUTH")
	acctStr := os.Getenv("LIGHTER_WS_ACCOUNT")
	if auth == "" || acctStr == "" {
		t.Skip("set LIGHTER_WS_AUTH and LIGHTER_WS_ACCOUNT to run authed WS conformance")
	}
	acct, err := strconv.ParseInt(acctStr, 10, 64)
	if err != nil {
		t.Fatalf("LIGHTER_WS_ACCOUNT %q is not an int: %v", acctStr, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cases := []struct{ channel, wantType string }{
		{fmt.Sprintf("account_all/%d", acct), "account_all"},
		{fmt.Sprintf("account_all_orders/%d", acct), "account_all_orders"},
		{fmt.Sprintf("account_all_trades/%d", acct), "account_all_trades"},
		{fmt.Sprintf("account_all_positions/%d", acct), "account_all_positions"},
		{fmt.Sprintf("account_tx/%d", acct), "account_tx"},
		{fmt.Sprintf("user_stats/%d", acct), "user_stats"},
		{fmt.Sprintf("pool_data/%d", acct), "pool_data"},
		{fmt.Sprintf("pool_info/%d", acct), "pool_info"},
		{fmt.Sprintf("notification/%d", acct), "notification"},
	}
	for _, tc := range cases {
		t.Run(tc.wantType, func(t *testing.T) {
			conn := dialAndAwaitConnected(t, ctx)
			defer conn.Close(websocket.StatusNormalClosure, "")
			// Authed channels carry the auth token in the subscribe frame.
			sub := fmt.Sprintf(`{"type":"subscribe","channel":%q,"auth":%q}`, tc.channel, auth)
			wctx, c := context.WithTimeout(ctx, 10*time.Second)
			defer c()
			if err := conn.Write(wctx, websocket.MessageText, []byte(sub)); err != nil {
				t.Fatalf("subscribe %s: %v", tc.channel, err)
			}
			want := "subscribed/" + tc.wantType
			for i := 0; i < 6; i++ {
				f := readFrame(t, ctx, conn)
				if f.Type == want {
					return
				}
				if f.Type == "error" {
					t.Fatalf("channel %q: server error instead of %q", tc.channel, want)
				}
			}
			t.Fatalf("channel %q: never received %q ack", tc.channel, want)
		})
	}
}

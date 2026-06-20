package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core "github.com/0xJord4n/lighter-go/client"
)

func TestReferralUse_Serialization(t *testing.T) {
	var (
		gotMethod      string
		gotPath        string
		gotAuth        string
		gotContentType string
		gotFields      = map[string]string{}
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		for _, k := range []string{"l1_address", "referral_code", "x", "auth", "telegram", "discord"} {
			gotFields[k] = r.PostFormValue(k)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	err := c.Referral().ReferralUse(context.Background(), "0xAbC", "HELLFUN", "myhandle", "authtoken123", &core.ReferralUseOpts{
		Telegram: "tguser",
	})
	if err != nil {
		t.Fatalf("ReferralUse returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/api/v1/referral/use" {
		t.Errorf("path = %s, want /api/v1/referral/use", gotPath)
	}
	// Lighter's OpenAPI specifies application/x-www-form-urlencoded for this body.
	if !strings.HasPrefix(gotContentType, "application/x-www-form-urlencoded") {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", gotContentType)
	}
	if gotAuth != "authtoken123" {
		t.Errorf("Authorization header = %q, want authtoken123", gotAuth)
	}
	want := map[string]string{
		"l1_address":    "0xAbC",
		"referral_code": "HELLFUN",
		"x":             "myhandle",
		"auth":          "authtoken123",
		"telegram":      "tguser",
		"discord":       "",
	}
	for k, v := range want {
		if gotFields[k] != v {
			t.Errorf("form field %s = %q, want %q", k, gotFields[k], v)
		}
	}
}

func TestReferralUse_OmitsEmptyX(t *testing.T) {
	var hasX bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, hasX = r.PostForm["x"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	if err := c.Referral().ReferralUse(context.Background(), "0xAbC", "HELLFUN", "", "tok", nil); err != nil {
		t.Fatalf("ReferralUse returned error: %v", err)
	}
	if hasX {
		t.Error("x field should be omitted when empty")
	}
}

func TestReferralUse_APIErrorReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"message":"invalid referral code"}`))
	}))
	defer ts.Close()

	c := NewFullClient(ts.URL)
	err := c.Referral().ReferralUse(context.Background(), "0xAbC", "BAD", "h", "tok", nil)
	if err == nil {
		t.Fatal("expected error for non-OK code, got nil")
	}
}

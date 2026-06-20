package client

import "testing"

func TestAreKeysEqual(t *testing.T) {
	cases := []struct {
		name      string
		k1, k2    string
		wantEqual bool
	}{
		{"identical no prefix", "abcd", "abcd", true},
		{"one prefixed", "0xabcd", "abcd", true},
		{"both prefixed", "0xabcd", "0xabcd", true},
		{"different", "0xabcd", "0xabce", false},
		{"prefix only mismatch on body", "0x1234", "5678", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := AreKeysEqual(tc.k1, tc.k2); got != tc.wantEqual {
				t.Errorf("AreKeysEqual(%q,%q) = %v, want %v", tc.k1, tc.k2, got, tc.wantEqual)
			}
		})
	}
}

func TestValidateAPIPrivateKeys(t *testing.T) {
	t.Run("empty set is rejected", func(t *testing.T) {
		if err := ValidateAPIPrivateKeys(map[int]string{}); err == nil {
			t.Fatal("expected error for empty key set")
		}
	})

	t.Run("blank key is rejected", func(t *testing.T) {
		if err := ValidateAPIPrivateKeys(map[int]string{0: "0x"}); err == nil {
			t.Fatal("expected error for blank key")
		}
	})

	t.Run("strips 0x in place", func(t *testing.T) {
		keys := map[int]string{0: "0xdeadbeef", 1: "cafe"}
		if err := ValidateAPIPrivateKeys(keys); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if keys[0] != "deadbeef" {
			t.Errorf("key[0] = %q, want deadbeef", keys[0])
		}
		if keys[1] != "cafe" {
			t.Errorf("key[1] = %q, want cafe", keys[1])
		}
	})
}

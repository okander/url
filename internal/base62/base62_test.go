package base62

import (
	"math"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	for _, n := range []uint64{0, 1, 61, 62, 63, 12345, 1 << 32, math.MaxUint64} {
		s := Encode(n)
		got, err := Decode(s)
		if err != nil {
			t.Fatalf("Decode(%q) returned error: %v", s, err)
		}
		if got != n {
			t.Fatalf("round trip failed: %d -> %q -> %d", n, s, got)
		}
	}
}

func TestEncodeKnownValues(t *testing.T) {
	cases := map[uint64]string{0: "0", 61: "Z", 62: "10"}
	for n, want := range cases {
		if got := Encode(n); got != want {
			t.Errorf("Encode(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestDecodeInvalid(t *testing.T) {
	for _, s := range []string{"", "!", "abc-def", "zzzzzzzzzzzzzzzz"} {
		if _, err := Decode(s); err == nil {
			t.Errorf("Decode(%q) expected an error", s)
		}
	}
}

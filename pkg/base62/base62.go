// Package base62 converts numeric IDs to short, URL-safe strings and back.
package base62

import (
	"errors"
	"math"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ErrInvalid is returned by Decode for empty, malformed or overflowing input.
var ErrInvalid = errors.New("base62: invalid input")

// Encode returns the base62 representation of n.
func Encode(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var buf [11]byte // 62^11 > MaxUint64, so 11 digits is always enough
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = alphabet[n%62]
		n /= 62
	}
	return string(buf[i:])
}

// Decode parses a string produced by Encode.
func Decode(s string) (uint64, error) {
	if s == "" {
		return 0, ErrInvalid
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		idx := strings.IndexByte(alphabet, s[i])
		if idx < 0 {
			return 0, ErrInvalid
		}
		if n > (math.MaxUint64-uint64(idx))/62 {
			return 0, ErrInvalid
		}
		n = n*62 + uint64(idx)
	}
	return n, nil
}

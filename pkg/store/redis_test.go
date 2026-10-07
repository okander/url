package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"shortener/pkg/base62"
)

// fakeUpstash mimics the tiny part of the Upstash REST API we use.
func fakeUpstash(t *testing.T) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	counter := 0
	data := map[string]string{}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		var args []string
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil || len(args) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad command"}`))
			return
		}

		mu.Lock()
		defer mu.Unlock()
		switch args[0] {
		case "INCR":
			counter++
			fmt.Fprintf(w, `{"result":%d}`, counter)
		case "SET":
			data[args[1]] = args[2]
			_, _ = w.Write([]byte(`{"result":"OK"}`))
		case "GET":
			v, ok := data[args[1]]
			if !ok {
				_, _ = w.Write([]byte(`{"result":null}`))
				return
			}
			b, _ := json.Marshal(v)
			fmt.Fprintf(w, `{"result":%s}`, b)
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"unknown command"}`))
		}
	}))
}

func TestRedisCreateAndGet(t *testing.T) {
	srv := fakeUpstash(t)
	defer srv.Close()
	r := NewRedis(srv.URL, "test-token")
	ctx := context.Background()

	link, err := r.Create(ctx, "https://example.com/a")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if want := base62.Encode(startID + 1); link.Code != want {
		t.Fatalf("code = %q, want %q", link.Code, want)
	}

	got, err := r.Get(ctx, link.Code)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.URL != "https://example.com/a" {
		t.Fatalf("URL = %q", got.URL)
	}

	second, err := r.Create(ctx, "https://example.com/b")
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}
	if second.Code == link.Code {
		t.Fatal("two links got the same code")
	}
}

func TestRedisGetMissing(t *testing.T) {
	srv := fakeUpstash(t)
	defer srv.Close()
	r := NewRedis(srv.URL, "test-token")

	if _, err := r.Get(context.Background(), "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestRedisBadTokenIsNotNotFound(t *testing.T) {
	srv := fakeUpstash(t)
	defer srv.Close()
	r := NewRedis(srv.URL, "wrong-token")

	_, err := r.Get(context.Background(), "abc")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want a non-NotFound error", err)
	}
}

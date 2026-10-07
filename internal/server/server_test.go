package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"shortener/internal/store"
)

func newTestServer() *Server {
	return New(store.NewMemory(), "http://localhost:8080", "")
}

func postLinkAuth(t *testing.T, srv *Server, target, auth string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(createRequest{URL: target})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func postLink(t *testing.T, srv *Server, target string) *httptest.ResponseRecorder {
	t.Helper()
	return postLinkAuth(t, srv, target, "")
}

func TestCreateAndRedirect(t *testing.T) {
	srv := newTestServer()
	target := "https://example.com/some/long/path?x=1"

	rec := postLink(t, srv, target)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got status %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp createResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code == "" || resp.ShortURL != "http://localhost:8080/"+resp.Code {
		t.Fatalf("unexpected response: %+v", resp)
	}

	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("redirect: got status %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != target {
		t.Fatalf("redirect: got Location %q, want %q", got, target)
	}
}

func TestCreateRejectsBadURLs(t *testing.T) {
	srv := newTestServer()
	for _, bad := range []string{"", "not a url", "javascript:alert(1)", "ftp://example.com", "https://"} {
		if rec := postLink(t, srv, bad); rec.Code != http.StatusBadRequest {
			t.Errorf("url %q: got status %d, want %d", bad, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestUnknownCodeIs404(t *testing.T) {
	srv := newTestServer()
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/doesnotexist", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHealth(t *testing.T) {
	srv := newTestServer()
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAPIKeyProtectsCreateOnly(t *testing.T) {
	srv := New(store.NewMemory(), "", "secret")
	target := "https://example.com/"

	if rec := postLink(t, srv, target); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no key: got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if rec := postLinkAuth(t, srv, target, "Bearer wrong"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong key: got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	rec := postLinkAuth(t, srv, target, "Bearer secret")
	if rec.Code != http.StatusCreated {
		t.Fatalf("right key: got status %d, want %d", rec.Code, http.StatusCreated)
	}
	var resp createResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	// With no base URL configured, the short URL is derived from the request
	// host (httptest requests use "example.com").
	if want := "http://example.com/" + resp.Code; resp.ShortURL != want {
		t.Fatalf("short_url = %q, want %q", resp.ShortURL, want)
	}

	// Redirects stay public: no key needed.
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("redirect: got status %d, want %d", rec.Code, http.StatusFound)
	}
}

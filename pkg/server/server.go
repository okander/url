// Package server contains the HTTP handlers.
package server

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"shortener/pkg/store"
)

const maxURLLength = 2048

// Server wires HTTP routes to a Store.
type Server struct {
	store   store.Store
	baseURL string
	apiKey  string
	mux     *http.ServeMux
}

// New builds a Server.
//
// baseURL is used to build the short links we return; if empty it is derived
// from each request's Host. If apiKey is non-empty, creating links requires
// "Authorization: Bearer <apiKey>" (redirects always stay public).
func New(s store.Store, baseURL, apiKey string) *Server {
	srv := &Server{
		store:   s,
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		mux:     http.NewServeMux(),
	}
	srv.mux.HandleFunc("POST /links", srv.createLink)
	srv.mux.HandleFunc("GET /{$}", srv.index)
	srv.mux.HandleFunc("GET /healthz", srv.health)
	srv.mux.HandleFunc("GET /{code}", srv.redirect)
	return srv
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	URL      string `json:"url"`
}

func (s *Server) createLink(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "missing or invalid API key")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	target, err := validateURL(req.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	link, err := s.store.Create(r.Context(), target)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create link")
		return
	}

	writeJSON(w, http.StatusCreated, createResponse{
		Code:     link.Code,
		ShortURL: s.publicBase(r) + "/" + link.Code,
		URL:      link.URL,
	})
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	link, err := s.store.Get(r.Context(), r.PathValue("code"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not look up link")
		return
	}

	// 302, not 301: browsers cache 301s forever, which would hide repeat
	// visits from the click analytics we add later.
	http.Redirect(w, r, link.URL, http.StatusFound)
}

func (s *Server) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// authorized reports whether the request may create links.
func (s *Server) authorized(r *http.Request) bool {
	if s.apiKey == "" {
		return true
	}
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, prefix) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(h[len(prefix):]), []byte(s.apiKey)) == 1
}

// publicBase returns the configured base URL, or one derived from the request.
func (s *Server) publicBase(r *http.Request) string {
	if s.baseURL != "" {
		return s.baseURL
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// validateURL only allows absolute http(s) URLs. Rejecting other schemes
// (javascript:, data:, file:, ...) keeps the shortener from being used to
// smuggle dangerous links.
func validateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("url is required")
	}
	if len(raw) > maxURLLength {
		return "", errors.New("url is too long")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("url is not valid")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("url must start with http:// or https://")
	}
	if u.Host == "" {
		return "", errors.New("url must include a host")
	}
	return u.String(), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

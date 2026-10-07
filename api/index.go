// Package handler is the Vercel entry point. Vercel turns every exported
// http.HandlerFunc in /api into a serverless function; vercel.json rewrites
// all paths to this one so the router in pkg/server decides what to do.
package handler

import (
	"net/http"
	"os"
	"sync"

	"shortener/pkg/server"
	"shortener/pkg/store"
)

var (
	once sync.Once
	h    http.Handler
)

// Handler serves every request.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		h = server.New(store.FromEnv(), os.Getenv("BASE_URL"), os.Getenv("API_KEY"))
	})
	h.ServeHTTP(w, r)
}

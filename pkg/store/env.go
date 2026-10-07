package store

import (
	"log"
	"os"
)

// FromEnv picks a Store based on environment variables. If Upstash/Vercel
// Redis credentials are present it returns the Redis store (needed on Vercel,
// where functions are stateless); otherwise the in-memory store for local use.
func FromEnv() Store {
	url := firstEnv("UPSTASH_REDIS_REST_URL", "KV_REST_API_URL")
	token := firstEnv("UPSTASH_REDIS_REST_TOKEN", "KV_REST_API_TOKEN")
	if url != "" && token != "" {
		return NewRedis(url, token)
	}
	log.Println("store: no Redis credentials found, using in-memory store (links are lost on restart)")
	return NewMemory()
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

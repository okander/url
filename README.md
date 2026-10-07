# shortener

A URL shortener in Go. Runs locally with zero setup, and deploys to Vercel
(free Hobby plan) with Upstash Redis for storage.

## Run locally

Requires Go 1.22+.

```bash
go test ./...
go run ./cmd/shortener      # or: make run
```

Locally it uses an in-memory store (links vanish on restart). Set
`UPSTASH_REDIS_REST_URL` + `UPSTASH_REDIS_REST_TOKEN` (or Vercel's
`KV_REST_API_URL` + `KV_REST_API_TOKEN`) to use Redis instead.

## API

| Method | Path        | Description                                   |
|--------|-------------|-----------------------------------------------|
| POST   | `/links`    | Body `{"url": "https://..."}` -> short link   |
| GET    | `/{code}`   | 302 redirect to the original URL              |
| GET    | `/`         | Web page for creating links (needs the key)   |
| GET    | `/healthz`  | Health check                                  |

If the `API_KEY` environment variable is set, `POST /links` requires
`Authorization: Bearer <API_KEY>`. Redirects are always public.

```powershell
# PowerShell
Invoke-RestMethod -Method Post -Uri https://YOUR-APP.vercel.app/links `
  -Headers @{ Authorization = "Bearer YOUR_API_KEY" } `
  -ContentType 'application/json' -Body '{"url":"https://example.com"}'
```

```bash
# bash
curl -X POST https://YOUR-APP.vercel.app/links \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com"}'
```

## Deploy to Vercel (free)

1. Push this repo to GitHub.
2. On vercel.com: **Add New > Project**, import the repo, deploy.
3. In the project's **Storage** tab, add an Upstash Redis database and
   connect it to the project (this adds the Redis environment variables).
4. In **Settings > Environment Variables**, add `API_KEY` (any long random
   string). Optionally add `BASE_URL` if you use a custom domain.
5. Redeploy so the new variables take effect.

## Design decisions

- **Base62 codes from an atomic counter (Redis `INCR`).** Short, URL-safe and
  collision-free even with many concurrent serverless instances. Tradeoff:
  codes are enumerable. Custom aliases / random IDs are on the roadmap.
- **302 redirects, not 301.** 301s get cached by browsers, which would hide
  repeat visits from click analytics.
- **Only http/https targets.** Blocks `javascript:`, `data:` and similar.
- **API key on link creation.** A public shortener that anyone can write to
  gets abused; reads stay open, writes need a key (constant-time compare).
- **`Store` interface.** Handlers don't know where data lives: memory for
  local dev and tests, Redis (Upstash REST API, stdlib only) in production.
- **Stateless by design.** Vercel functions don't share memory, so all state
  lives in Redis.

## Roadmap

- [x] Create + redirect, validation, tests
- [x] Redis store, API key, Vercel deployment
- [ ] Async click analytics
- [ ] Custom aliases, expiry, rate limiting
- [ ] TypeScript dashboard
- [ ] GitHub Actions CI, benchmarks (k6)

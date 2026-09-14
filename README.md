# URL Shortener

Project 1 of 8 in a hands-on system design learning series.

## What it does

- `POST /api/shorten` with a JSON body `{"url": "https://example.com/some/long/path"}`
  returns a short code and short URL.
- `GET /{code}` redirects to the original long URL.

## Design notes (V1)

- **Storage**: in-memory Go map, guarded by a `sync.RWMutex`. No persistence yet —
  restarting the server loses all URLs. Swapping this for a real database is a
  planned follow-up milestone (this is where SQL vs NoSQL, indexing, and
  replication tradeoffs come in).
- **Code generation**: an auto-incrementing counter, encoded in base62
  (`0-9A-Za-z`). This guarantees uniqueness with no collision-retry logic,
  unlike hashing or random-string approaches.
- **Concurrency**: multiple requests can safely read/write the map at once
  thanks to the mutex.

## Running it

```bash
go run main.go
```

Then in another terminal:

```bash
curl -X POST http://localhost:8080/api/shorten -d '{"url":"https://go.dev"}'
# => {"short_code":"1","short_url":"http://localhost:8080/1"}

curl -i http://localhost:8080/1
# => HTTP/1.1 302 Found, Location: https://go.dev
```

## Possible next milestones

- Persist mappings to a real database (SQLite to start, then Postgres).
- Add expiration / TTL for links.
- Add click analytics (count redirects per code).
- Add rate limiting on the `/api/shorten` endpoint (ties into Project 2).

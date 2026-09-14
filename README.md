# URL Shortener

Project 1 of 8 in a hands-on system design learning series.

## What it does

- `POST /api/shorten` with a JSON body `{"url": "https://example.com/some/long/path"}`
  returns a short code and short URL.
- `GET /{code}` redirects to the original long URL.

## Design notes (V2)

- **Storage**: SQLite database (`urlshortener.db`, created automatically on
  first run, ignored by git). Data survives server restarts. Uses
  [modernc.org/sqlite](https://modernc.org/sqlite), a pure-Go SQLite driver,
  so no C compiler / cgo setup is required.
- **Code generation**: short codes are the row's auto-incrementing SQLite `id`,
  base62-encoded (`0-9A-Za-z`). No separate in-memory counter is needed — the
  database is the single source of truth, and encoding is reversible
  (`decodeBase62`), so a `GET /{code}` just decodes back to a row id and looks
  it up directly.
- **Concurrency**: `database/sql`'s connection pool handles concurrent
  requests safely.

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

- Migrate from SQLite to Postgres (tradeoffs: concurrent writers, running as
  a separate service, replication).
- Add expiration / TTL for links.
- Add click analytics (count redirects per code).
- Add rate limiting on the `/api/shorten` endpoint (ties into Project 2).

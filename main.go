package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Store holds the URL mappings in memory.
// A real system would swap this for a database (that's a later project milestone).
type Store struct {
	mu      sync.RWMutex
	urls    map[string]string // short code -> long URL
	counter uint64
}

func NewStore() *Store {
	return &Store{urls: make(map[string]string)}
}

// Save assigns the next counter value a short code and stores the mapping.
func (s *Store) Save(longURL string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	code := encodeBase62(s.counter)
	s.urls[code] = longURL
	return code
}

func (s *Store) Get(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	longURL, ok := s.urls[code]
	return longURL, ok
}

// encodeBase62 turns a counter into a short, URL-safe string.
// e.g. 1 -> "1", 62 -> "10", 125 -> "21"
func encodeBase62(n uint64) string {
	if n == 0 {
		return string(base62Chars[0])
	}
	var sb strings.Builder
	for n > 0 {
		remainder := n % 62
		sb.WriteByte(base62Chars[remainder])
		n /= 62
	}
	// digits came out least-significant-first; reverse them
	encoded := sb.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func main() {
	store := NewStore()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/shorten", func(w http.ResponseWriter, r *http.Request) {
		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.URL) == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}

		code := store.Save(req.URL)
		resp := shortenResponse{
			ShortCode: code,
			ShortURL:  "http://localhost:8080/" + code,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		longURL, ok := store.Get(code)
		if !ok {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, longURL, http.StatusFound)
	})

	log.Println("url-shortener listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

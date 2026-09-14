package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

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

// decodeBase62 reverses encodeBase62, turning a short code back into its counter value.
func decodeBase62(s string) (uint64, error) {
	var n uint64
	for _, c := range s {
		idx := strings.IndexRune(base62Chars, c)
		if idx == -1 {
			return 0, fmt.Errorf("invalid character %q in code", c)
		}
		n = n*62 + uint64(idx)
	}
	return n, nil
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func main() {
	store, err := NewSQLiteStore("urlshortener.db")
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}
	defer store.Close()

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

		code, err := store.Save(req.URL)
		if err != nil {
			log.Printf("save error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		resp := shortenResponse{
			ShortCode: code,
			ShortURL:  "http://localhost:8080/" + code,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		longURL, found, err := store.Get(code)
		if err != nil {
			log.Printf("get error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !found {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, longURL, http.StatusFound)
	})

	log.Println("url-shortener listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

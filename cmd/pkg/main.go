package main

import (
	"log"
	"net/http"
	"strconv"
)

func applyRateLimiter(next http.Handler, rateLimiter *model) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.URL.Query().Get("userID"))
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid user id"))
			return
		}

		statusCode := rateLimiter.Get(userID)
		if statusCode >= 300 {
			w.WriteHeader(statusCode)
			w.Write([]byte(http.StatusText(statusCode)))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	rateLimiter := NewModel(10)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: applyRateLimiter(mux, rateLimiter),
	}

	log.Println("starting server on port 8080")
	err := server.ListenAndServe()
	log.Fatal(err)
}

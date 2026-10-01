package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"example.com/captcha-game-signup/internal/game"
	"example.com/captcha-game-signup/internal/infrai"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = ":8080"
	}

	pipeline := game.NewPipeline(infrai.New(apiKey))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", signupHandler(pipeline))

	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("signup gate listening on %s", address)
	log.Fatal(server.ListenAndServe())
}

func signupHandler(pipeline *game.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var registration game.Registration
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&registration); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		registration.IP = clientIP(r)

		result, err := pipeline.Signup(r.Context(), registration)
		if err != nil {
			var inputErr *game.InputError
			if errors.As(err, &inputErr) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": inputErr.Error()})
				return
			}
			writeJSON(w, infrai.Status(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, result)
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

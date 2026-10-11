package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"context"
	"os"

	"github.com/Shannon-Reed/SIP-Routing-Flow/internal/database"
	"github.com/Shannon-Reed/SIP-Routing-Flow/internal/routing"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	// Retrieve database connection string
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is missing")
	}

	// Connect to Neon PostgreSQL
	db, err := database.Connect(context.Background(), databaseURL)

	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	defer db.Close()

	log.Println("Successfully connected to Neon PostgreSQL")

	router := routing.NewRouter(db)
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "sip-routing-flow",
		})
	})

	// SIP routing endpoint
	mux.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var call routing.CallRequest

		decoder := json.NewDecoder(
			http.MaxBytesReader(w, r.Body, 1<<20),
		)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&call); err != nil {
			http.Error(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		// Reject trailing JSON values.
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			http.Error(w, "Expected a single JSON object", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		response, err := router.FindRoute(ctx, call)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
			return
		}

		log.Printf("Call %s routed to %s", call.CallID, response.Route)
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("SIP Routing Flow API listening on :8080")
	log.Fatal(server.ListenAndServe())
}

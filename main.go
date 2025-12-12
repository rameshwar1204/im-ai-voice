package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Initialize storage directories
	if err := InitStorageDirs(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Println("Storage directories initialized")

	// Initialize AI client (Gemini)
	ai, err := NewAIClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize AI client: %v", err)
	}
	defer ai.Close()
	log.Println("AI client initialized (Gemini)")

	// Initialize service
	svc := NewService(ai)

	// Create cancellable context for background tasks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start aggregation ticker (runs periodically)
	svc.StartAggregationTicker(ctx)

	// Initialize router
	router := NewRouter(svc)
	router.RegisterRoutes()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		os.Exit(0)
	}()

	// Print startup info
	fmt.Println("=========================================")
	fmt.Println("  IndiaMART Voice AI Analysis Server")
	fmt.Println("=========================================")
	fmt.Printf("Server running on http://localhost%s\n", SERVER_LISTEN_ADDR)
	fmt.Println()
	fmt.Println("API Endpoints:")
	fmt.Println("  POST /ingest              - Ingest call transcript")
	fmt.Println("  POST /analyze             - Analyze transcript directly")
	fmt.Println("  POST /analyze/trigger     - Process all unprocessed")
	fmt.Println("  GET  /calls/{id}          - Get call analysis")
	fmt.Println("  GET  /aggregates          - List aggregates")
	fmt.Println("  GET  /aggregates/{date}   - Get daily aggregate")
	fmt.Println("  POST /aggregates/trigger  - Run aggregation")
	fmt.Println("  GET  /tickets             - List ticket dates")
	fmt.Println("  GET  /tickets/{date}      - Get tickets for date")
	fmt.Println("  GET  /dashboard?date=...  - Get daily dashboard")
	fmt.Println("  GET  /health              - Health check")
	fmt.Println()
	fmt.Printf("Using LLM: Google Gemini (%s)\n", GeminiModel)
	fmt.Printf("Data directory: %s\n", STORAGE_BASE)
	fmt.Println("=========================================")

	// Start HTTP server
	if err := http.ListenAndServe(SERVER_LISTEN_ADDR, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

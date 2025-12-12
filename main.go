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

	// Initialize MongoDB (optional - if MONGODB_URI is set)
	if err := InitMongoDB(); err != nil {
		log.Printf("Warning: MongoDB initialization failed: %v", err)
		log.Println("Continuing without MongoDB sync...")
	}
	if MongoDB != nil && MongoDB.enabled {
		defer MongoDB.Close()
	}

	// Initialize AI client (Gemini)
	ai, err := NewAIClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize AI client: %v", err)
	}
	defer ai.Close()
	log.Println("AI client initialized (Gemini)")

	// Initialize service
	svc := NewService(ai)

	// Create cancellable context for shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start transcript watcher (event-driven analysis)
	watcher := NewTranscriptWatcher(svc, TRANSCRIPTS_DIR)
	watcher.Start()
	defer watcher.Stop()

	// Initialize router
	router := NewRouter(svc)
	router.RegisterRoutes()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		watcher.Stop()
		cancel()
		os.Exit(0)
	}()

	// Print startup info
	fmt.Println("=========================================")
	fmt.Println("  IndiaMART Voice AI Analysis Server")
	fmt.Println("=========================================")
	fmt.Printf("Server running on http://localhost%s\n", SERVER_LISTEN_ADDR)
	fmt.Println()
	fmt.Println("🤖 EVENT-DRIVEN AUTOMATED FLOW:")
	fmt.Println("   1. New transcript in data/transcripts/ → Auto-analyze")
	fmt.Println("   2. Seller profile updated (data/profiles/seller_{id}.json)")
	fmt.Println("   3. After 10 new analyses → Auto-aggregate + tickets")
	fmt.Println()

	// MongoDB status
	if MongoDB != nil && MongoDB.enabled {
		fmt.Println("💾 MongoDB: ✅ PRIMARY STORAGE")
		fmt.Printf("   Database: %s\n", DB_NAME)
		fmt.Println("   Collections: seller_profiles, call_analyses, tickets, daily_aggregates")
		fmt.Println("   Mode: MongoDB-first (no local files)")
	} else {
		fmt.Println("💾 MongoDB: ❌ DISABLED (set MONGODB_URI to enable)")
		fmt.Println("   Mode: Local JSON files only")
	}
	fmt.Println()

	fmt.Println("API Endpoints:")
	fmt.Println("  POST /ingest              - Ingest call transcript")
	fmt.Println("  POST /analyze             - Analyze transcript directly")
	fmt.Println("  POST /analyze/trigger     - Process all unprocessed")
	fmt.Println("  GET  /calls/{id}          - Get call analysis")
	fmt.Println()
	fmt.Println("  📊 SELLER PROFILES (Dashboard-Ready):")
	fmt.Println("  GET  /sellers             - List all sellers with status")
	fmt.Println("  GET  /sellers/{gluser_id} - Get full seller profile")
	fmt.Println()
	fmt.Println("  GET  /aggregates          - List aggregates")
	fmt.Println("  GET  /aggregates/{date}   - Get daily aggregate")
	fmt.Println("  POST /aggregates/trigger  - Run aggregation manually")
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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// toBsonM converts any struct to bson.M using JSON tags
// This ensures field names match JSON tags (lowercase with underscores)
func toBsonM(v interface{}) (bson.M, error) {
	// Convert to JSON first (uses json tags)
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// Convert JSON to bson.M
	var doc bson.M
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// MongoDB collections
const (
	DB_NAME               = "indiamart_voice"
	COLLECTION_PROFILES   = "seller_profiles"
	COLLECTION_ANALYSES   = "call_analyses"
	COLLECTION_TICKETS    = "tickets"
	COLLECTION_AGGREGATES = "daily_aggregates"
)

// MongoClient wraps the MongoDB client
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	enabled  bool
}

// Global MongoDB client instance
var MongoDB *MongoClient

// InitMongoDB initializes the MongoDB connection
// Set MONGODB_URI environment variable to enable
func InitMongoDB() error {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Println("⚠️  MONGODB_URI not set - MongoDB sync disabled")
		log.Println("   Data will only be saved to local JSON files")
		MongoDB = &MongoClient{enabled: false}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(DB_NAME)

	// Create indexes for better query performance
	createIndexes(ctx, database)

	MongoDB = &MongoClient{
		client:   client,
		database: database,
		enabled:  true,
	}

	log.Println("✅ MongoDB connected successfully")
	log.Printf("   Database: %s", DB_NAME)
	return nil
}

// createIndexes creates indexes for collections
func createIndexes(ctx context.Context, db *mongo.Database) {
	// Seller profiles - index on gluser_id
	db.Collection(COLLECTION_PROFILES).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "gluser_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Call analyses - index on call_id and seller_id
	db.Collection(COLLECTION_ANALYSES).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "call_id", Value: 1}}},
		{Keys: bson.D{{Key: "seller_id", Value: 1}}},
		{Keys: bson.D{{Key: "timestamp", Value: -1}}},
	})

	// Tickets - index on date and status
	db.Collection(COLLECTION_TICKETS).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "ticket_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "date", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "feature_bucket", Value: 1}}},
	})

	// Aggregates - index on date
	db.Collection(COLLECTION_AGGREGATES).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "date", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

// Close closes the MongoDB connection
func (m *MongoClient) Close() error {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.client.Disconnect(ctx)
	}
	return nil
}

// ==================== SYNC FUNCTIONS ====================
// These functions push data to MongoDB (called alongside local file saves)

// SyncSellerProfile pushes seller profile to MongoDB
func SyncSellerProfile(profile *SellerProfile) {
	if MongoDB == nil || !MongoDB.enabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collection := MongoDB.database.Collection(COLLECTION_PROFILES)

		// Convert to bson.M using JSON tags
		doc, err := toBsonM(profile)
		if err != nil {
			log.Printf("⚠️  MongoDB marshal failed for profile %s: %v", profile.GluserID, err)
			return
		}

		// Upsert - update if exists, insert if not
		filter := bson.M{"gluser_id": profile.GluserID}
		opts := options.Replace().SetUpsert(true)

		_, err = collection.ReplaceOne(ctx, filter, doc, opts)
		if err != nil {
			log.Printf("⚠️  MongoDB sync failed for profile %s: %v", profile.GluserID, err)
		} else {
			log.Printf("   📤 Synced profile to MongoDB: %s", profile.GluserID)
		}
	}()
}

// SyncAnalysis pushes call analysis to MongoDB
func SyncAnalysis(analysis *AnalysisResult) {
	if MongoDB == nil || !MongoDB.enabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collection := MongoDB.database.Collection(COLLECTION_ANALYSES)

		// Convert to bson.M using JSON tags
		doc, err := toBsonM(analysis)
		if err != nil {
			log.Printf("⚠️  MongoDB marshal failed for analysis %s: %v", analysis.CallID, err)
			return
		}

		// Upsert by call_id
		filter := bson.M{"call_id": analysis.CallID}
		opts := options.Replace().SetUpsert(true)

		_, err = collection.ReplaceOne(ctx, filter, doc, opts)
		if err != nil {
			log.Printf("⚠️  MongoDB sync failed for analysis %s: %v", analysis.CallID, err)
		}
	}()
}

// SyncTicket pushes a ticket to MongoDB
func SyncTicket(ticket *Ticket) {
	if MongoDB == nil || !MongoDB.enabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collection := MongoDB.database.Collection(COLLECTION_TICKETS)

		// Convert to bson.M using JSON tags
		doc, err := toBsonM(ticket)
		if err != nil {
			log.Printf("⚠️  MongoDB marshal failed for ticket %s: %v", ticket.TicketID, err)
			return
		}

		// Upsert by ticket_id
		filter := bson.M{"ticket_id": ticket.TicketID}
		opts := options.Replace().SetUpsert(true)

		_, err = collection.ReplaceOne(ctx, filter, doc, opts)
		if err != nil {
			log.Printf("⚠️  MongoDB sync failed for ticket %s: %v", ticket.TicketID, err)
		} else {
			log.Printf("   📤 Synced ticket to MongoDB: %s", ticket.TicketID)
		}
	}()
}

// SyncAggregate pushes daily aggregate to MongoDB
func SyncAggregate(aggregate *DailyAggregate) {
	if MongoDB == nil || !MongoDB.enabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collection := MongoDB.database.Collection(COLLECTION_AGGREGATES)

		// Convert to bson.M using JSON tags
		doc, err := toBsonM(aggregate)
		if err != nil {
			log.Printf("⚠️  MongoDB marshal failed for aggregate %s: %v", aggregate.Date, err)
			return
		}

		// Upsert by date
		filter := bson.M{"date": aggregate.Date}
		opts := options.Replace().SetUpsert(true)

		_, err = collection.ReplaceOne(ctx, filter, doc, opts)
		if err != nil {
			log.Printf("⚠️  MongoDB sync failed for aggregate %s: %v", aggregate.Date, err)
		} else {
			log.Printf("   📤 Synced aggregate to MongoDB: %s", aggregate.Date)
		}
	}()
}

// ==================== BULK SYNC (for initial load) ====================

// SyncAllProfilesToMongo syncs all existing profiles to MongoDB
func SyncAllProfilesToMongo() error {
	if MongoDB == nil || !MongoDB.enabled {
		return fmt.Errorf("MongoDB not enabled")
	}

	ids, err := ListSellerProfiles()
	if err != nil {
		return err
	}

	log.Printf("📤 Syncing %d seller profiles to MongoDB...", len(ids))

	for _, id := range ids {
		profile, err := LoadSellerProfile(id)
		if err != nil || profile == nil {
			continue
		}
		SyncSellerProfile(profile)
	}

	return nil
}

// SyncAllTicketsToMongo syncs all existing tickets to MongoDB
func SyncAllTicketsToMongo() error {
	if MongoDB == nil || !MongoDB.enabled {
		return fmt.Errorf("MongoDB not enabled")
	}

	dates, err := ListTicketDates()
	if err != nil {
		return err
	}

	log.Printf("📤 Syncing tickets for %d dates to MongoDB...", len(dates))

	for _, date := range dates {
		tickets, err := LoadTicketsForDate(date)
		if err != nil {
			continue
		}
		for _, ticket := range tickets {
			SyncTicket(&ticket)
		}
	}

	return nil
}

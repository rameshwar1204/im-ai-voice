package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
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

// ==================== READ FUNCTIONS (MongoDB-first) ====================

// GetSellerProfileFromMongo loads a seller profile from MongoDB
func GetSellerProfileFromMongo(gluserID string) (*SellerProfile, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_PROFILES)
	filter := bson.M{"gluser_id": gluserID}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found
		}
		return nil, err
	}

	// Convert bson.M to SellerProfile via JSON
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	var profile SellerProfile
	if err := json.Unmarshal(jsonBytes, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetAllAnalysesForDateFromMongo loads all analyses for a date from MongoDB
func GetAllAnalysesForDateFromMongo(date string) ([]AnalysisResult, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_ANALYSES)

	// Parse date to create time range
	startTime, _ := time.Parse("2006-01-02", date)
	endTime := startTime.Add(24 * time.Hour)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startTime.Format(time.RFC3339),
			"$lt":  endTime.Format(time.RFC3339),
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []AnalysisResult
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// Convert to AnalysisResult via JSON
		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			continue
		}

		var ar AnalysisResult
		if err := json.Unmarshal(jsonBytes, &ar); err != nil {
			continue
		}
		results = append(results, ar)
	}

	return results, nil
}

// GetAllAnalysesFromMongo loads all analyses from MongoDB (for aggregation)
func GetAllAnalysesFromMongo() ([]AnalysisResult, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_ANALYSES)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []AnalysisResult
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			continue
		}

		var ar AnalysisResult
		if err := json.Unmarshal(jsonBytes, &ar); err != nil {
			continue
		}
		results = append(results, ar)
	}

	return results, nil
}

// CountAnalysesFromMongo returns count of all analyses in MongoDB
func CountAnalysesFromMongo() (int64, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return 0, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_ANALYSES)
	return collection.CountDocuments(ctx, bson.M{})
}

// GetAnalysisFromMongo loads a single analysis by call_id
func GetAnalysisFromMongo(callID string) (*AnalysisResult, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_ANALYSES)
	filter := bson.M{"call_id": callID}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	var ar AnalysisResult
	if err := json.Unmarshal(jsonBytes, &ar); err != nil {
		return nil, err
	}

	return &ar, nil
}

// AnalysisExistsInMongo checks if an analysis exists in MongoDB
func AnalysisExistsInMongo(callID string) bool {
	if MongoDB == nil || !MongoDB.enabled {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_ANALYSES)
	count, err := collection.CountDocuments(ctx, bson.M{"call_id": callID})
	return err == nil && count > 0
}

// GetAggregateFromMongo loads a daily aggregate from MongoDB
func GetAggregateFromMongo(date string) (*DailyAggregate, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_AGGREGATES)
	filter := bson.M{"date": date}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	var agg DailyAggregate
	if err := json.Unmarshal(jsonBytes, &agg); err != nil {
		return nil, err
	}

	return &agg, nil
}

// GetTicketsForDateFromMongo loads all tickets for a date from MongoDB
func GetTicketsForDateFromMongo(date string) ([]Ticket, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_TICKETS)
	filter := bson.M{"date": date}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tickets []Ticket
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			continue
		}

		var ticket Ticket
		if err := json.Unmarshal(jsonBytes, &ticket); err != nil {
			continue
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

// ListAllSellerIDsFromMongo returns all seller IDs from MongoDB
func ListAllSellerIDsFromMongo() ([]string, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_PROFILES)

	// Use distinct to get unique gluser_ids
	ids, err := collection.Distinct(ctx, "gluser_id", bson.M{})
	if err != nil {
		return nil, err
	}

	var result []string
	for _, id := range ids {
		if s, ok := id.(string); ok {
			result = append(result, s)
		}
	}

	return result, nil
}

// ListAggregateDatesFromMongo returns all aggregate dates from MongoDB
func ListAggregateDatesFromMongo() ([]string, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_AGGREGATES)

	dates, err := collection.Distinct(ctx, "date", bson.M{})
	if err != nil {
		return nil, err
	}

	var result []string
	for _, d := range dates {
		if s, ok := d.(string); ok {
			result = append(result, s)
		}
	}

	// Sort descending
	sort.Sort(sort.Reverse(sort.StringSlice(result)))
	return result, nil
}

// ListTicketDatesFromMongo returns all unique ticket dates from MongoDB
func ListTicketDatesFromMongo() ([]string, error) {
	if MongoDB == nil || !MongoDB.enabled {
		return nil, fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_TICKETS)

	dates, err := collection.Distinct(ctx, "date", bson.M{})
	if err != nil {
		return nil, err
	}

	var result []string
	for _, d := range dates {
		if s, ok := d.(string); ok {
			result = append(result, s)
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(result)))
	return result, nil
}

// IsMongoEnabled returns true if MongoDB is connected and enabled
func IsMongoEnabled() bool {
	return MongoDB != nil && MongoDB.enabled
}

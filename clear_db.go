//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uri := "mongodb+srv://rameshwar:rameshwar123@cluster0.zaauneb.mongodb.net/?appName=Cluster0"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("indiamart_voice")

	collections := []string{"seller_profiles", "call_analyses", "tickets", "daily_aggregates"}
	for _, coll := range collections {
		result, err := db.Collection(coll).DeleteMany(ctx, bson.M{})
		if err != nil {
			log.Printf("Error clearing %s: %v", coll, err)
		} else {
			fmt.Printf("Cleared %s: %d documents deleted\n", coll, result.DeletedCount)
		}
	}

	fmt.Println("\n✅ All MongoDB collections cleared!")
}

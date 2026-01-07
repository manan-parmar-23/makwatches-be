package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Order represents the order structure
type Order struct {
	ID          primitive.ObjectID `bson:"_id"`
	OrderNumber string             `bson:"order_number"`
	CreatedAt   time.Time          `bson:"created_at"`
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get MongoDB URI from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = os.Getenv("MONGODB_URI")
	}
	if mongoURI == "" {
		log.Fatal("MONGO_URI or MONGODB_URI environment variable is required")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully!")

	// Get orders collection
	db := client.Database("makwatches")
	ordersCollection := db.Collection("orders")

	// Find all orders without order_number or with empty order_number
	filter := bson.M{
		"$or": []bson.M{
			{"order_number": bson.M{"$exists": false}},
			{"order_number": ""},
		},
	}

	cursor, err := ordersCollection.Find(ctx, filter)
	if err != nil {
		log.Fatalf("Failed to find orders: %v", err)
	}
	defer cursor.Close(ctx)

	var ordersToFix []Order
	if err := cursor.All(ctx, &ordersToFix); err != nil {
		log.Fatalf("Failed to decode orders: %v", err)
	}

	if len(ordersToFix) == 0 {
		log.Println("No orders need fixing - all orders already have order numbers!")
		return
	}

	log.Printf("Found %d orders without order numbers. Fixing...", len(ordersToFix))

	// Sort orders by creation date
	sort.Slice(ordersToFix, func(i, j int) bool {
		return ordersToFix[i].CreatedAt.Before(ordersToFix[j].CreatedAt)
	})

	// Group orders by date and generate order numbers
	dateCounters := make(map[string]int)

	for _, order := range ordersToFix {
		dateStr := order.CreatedAt.Format("20060102")
		dateCounters[dateStr]++

		// Check if there are existing orders with order numbers for this date
		existingCount, _ := ordersCollection.CountDocuments(ctx, bson.M{
			"order_number": bson.M{"$regex": fmt.Sprintf("^MAK-%s-", dateStr)},
		})

		// Generate order number
		sequenceNum := int(existingCount) + dateCounters[dateStr]
		orderNumber := fmt.Sprintf("MAK-%s-%03d", dateStr, sequenceNum)

		// Update the order
		_, err := ordersCollection.UpdateOne(
			ctx,
			bson.M{"_id": order.ID},
			bson.M{"$set": bson.M{"order_number": orderNumber}},
		)
		if err != nil {
			log.Printf("Failed to update order %s: %v", order.ID.Hex(), err)
			continue
		}

		log.Printf("Updated order %s -> %s (created: %s)",
			order.ID.Hex(),
			orderNumber,
			order.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	log.Println("Done! All orders now have order numbers.")
}

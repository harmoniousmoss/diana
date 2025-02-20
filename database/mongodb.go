package database

import (
	"context"
	"fmt"
	"log"
	"myfiberproject/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// ConnectMongoDB initializes the MongoDB connection
func ConnectMongoDB() error {
	log.Println("Attempting to connect to MongoDB...")

	mongoURI := config.GetEnv("MONGO_URI", "")
	if mongoURI == "" {
		log.Println("MONGO_URI environment variable is not set.")
		return fmt.Errorf("MONGO_URI environment variable is not set")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error

	MongoClient, err = connect(clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully!")

	return nil
}

// connect attempts to establish a MongoDB connection with retries
func connect(clientOptions *options.ClientOptions) (*mongo.Client, error) {
	var client *mongo.Client
	var err error

	for attempt := 1; attempt <= 5; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			err = client.Ping(ctx, nil)
			if err == nil {
				cancel()
				return client, nil
			}
		}

		log.Printf("Failed to connect to MongoDB: %v, retrying in %d seconds", err, attempt*2)
		time.Sleep(time.Duration(attempt*2) * time.Second)
		cancel()
	}

	return nil, fmt.Errorf("failed to connect to MongoDB after multiple attempts: %v", err)
}

// GetMongoClient returns the MongoDB client instance
func GetMongoClient() *mongo.Client {
	return MongoClient
}

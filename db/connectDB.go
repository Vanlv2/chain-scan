// db/db.go
package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

// ConnectMongoDB creates and returns a MongoDB collection reference
func ConnectMongoDB(databaseName string, collectionName string) *mongo.Collection {
	// Check if the client is already initialized
	if mongoClient == nil {
		var err error
		// Create MongoDB client
		mongoClient, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			log.Fatal("Error creating MongoDB client: ", err)
			return nil
		}

		// Connect to MongoDB
		err = mongoClient.Connect(context.Background())
		if err != nil {
			log.Fatal("Error connecting to MongoDB: ", err)
			return nil
		}
	}

	// Return the MongoDB collection
	collection := mongoClient.Database(databaseName).Collection(collectionName)
	return collection
}

// DisconnectMongoDB disconnects the MongoDB client
func DisconnectMongoDB() {
	if mongoClient != nil {
		err := mongoClient.Disconnect(context.Background())
		if err != nil {
			log.Fatal("Error disconnecting from MongoDB: ", err)
		}
	}
}

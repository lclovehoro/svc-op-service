package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/klog/v2"
)

var MongoClient *mongoClient

type mongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func init() {
	if err := NewMongoDBClient(); err != nil {
		klog.Errorf("Failed to connect to MongoDB: %v", err)
	}
}

func NewMongoDBClient() error {
	connectionString := "mongodb://mongo.betawm.beta:27017"
	databaseString := "operation_db"
	clientoptions := options.Client().ApplyURI(connectionString)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientoptions)
	if err != nil {
		klog.Fatal(err)
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		klog.Fatal(err)
		return err
	}

	klog.V(1).Info("Connected to MongoDB!")

	database := client.Database(databaseString)
	MongoClient = &mongoClient{
		Client:   client,
		Database: database,
	}
	return nil
}

package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.uber.org/zap"
)

func connectMongo(connectionString string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(connectionString).SetDirect(true)
	clientOptions.Monitor = otelmongo.NewMonitor()
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		return nil, err
	}

	// Check the connection
	if err = client.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	zap.L().Info("MongoClient connected")

	return client, nil
}

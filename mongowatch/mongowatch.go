package mongowatch

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
)

func init() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	database = client.Database("test")
	collection = database.Collection("demo")
}

func GetWatch() *mongo.ChangeStream {
	stream, err := collection.Watch(context.TODO(), mongo.Pipeline{})
	if err != nil {
		panic(err)
	}
	return stream
}

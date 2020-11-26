package watch

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
)

func init() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("localhost:27017"))
	if err != nil {
		panic(err)
	}
	database = client.Database("test")
	collection = database.Collection("demo")
}

func Watch() {
	stream, err := collection.Watch(context.TODO(), mongo.Pipeline{})
	if err != nil {
		panic(err)
	}
	defer stream.Close(context.TODO())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for stream.Next(context.TODO()) {
			var data bson.M
			if err := stream.Decode(&data); err != nil {
				panic(err)
			}
			fmt.Printf("%v\n", data)
		}
	}()
	wg.Wait()
}

package repository

import (
	"context"
	"demo/common"
	"demo/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	serverURI      string
	databaseName   string
	collectionName string
)
var mongoRepo *MongoRecordRepo

type MongoRecordRepo struct {
	client *mongo.Client
}

func GetMongoRepo() *MongoRecordRepo {
	return mongoRepo
}
func init() {
	serverURI = common.GetEnv("MONGO_URI", "mongodb://localhost:27017")
	databaseName = common.GetEnv("DATABASE", "test")
	collectionName = common.GetEnv("COLLECTION", "demo")
	// Set client options
	clientOptions := options.Client().ApplyURI(serverURI)
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	mongoRepo = &MongoRecordRepo{client: client}
}

func (repo *MongoRecordRepo) InsertRecord(record models.Record) (Id string, err error) {
	collection := repo.client.Database(databaseName).Collection(collectionName)
	res, err := collection.InsertOne(context.TODO(), &record)
	Id = res.InsertedID.(string)
	return
}

func (repo *MongoRecordRepo) ReplaceRecord(Id string, newRecord models.Record) (createNew bool, err error) {
	collection := repo.client.Database(databaseName).Collection(collectionName)
	option := options.FindOneAndReplace()
	option.SetUpsert(true)
	res := collection.FindOneAndReplace(context.TODO(), bson.D{{"_id", Id}}, &newRecord, option)
	if err = res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			createNew = true
			err = nil
			return
		}
	}
	return
}

func (repo *MongoRecordRepo) ReplaceRecordIfMatch(etag, Id string, newRecord models.Record) (err error) {
	collection := repo.client.Database(databaseName).Collection(collectionName)
	// option := options.FindOneAndReplace()
	// option.SetUpsert(true)
	res := collection.FindOneAndReplace(context.TODO(), bson.D{{"_id", Id}, {"etag", etag}}, &newRecord)
	if err = res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			_, err2 := repo.GetRecord(Id)
			if err2 == nil {
				err = common.ErrorStale
			} else if err2 == mongo.ErrNoDocuments {
				err = err2
			}
			return
		}
	}
	return
}

func (repo *MongoRecordRepo) GetRecord(Id string) (record models.Record, err error) {
	collection := repo.client.Database(databaseName).Collection(collectionName)
	res := collection.FindOne(context.TODO(), bson.D{{"_id", Id}})
	if err = res.Err(); err != nil {
		return
	}
	err = res.Decode(&record)
	return
}
func (repo *MongoRecordRepo) DeleteRecord(Id string) (err error) {
	collection := repo.client.Database(databaseName).Collection(collectionName)
	res, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", Id}})
	if err != nil {
		return
	}
	if res.DeletedCount == 0 {
		err = common.ErrorNotFound
	}
	return
}
func (repo *MongoRecordRepo) DeleteRecordIfMatch(etag, Id string) (err error) {
	collection := repo.client.Database(databaseName).Collection(collectionName)
	res, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", Id}, {"etag", etag}})
	if err != nil {
		return
	}
	if res.DeletedCount == 0 {
		_, err2 := repo.GetRecord(Id)
		if err2 == nil {
			err = common.ErrorStale
		} else if err2 == mongo.ErrNoDocuments {
			err = common.ErrorNotFound
		}
		return
	}
	return
}

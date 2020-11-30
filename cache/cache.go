package cache

import (
	"container/list"
	"context"
	"demo/models"
	"demo/mongowatch"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
)

type Cache interface {
	Get(id string) models.Record
	Put(record models.Record)
	Delete(id string)
}

func CacheSelf() Cache {
	return recordCache
}

type cacheImp struct {
	sync.Mutex
	dict map[string]*list.Element
	l    *list.List
	size int
}

var recordCache Cache

func init() {
	recordCache = &cacheImp{
		dict: make(map[string]*list.Element),
		l:    list.New(),
		size: 10,
	}
	stream := mongowatch.GetWatch()
	go func() {
		defer stream.Close(context.TODO())
		for stream.Next(context.TODO()) {
			var data bson.M
			if err := stream.Decode(&data); err != nil {
				log.Fatal(err)
			}
			update(data)
		}
	}()
}

func update(data bson.M) {
	switch data["operationType"] {
	case "insert", "update", "replace":
		{
			var record models.Record
			bsonBytes, err := bson.Marshal(data["fullDocument"])
			if err != nil {
				log.Fatal(err)
			}
			if err := bson.Unmarshal(bsonBytes, &record); err != nil {
				log.Fatal(err)
			}
			CacheSelf().Put(record)
		}
	case "delete":
		{
			id := data["documentKey"].(bson.M)["_id"].(string)
			CacheSelf().Delete(id)
		}
	}

}

func (c *cacheImp) Get(id string) models.Record {
	c.Lock()
	r := c.dict[id].Value.(*models.Record)
	c.Unlock()
	return *r
}

func (c *cacheImp) printList() {
	head := c.l.Front()
	if head != nil {
		for head != nil {
			log.Println(*head.Value.(*models.Record))
			head = head.Next()
		}
	}
	log.Println("-------------")
}

func (c *cacheImp) Put(record models.Record) {
	c.Lock()
	defer c.Unlock()
	v, ok := c.dict[record.Id]
	if ok {
		v.Value = &record
		c.l.MoveToFront(v)
	} else {
		if c.l.Len() >= c.size {
			c.l.Remove(c.l.Back())
		}
		c.l.PushFront(&record)
		c.dict[record.Id] = c.l.Front()
	}

	c.printList()
}

func (c *cacheImp) Delete(id string) {
	c.Lock()
	defer c.Unlock()
	v, ok := c.dict[id]
	if !ok {
		return
	}
	c.l.Remove(v)
	delete(c.dict, id)
	c.printList()
}

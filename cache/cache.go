package cache

import (
	"container/heap"
	"context"
	"demo/models"
	"demo/mongowatch"
	"log"
	"sync"
	"time"

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

type recordHeap []*models.Record

type cacheImp struct {
	sync.Mutex
	dict map[string]*models.Record
	tree recordHeap
	size int
}

var recordCache Cache

func init() {
	recordCache = &cacheImp{
		dict: make(map[string]*models.Record),
		tree: make([]*models.Record, 0),
		size: 1000000,
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
			record.Id = data["fullDocument"].(bson.M)["_id"].(string)
			record.Data = data["fullDocument"].(bson.M)["data"]
			record.LastModified = data["fullDocument"].(bson.M)["lastmodified"].(time.Time)
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
	r := c.dict[id]
	c.Unlock()
	return *r
}

func (c *cacheImp) Put(record models.Record) {
	c.Lock()
	defer c.Unlock()
	v, ok := c.dict[record.Id]
	if ok {
		*v = record
	} else {
		if len(c.tree) >= c.size {
			v := c.tree[0]
			delete(c.dict, v.Id)
			*v = record
			heap.Fix(&c.tree, 0)
		} else {
			c.dict[record.Id] = &record
			heap.Push(&c.tree, &record)
		}

	}
	log.Println(record)
}

func (c *cacheImp) Delete(id string) {
	c.Lock()
	defer c.Unlock()
	v, ok := c.dict[id]
	if !ok {
		return
	}
	v.Data = nil
}

func (h recordHeap) Len() int {
	return len(h)
}

func (h recordHeap) Less(i, j int) bool {
	return h[i].LastModified.After(h[j].LastModified)
}

func (h recordHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *recordHeap) Push(e interface{}) {
	*h = append(*h, e.(*models.Record))
}

func (h *recordHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

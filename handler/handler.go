package handler

import (
	"demo/cache"
	"demo/common"
	"demo/models"
	"demo/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

var loc *time.Location
var timeFormat string
var recordCache cache.Cache

func init() {
	loc, _ = time.LoadLocation("UTC")
	timeFormat = "Mon, 2 Jan 2006 15:04:05 GMT"
	recordCache = cache.CacheSelf()
}

func InsertRecord(c *gin.Context) {
	var record models.Record
	if err := c.BindJSON(&record); err != nil {
		return
	}
	record.Id = uuid.New().String()
	repo := repository.NewRepo()
	Id, err := repo.InsertRecord(record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusCreated, Id)
	return
}

func DeleteRecord(c *gin.Context) {
	Id := c.Param("Id")
	repo := repository.NewRepo()
	err := repo.DeleteRecord(Id)
	if err != nil {
		if err == common.ErrorNotFound {
			c.JSON(http.StatusNotFound, err)
		} else {
			c.JSON(http.StatusInternalServerError, err)
		}
		return
	}
	c.Status(http.StatusOK)
}

func ReplaceRecord(c *gin.Context) {
	var record models.Record
	if err := c.BindJSON(&record); err != nil {
		return
	}
	Id := c.Param("Id")
	repo := repository.NewRepo()
	var createNew bool
	var err error
	createNew, err = repo.ReplaceRecord(Id, record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if createNew {
		c.Status(http.StatusCreated)
	} else {
		c.Status(http.StatusNoContent)
	}
}

func GetRecord(c *gin.Context) {
	Id := c.Param("Id")
	record, err := recordCache.Get(Id)
	if err != nil {
		repo := repository.NewRepo()
		record, err = repo.GetRecord(Id)
	}
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.Status(http.StatusNotFound)
		} else {
			c.JSON(http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, &record)
}

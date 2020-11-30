package repository

import "demo/models"

type RecordRepository interface {
	InsertRecord(record models.Record) (Id string, err error)
	ReplaceRecord(Id string, record models.Record) (createNew bool, err error)
	ReplaceRecordIfMatch(etag, Id string, record models.Record) (err error)
	GetRecord(Id string) (record models.Record, err error)
	DeleteRecord(Id string) (err error)
	DeleteRecordIfMatch(etag, Id string) (err error)
}

func NewRepo() RecordRepository {
	return GetMongoRepo()
}

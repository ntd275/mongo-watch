package models

import (
	"time"
)

type Record struct {
	Id           string
	Data         interface{}
	LastModified time.Time
}

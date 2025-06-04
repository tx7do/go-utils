package id

import (
	"strings"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewGUIDv4(withHyphen bool) string {
	id := uuid.NewString()
	if !withHyphen {
		id = strings.ReplaceAll(id, "-", "")
	}
	return id
}

func NewShortUUID() string {
	return shortuuid.New()
}

func NewKSUID() string {
	return ksuid.New().String()
}

func NewXID() string {
	return xid.New().String()
}

func NewMongoObjectID() string {
	objID := primitive.NewObjectID()
	return objID.String()
}

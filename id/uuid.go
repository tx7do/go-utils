package id

import (
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewGUIDv4 generates a new GUID/UUID v4 string.
func NewGUIDv4(withHyphen bool) string {
	u := uuid.New()

	if withHyphen {
		return u.String()
	}

	var buf [32]byte
	hex.Encode(buf[:], u[:])
	return string(buf[:])
}

// NewGUIDv7 generates a new GUID/UUID v7 string.
func NewGUIDv7(withHyphen bool) string {
	u, err := uuid.NewV7()
	if err != nil {
		// Fallback to v4 if system clock is unreliable
		return NewGUIDv4(withHyphen)
	}

	if withHyphen {
		return u.String()
	}

	var buf [32]byte
	hex.Encode(buf[:], u[:])
	return string(buf[:])
}

// NewShortUUID generates a new ShortUUID string.
func NewShortUUID() string {
	return shortuuid.New()
}

// NewKSUID generates a new KSUID string.
func NewKSUID() string {
	return ksuid.New().String()
}

// NewXID generates a new XID string.
func NewXID() string {
	return xid.New().String()
}

// NewMongoObjectID generates a new MongoDB ObjectID string.
func NewMongoObjectID() string {
	objID := primitive.NewObjectID()
	return objID.String()
}

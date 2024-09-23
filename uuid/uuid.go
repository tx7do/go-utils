package uuid

import (
	"github.com/google/uuid"

	"github.com/tx7do/go-utils/trans"
)

func ToUuidPtr(str *string) *uuid.UUID {
	var id *uuid.UUID
	if str != nil {
		_id, err := uuid.Parse(*str)
		if err != nil {
			return nil
		}
		id = &_id
	}
	return id
}

func ToUuid(str string) uuid.UUID {
	id, _ := uuid.Parse(str)
	return id
}

func ToStringPtr(id *uuid.UUID) *string {
	var strUUID *string
	if id != nil {
		strUUID = trans.String(id.String())
	}
	return strUUID
}

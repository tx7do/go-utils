package uuid

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUID(t *testing.T) {
	t.Run("ToUuidPtr_NilString", func(t *testing.T) {
		var str *string
		result := ToUuidPtr(str)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("ToUuidPtr_ValidString", func(t *testing.T) {
		str := "550e8400-e29b-41d4-a716-446655440000"
		result := ToUuidPtr(&str)
		if result == nil || result.String() != str {
			t.Errorf("expected %v, got %v", str, result)
		}
	})

	t.Run("ToUuidPtr_InvalidString", func(t *testing.T) {
		str := "invalid-uuid"
		result := ToUuidPtr(&str)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("ToUuid_ValidString", func(t *testing.T) {
		str := "550e8400-e29b-41d4-a716-446655440000"
		result := ToUuid(str)
		if result.String() != str {
			t.Errorf("expected %v, got %v", str, result)
		}
	})

	t.Run("ToUuid_InvalidString", func(t *testing.T) {
		str := "invalid-uuid"
		result := ToUuid(str)
		if result.String() == str {
			t.Errorf("expected invalid UUID, got %v", result)
		}
	})

	t.Run("ToStringPtr_NilUUID", func(t *testing.T) {
		var id *uuid.UUID
		result := ToStringPtr(id)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("ToStringPtr_ValidUUID", func(t *testing.T) {
		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		result := ToStringPtr(&id)
		expected := "550e8400-e29b-41d4-a716-446655440000"
		if result == nil || *result != expected {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})
}

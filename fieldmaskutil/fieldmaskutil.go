package fieldmaskutil

import (
	"fmt"

	"github.com/tx7do/go-utils/stringcase"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Filter keeps the msg fields that are listed in the paths and clears all the rest.
//
// This is a handy wrapper for NestedMask.Filter method.
// If the same paths are used to process multiple proto messages use NestedMask.Filter method directly.
func Filter(msg proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Filter(msg)
}

// Prune clears all the fields listed in paths from the given msg.
//
// This is a handy wrapper for NestedMask.Prune method.
// If the same paths are used to process multiple proto messages use NestedMask.Filter method directly.
func Prune(msg proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Prune(msg)
}

// Overwrite overwrites all the fields listed in paths in the dest msg using values from src msg.
//
// This is a handy wrapper for NestedMask.Overwrite method.
// If the same paths are used to process multiple proto messages use NestedMask.Overwrite method directly.
func Overwrite(src, dest proto.Message, paths []string) {
	NestedMaskFromPaths(paths).Overwrite(src, dest)
}

// Validate checks if all paths are valid for specified message
//
// This is a handy wrapper for NestedMask.Validate method.
// If the same paths are used to process multiple proto messages use NestedMask.Validate method directly.
func Validate(validationModel proto.Message, paths []string) error {
	return NestedMaskFromPaths(paths).Validate(validationModel)
}

// ValidateFieldMask checks if all paths in the given FieldMask are valid for the specified message.
func ValidateFieldMask(msg proto.Message, fm *fieldmaskpb.FieldMask) error {
	if fm == nil {
		return nil
	}

	if !fm.IsValid(msg) {
		return fmt.Errorf("invalid field mask: paths %v are not valid for message type %T",
			fm.GetPaths(), msg)
	}

	return nil
}

// FilterByFieldMask keeps the msg fields that are listed in the given FieldMask and clears all the rest.
func FilterByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error {
	if msg == nil {
		return nil
	}

	if fm == nil {
		return nil
	}

	// normalize and validate field mask
	fm.Normalize()

	if err := ValidateFieldMask(*msg, fm); err != nil {
		return err
	}

	NestedMaskFromPaths(fm.GetPaths()).Filter(*msg)
	return nil
}

// PruneByFieldMask clears all the fields listed in the given FieldMask from the msg.
func PruneByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error {
	if msg == nil {
		return nil
	}

	if fm == nil {
		return nil
	}

	// normalize and validate field mask
	fm.Normalize()

	if err := ValidateFieldMask(*msg, fm); err != nil {
		return err
	}

	NestedMaskFromPaths(fm.GetPaths()).Prune(*msg)
	return nil
}

// OverwriteByFieldMask overwrites all the fields listed in the given FieldMask in the dest msg using values from src msg.
func OverwriteByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error {
	if msg == nil {
		return nil
	}

	if fm == nil {
		return nil
	}

	// normalize and validate field mask
	fm.Normalize()

	if err := ValidateFieldMask(*msg, fm); err != nil {
		return err
	}

	NestedMaskFromPaths(fm.GetPaths()).Overwrite(*msg, *msg)
	return nil
}

// NormalizeFieldMaskPaths normalizes the paths in the given FieldMask to snake_case
func NormalizeFieldMaskPaths(fm *fieldmaskpb.FieldMask) {
	if fm == nil || len(fm.GetPaths()) == 0 {
		return
	}

	paths := make([]string, len(fm.Paths))
	for i, field := range fm.GetPaths() {
		if field == "id_" || field == "_id" {
			field = "id"
		}
		paths[i] = stringcase.ToSnakeCase(field)
	}
}

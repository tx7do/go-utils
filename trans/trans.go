package trans

import (
	"time"

	"github.com/google/uuid"
)

func String(a string) *string {
	return &a
}

func StringValue(a *string) string {
	if a == nil {
		return ""
	}
	return *a
}

func Int(a int) *int {
	return &a
}

func IntValue(a *int) int {
	if a == nil {
		return 0
	}
	return *a
}

func Int8(a int8) *int8 {
	return &a
}

func Int8Value(a *int8) int8 {
	if a == nil {
		return 0
	}
	return *a
}

func Int16(a int16) *int16 {
	return &a
}

func Int16Value(a *int16) int16 {
	if a == nil {
		return 0
	}
	return *a
}

func Int32(a int32) *int32 {
	return &a
}

func Int32Value(a *int32) int32 {
	if a == nil {
		return 0
	}
	return *a
}

func Int64(a int64) *int64 {
	return &a
}

func Int64Value(a *int64) int64 {
	if a == nil {
		return 0
	}
	return *a
}

func Bool(a bool) *bool {
	return &a
}

func BoolValue(a *bool) bool {
	if a == nil {
		return false
	}
	return *a
}

func Uint(a uint) *uint {
	return &a
}

func UintValue(a *uint) uint {
	if a == nil {
		return 0
	}
	return *a
}

func Uint8(a uint8) *uint8 {
	return &a
}

func Uint8Value(a *uint8) uint8 {
	if a == nil {
		return 0
	}
	return *a
}

func Uint16(a uint16) *uint16 {
	return &a
}

func Uint16Value(a *uint16) uint16 {
	if a == nil {
		return 0
	}
	return *a
}

func Uint32(a uint32) *uint32 {
	return &a
}

func Uint32Value(a *uint32) uint32 {
	if a == nil {
		return 0
	}
	return *a
}

func Uint64(a uint64) *uint64 {
	return &a
}

func Uint64Value(a *uint64) uint64 {
	if a == nil {
		return 0
	}
	return *a
}

func Float32(a float32) *float32 {
	return &a
}

func Float32Value(a *float32) float32 {
	if a == nil {
		return 0
	}
	return *a
}

func Float64(a float64) *float64 {
	return &a
}

func Float64Value(a *float64) float64 {
	if a == nil {
		return 0
	}
	return *a
}

func Time(a time.Time) *time.Time {
	return &a
}

func TimeValue(a *time.Time) time.Time {
	if a == nil {
		return time.Now()
	}
	return *a
}

func IntSlice(a []int) []*int {
	if a == nil {
		return nil
	}
	res := make([]*int, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func IntValueSlice(a []*int) []int {
	if a == nil {
		return nil
	}
	res := make([]int, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Int8Slice(a []int8) []*int8 {
	if a == nil {
		return nil
	}
	res := make([]*int8, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Int8ValueSlice(a []*int8) []int8 {
	if a == nil {
		return nil
	}
	res := make([]int8, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Int16Slice(a []int16) []*int16 {
	if a == nil {
		return nil
	}
	res := make([]*int16, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Int16ValueSlice(a []*int16) []int16 {
	if a == nil {
		return nil
	}
	res := make([]int16, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Int32Slice(a []int32) []*int32 {
	if a == nil {
		return nil
	}
	res := make([]*int32, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Int32ValueSlice(a []*int32) []int32 {
	if a == nil {
		return nil
	}
	res := make([]int32, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Int64Slice(a []int64) []*int64 {
	if a == nil {
		return nil
	}
	res := make([]*int64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Int64ValueSlice(a []*int64) []int64 {
	if a == nil {
		return nil
	}
	res := make([]int64, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func UintSlice(a []uint) []*uint {
	if a == nil {
		return nil
	}
	res := make([]*uint, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func UintValueSlice(a []*uint) []uint {
	if a == nil {
		return nil
	}
	res := make([]uint, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Uint8Slice(a []uint8) []*uint8 {
	if a == nil {
		return nil
	}
	res := make([]*uint8, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Uint8ValueSlice(a []*uint8) []uint8 {
	if a == nil {
		return nil
	}
	res := make([]uint8, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Uint16Slice(a []uint16) []*uint16 {
	if a == nil {
		return nil
	}
	res := make([]*uint16, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Uint16ValueSlice(a []*uint16) []uint16 {
	if a == nil {
		return nil
	}
	res := make([]uint16, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Uint32Slice(a []uint32) []*uint32 {
	if a == nil {
		return nil
	}
	res := make([]*uint32, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Uint32ValueSlice(a []*uint32) []uint32 {
	if a == nil {
		return nil
	}
	res := make([]uint32, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Uint64Slice(a []uint64) []*uint64 {
	if a == nil {
		return nil
	}
	res := make([]*uint64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Uint64ValueSlice(a []*uint64) []uint64 {
	if a == nil {
		return nil
	}
	res := make([]uint64, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Float32Slice(a []float32) []*float32 {
	if a == nil {
		return nil
	}
	res := make([]*float32, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Float32ValueSlice(a []*float32) []float32 {
	if a == nil {
		return nil
	}
	res := make([]float32, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func Float64Slice(a []float64) []*float64 {
	if a == nil {
		return nil
	}
	res := make([]*float64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func Float64ValueSlice(a []*float64) []float64 {
	if a == nil {
		return nil
	}
	res := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func StringSlice(a []string) []*string {
	if a == nil {
		return nil
	}
	res := make([]*string, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func StringSliceValue(a []*string) []string {
	if a == nil {
		return nil
	}
	res := make([]string, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

func BoolSlice(a []bool) []*bool {
	if a == nil {
		return nil
	}
	res := make([]*bool, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = &a[i]
	}
	return res
}

func BoolSliceValue(a []*bool) []bool {
	if a == nil {
		return nil
	}
	res := make([]bool, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] != nil {
			res[i] = *a[i]
		}
	}
	return res
}

type mapKeyValueType interface {
	~string | ~bool |
		~int | ~int8 | ~int16 | ~int32 |
		~uint | ~uint8 | ~uint16 | ~uint32 |
		~float32 | ~float64
}

// MapKeys 获取map的键
func MapKeys[TKey mapKeyValueType, TValue mapKeyValueType](source map[TKey]TValue) []TKey {
	var target []TKey
	for k := range source {
		target = append(target, k)
	}
	return target
}

// MapValues 获取map的值
func MapValues[TKey mapKeyValueType, TValue mapKeyValueType](source map[TKey]TValue) []TValue {
	var target []TValue
	for _, v := range source {
		target = append(target, v)
	}
	return target
}

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
		strUUID = Ptr(id.String())
	}
	return strUUID
}

package types

import (
	"reflect"
)

func IsStruct(v any) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.Kind() == reflect.Struct
}

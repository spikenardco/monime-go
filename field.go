package monime

import (
	"encoding/json"
	"errors"
	"reflect"
)

// Field represents an omitted, explicitly null, or assigned PATCH field.
type Field[T any] struct {
	value T
	set   bool
	null  bool
}

// Set returns a Field that encodes value. Nil values encode as JSON null.
func Set[T any](value T) Field[T] {
	if isNil(value) {
		return Null[T]()
	}

	return Field[T]{
		value: value,
		set:   true,
	}
}

// Null returns a Field that encodes as JSON null.
func Null[T any]() Field[T] {
	return Field[T]{
		set:  true,
		null: true,
	}
}

// IsSet reports whether the field is assigned or explicitly null.
func (f Field[T]) IsSet() bool {
	return f.set
}

// IsNull reports whether the field encodes as JSON null.
func (f Field[T]) IsNull() bool {
	return f.null
}

// IsZero reports whether the field is omitted by json:"...,omitzero".
func (f Field[T]) IsZero() bool {
	return !f.set
}

// MarshalJSON encodes an assigned or explicitly null Field.
func (f Field[T]) MarshalJSON() ([]byte, error) {
	if !f.set {
		return nil, errors.New("monime: cannot marshal an omitted field")
	}

	if f.null {
		return []byte("null"), nil
	}

	return json.Marshal(f.value)
}

func isNil[T any](value T) bool {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return true
	}

	for reflected.Kind() == reflect.Interface || reflected.Kind() == reflect.Pointer {
		if reflected.IsNil() {
			return true
		}

		reflected = reflected.Elem()
	}

	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Slice:
		return reflected.IsNil()
	}

	return false
}

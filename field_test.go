package monime

import (
	"encoding/json"
	"testing"
)

func TestField_JSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name: "omitted string",
			value: struct {
				Value Field[string] `json:"value,omitzero"`
			}{},
			want: `{}`,
		},
		{
			name: "string value",
			value: struct {
				Value Field[string] `json:"value,omitzero"`
			}{Value: Set("name")},
			want: `{"value":"name"}`,
		},
		{
			name: "empty string value",
			value: struct {
				Value Field[string] `json:"value,omitzero"`
			}{Value: Set("")},
			want: `{"value":""}`,
		},
		{
			name: "integer zero value",
			value: struct {
				Value Field[int] `json:"value,omitzero"`
			}{Value: Set(0)},
			want: `{"value":0}`,
		},
		{
			name: "false value",
			value: struct {
				Value Field[bool] `json:"value,omitzero"`
			}{Value: Set(false)},
			want: `{"value":false}`,
		},
		{
			name: "explicit null",
			value: struct {
				Value Field[string] `json:"value,omitzero"`
			}{Value: Null[string]()},
			want: `{"value":null}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			encoded, err := json.Marshal(test.value)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(encoded) != test.want {
				t.Errorf("json.Marshal() = %s, want %s", encoded, test.want)
			}
		})
	}
}

func TestSet_NilNormalizesToNull(t *testing.T) {
	t.Parallel()

	var pointer *string
	var metadata map[string]string
	var typedNil any = pointer

	tests := []struct {
		name  string
		field interface {
			IsSet() bool
			IsNull() bool
		}
		value any
	}{
		{
			name:  "pointer",
			field: Set(pointer),
			value: struct {
				Value Field[*string] `json:"value,omitzero"`
			}{Value: Set(pointer)},
		},
		{
			name:  "map",
			field: Set(metadata),
			value: struct {
				Value Field[map[string]string] `json:"value,omitzero"`
			}{Value: Set(metadata)},
		},
		{
			name:  "typed nil interface",
			field: Set(typedNil),
			value: struct {
				Value Field[any] `json:"value,omitzero"`
			}{Value: Set(typedNil)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if !test.field.IsSet() {
				t.Error("IsSet() = false, want true")
			}
			if !test.field.IsNull() {
				t.Error("IsNull() = false, want true")
			}

			encoded, err := json.Marshal(test.value)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(encoded) != `{"value":null}` {
				t.Errorf("json.Marshal() = %s, want {\"value\":null}", encoded)
			}
		})
	}
}

func TestSet_NonNilValueIsNotNull(t *testing.T) {
	t.Parallel()

	value := "name"
	field := Set(&value)
	if !field.IsSet() {
		t.Error("IsSet() = false, want true")
	}
	if field.IsNull() {
		t.Error("IsNull() = true, want false")
	}
}

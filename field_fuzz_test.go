package monime

import (
	"encoding/json"
	"testing"
)

func FuzzFieldJSON(f *testing.F) {
	f.Add("", uint8(0))
	f.Add("name", uint8(1))
	f.Add("name", uint8(2))

	f.Fuzz(func(t *testing.T, value string, state uint8) {
		switch state % 3 {
		case 0:
			encoded, err := json.Marshal(struct {
				Value Field[string] `json:"value,omitzero"`
			}{})
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(encoded) != `{}` {
				t.Errorf("json.Marshal() = %s, want {}", encoded)
			}
		case 1:
			field := Set(value)
			if !field.IsSet() || field.IsNull() {
				t.Errorf("Set(%q) state = set:%t null:%t, want set:true null:false", value, field.IsSet(), field.IsNull())
			}
		case 2:
			field := Null[string]()
			if !field.IsSet() || !field.IsNull() {
				t.Errorf("Null() state = set:%t null:%t, want set:true null:true", field.IsSet(), field.IsNull())
			}
		}
	})
}

package integrations

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestStrictBoolFormYAMLParity(t *testing.T) {
	for _, value := range []string{"yes", "no", "null", "\"true\""} {
		var field struct {
			Required StrictBool `yaml:"required" json:"required"`
		}
		if e := yaml.Unmarshal([]byte("required: "+value), &field); e == nil && value != "null" {
			t.Fatal("accepted coerced YAML boolean", value)
		}
		if e := json.Unmarshal([]byte(`{"required":`+value+`}`), &field); e == nil {
			t.Fatal("accepted coerced JSON boolean", value)
		}
	}
	var field struct {
		Required StrictBool `yaml:"required" json:"required"`
	}
	if e := yaml.Unmarshal([]byte("required: true"), &field); e != nil || !field.Required {
		t.Fatal(field, e)
	}
	b, e := json.Marshal(field)
	if e != nil || string(b) != `{"required":true}` {
		t.Fatal(string(b), e)
	}
}

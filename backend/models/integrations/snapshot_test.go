package integrations

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestResourcePreservesLegacyFlatJSON(t *testing.T) {
	legacy := []byte(`{"contract_version":"waste_schedule/v1","resource_id":"resource","revision":2,"timezone":"Europe/London","services":[{"service_id":"service","dates":["2026-10-27"]}],"source_retrieved_at":null}`)
	original := append([]byte(nil), legacy...)
	var resource Resource
	if err := json.Unmarshal(legacy, &resource); err != nil {
		t.Fatal(err)
	}
	legacy[0] = 'x' // The decoder's input buffer must not remain owned by the resource.
	encoded, err := json.Marshal(resource)
	if err != nil {
		t.Fatal(err)
	}
	var before, after any
	_ = json.Unmarshal(original, &before)
	_ = json.Unmarshal(encoded, &after)
	if !reflect.DeepEqual(before, after) || resource.Revision != 2 {
		t.Fatal("changed legacy shape", string(encoded))
	}
	encoded[0] = 'x'
	if _, err = json.Marshal(resource); err != nil {
		t.Fatal("caller changed internal bytes", err)
	}
}

func TestResourceRejectsInvalidHeaders(t *testing.T) {
	for _, sample := range []string{`null`, `[]`, `{}`, `{"contract_version":"sample/v1","resource_id":"r","revision":0}`, `{"contract_version":"sample/v1","resource_id":"r","revision":"one"}`} {
		var resource Resource
		if err := json.Unmarshal([]byte(sample), &resource); err == nil {
			t.Fatalf("accepted %s", sample)
		}
	}
	if _, err := json.Marshal(Resource{}); err == nil {
		t.Fatal("accepted zero resource")
	}
}

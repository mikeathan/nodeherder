package operations

import "testing"

func TestPointerJSONLocations(t *testing.T) {
	root := map[string]any{
		"result": map[string]any{"collections": []any{map[string]any{"name": "Paper"}}},
		"null":   nil,
		"a/b":    map[string]any{"~1": "escaped"},
	}
	for _, test := range []struct {
		name, path string
		want       any
		found      bool
	}{
		{"nested array", "/result/collections/0/name", "Paper", true},
		{"explicit null", "/null", nil, true},
		{"missing", "/missing", nil, false},
		{"escaped names", "/a~1b/~01", "escaped", true},
		{"leading zero", "/result/collections/00", nil, false},
		{"append unsupported", "/result/collections/-", nil, false},
		{"invalid escape", "/bad~2", nil, false},
		{"scalar traversal", "/result/collections/0/name/child", nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, found := Pointer(root, test.path)
			if got != test.want || found != test.found {
				t.Fatalf("Pointer(%q) = (%v,%v), want (%v,%v)", test.path, got, found, test.want, test.found)
			}
		})
	}
	if got, found := Pointer(root, ""); !found || got.(map[string]any)["result"] == nil {
		t.Fatal("empty location must return the full root")
	}
}

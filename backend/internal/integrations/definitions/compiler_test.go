package definitions_test

import (
	"encoding/json"
	"strings"
	"testing"

	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/capabilities/waste"
	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
)

func seed(t *testing.T) string {
	t.Helper()
	b, e := defaults.Files.ReadFile("waste_collection/whenisbins.yaml")
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func TestCompileYAML(t *testing.T) {
	source := seed(t)
	c, e := compiler(t).CompileYAML(source)
	if e != nil {
		t.Fatal(e)
	}
	if c.Origin != "https://whenisbins.com" {
		t.Fatal(c.Origin)
	}
	cases := map[string]string{
		"unknown": source + "unknown: true\n", "duplicate": source + "id: duplicate\n", "alias": strings.Replace(source, "name: WhenIsBins", "name: &a WhenIsBins", 1), "documents": source + "---\nid: extra\n", "missing request": strings.Replace(source, "request: discover", "request: missing", 1), "forward capture": strings.Replace(source, "ref: input.postcode", "ref: state.property_token", 1), "complete capture in poll": strings.Replace(source, "ref: state.lookup_id", "ref: state.property_token", 1), "overlap": strings.Replace(source, "done: [done]", "done: [queued]", 1), "http": strings.Replace(source, "https://whenisbins.com", "http://localhost", 1), "prefix number": strings.Replace(source, "type: string, when: complete, strip_prefix", "type: number, when: complete, strip_prefix", -1), "bad pointer": strings.Replace(source, "items: /collections", "items: collections", 1), "version": strings.Replace(source, "nodeherder.integration/v1", "nodeherder.integration/v9", 1), "reserved": strings.Replace(source, "lookup_id: {pointer", "etag: {pointer", 1),
		"null boolean":                         strings.Replace(source, "required: true", "required: null", 1),
		"optional required state":              strings.ReplaceAll(source, "when: complete, strip_prefix", "optional: true, when: complete, strip_prefix"),
		"missing immediate completion capture": strings.Replace(source, "property_token: {pointer: /result/property_id, type: string, when: complete, strip_prefix: 'p:'}", "unused: {pointer: /result/property_id, type: string, when: complete, strip_prefix: 'p:'}", 1),
	}
	for name, y := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := compiler(t).CompileYAML(y); err == nil {
				t.Fatal("accepted invalid definition")
			}
		})
	}
	renamed := strings.ReplaceAll(source, "discover", "find_properties")
	if _, e := compiler(t).CompileYAML(renamed); e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(c.Definition)
	again, e := compiler(t).CompileDocument(raw)
	if e != nil {
		t.Fatal(e)
	}
	if again.Definition.Requests["discover"].Query["postcode"].Ref != "input.postcode" {
		t.Fatal("lost reference")
	}
}
func TestCompileRejectsUnreachableState(t *testing.T) {
	c, e := compiler(t).CompileYAML(seed(t))
	if e != nil {
		t.Fatal(e)
	}
	d := c.Definition
	unused := d.Requests["schedule"]
	unused.Captures = map[string]model.Capture{"ghost": {Pointer: "/id", Type: "string"}}
	d.Requests["unused"] = unused
	request := d.Requests["schedule"]
	request.Path = append(request.Path, model.Value{Ref: "state.ghost"})
	d.Requests["schedule"] = request
	raw, _ := json.Marshal(d)
	if _, e = compiler(t).CompileDocument(raw); e == nil {
		t.Fatal("unreachable capture was considered initialized")
	}
}

func compiler(t *testing.T) *definitions.Compiler {
	r, err := capabilities.NewRegistry(waste.Adapter{})
	if err != nil {
		t.Fatal(err)
	}
	return definitions.NewCompiler(r)
}

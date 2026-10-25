package integrationdefaults_test

import (
	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations"
	"testing"
)

func TestFiles(t *testing.T) {
	b, e := defaults.Files.ReadFile("waste_collection/whenisbins.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := integrations.CompileYAML(string(b))
	if e != nil {
		t.Fatal(e)
	}
	if integrations.DefinitionKey(d.Definition) != "waste_collection/whenisbins" {
		t.Fatal("seed identity")
	}
}

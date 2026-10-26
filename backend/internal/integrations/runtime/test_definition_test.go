package runtime

import (
	"testing"
)

func TestDefinitionPreviewIsolation(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	d, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	preview, e := s.TestDefinition(i.ProviderKey, d.YAML, i.ID, "live", nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	preview, e = s.Setup(preview.ID)
	if e != nil || preview.Preview == nil {
		t.Fatal(preview, e)
	}
	after, e := s.Instance(i.ID)
	if e != nil || after.Revision != i.Revision || after.Resources[0].Revision != i.Resources[0].Revision {
		t.Fatal("test changed active instance", after, e)
	}
}

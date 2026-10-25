package integrations

import (
	"context"
	"testing"
)

func TestBootstrapMissingKeyIsDegraded(t *testing.T) {
	t.Setenv("INTEGRATION_KEYRING_FILE", t.TempDir()+"/absent.json")
	service := Bootstrap(context.Background())
	catalog, e := service.Catalog()
	if e != nil || catalog.Available || catalog.Error == nil || catalog.Error.Code != "key_unavailable" {
		t.Fatal(catalog, e)
	}
	if e = service.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
}

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

func TestBootstrapNativeFirstRun(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("INTEGRATION_KEYRING_FILE", "")
	t.Setenv("DATA_DIR", root+"/data")
	t.Setenv("INTEGRATIONS_CONFIG_DIR", root+"/definitions")
	for n := 0; n < 2; n++ {
		service := Bootstrap(context.Background())
		catalog, err := service.Catalog()
		if err != nil || !catalog.Available || len(catalog.Providers) != 1 {
			t.Fatalf("native startup %d: %+v, %v", n, catalog, err)
		}
		if err := service.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	keys, err := bootstrapKeyring(root + "/data")
	if err != nil || keys.ActiveKeyID == "" {
		t.Fatal("native key not reused", err)
	}
}

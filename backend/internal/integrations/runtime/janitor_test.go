package runtime

import (
	"net/http"
	"testing"
	"time"
)

func TestCleanupExpiryAndNoProviderWork(t *testing.T) {
	calls := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) { calls++; return standardProvider(r) }))
	p, e := s.CreateSetup("waste_collection/whenisbins", "", "")
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(31 * time.Minute)
	if e = s.Cleanup(); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Setup(p.ID); e == nil || failure(e).Code != "setup_expired" {
		t.Fatal("expiry not reported", e)
	}
	if calls != 0 {
		t.Fatal("janitor called provider")
	}
}

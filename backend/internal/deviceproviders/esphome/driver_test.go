package esphome

import (
	"testing"

	"github.com/google/uuid"
	"node-herder/internal/deviceproviders"
)

func TestDriverValidatesBeforeStarting(t *testing.T) {
	d := NewDriver(&testPipeline{})
	item := deviceproviders.Enrollment{ID: uuid.NewString(), Provider: "esphome", Address: "192.168.1.2", Port: 6053}
	if _, err := d.NewInstance(item); err != nil {
		t.Fatal(err)
	}
	item.Credential = "invalid"
	if _, err := d.NewInstance(item); err == nil {
		t.Fatal("invalid key accepted")
	}
	item.Credential, item.Address = "", "127.0.0.1"
	if _, err := d.NewInstance(item); err == nil {
		t.Fatal("unsafe target accepted")
	}
}

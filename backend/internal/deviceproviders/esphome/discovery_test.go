package esphome

import (
	"net"
	"testing"

	"github.com/miekg/dns"
)

func TestDiscoveryCorrelatesRecordsAndRejectsUnsafeAddresses(t *testing.T) {
	records := newDiscoveryRecords()
	// Addresses may arrive before service advertisements.
	records.add(&dns.Msg{Extra: []dns.RR{
		&dns.A{Hdr: dns.RR_Header{Name: "radar.local.", Ttl: 60}, A: net.ParseIP("192.168.1.2")},
		&dns.A{Hdr: dns.RR_Header{Name: "unsafe.local.", Ttl: 60}, A: net.ParseIP("127.0.0.1")},
	}})
	records.add(&dns.Msg{Answer: []dns.RR{
		&dns.PTR{Hdr: dns.RR_Header{Name: discoveryService, Ttl: 60}, Ptr: "radar." + discoveryService},
		&dns.PTR{Hdr: dns.RR_Header{Name: discoveryService, Ttl: 60}, Ptr: "unsafe." + discoveryService},
	}, Extra: []dns.RR{
		&dns.SRV{Hdr: dns.RR_Header{Name: "radar." + discoveryService, Ttl: 60}, Target: "radar.local.", Port: 6053},
		&dns.SRV{Hdr: dns.RR_Header{Name: "unsafe." + discoveryService, Ttl: 60}, Target: "unsafe.local.", Port: 6053},
	}})
	got := records.candidates()
	if len(got) != 1 || got[0].Name != "radar" || got[0].Address != "192.168.1.2" || got[0].Port != 6053 {
		t.Fatalf("candidates: %+v", got)
	}
}

func TestDiscoveryRequestsMissingRecordsOnce(t *testing.T) {
	records := newDiscoveryRecords()
	records.add(&dns.Msg{Answer: []dns.RR{&dns.PTR{Hdr: dns.RR_Header{Name: discoveryService, Ttl: 60}, Ptr: "radar." + discoveryService}}})
	queried := make(map[string]bool)
	questions := records.missingQuestions(queried)
	if len(questions) != 1 || questions[0].Qtype != dns.TypeSRV {
		t.Fatalf("questions: %+v", questions)
	}
	if len(records.missingQuestions(queried)) != 0 {
		t.Fatal("repeated query")
	}
	records.add(&dns.Msg{Answer: []dns.RR{&dns.SRV{Hdr: dns.RR_Header{Name: "radar." + discoveryService, Ttl: 60}, Target: "radar.local.", Port: 6053}}})
	if questions := records.missingQuestions(queried); len(questions) != 2 {
		t.Fatalf("address questions: %+v", questions)
	}
}

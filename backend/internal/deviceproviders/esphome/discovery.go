package esphome

// This file performs one bounded mDNS browse for ESPHome advertisements. Results
// are untrusted candidates: they are validated as private numeric endpoints and
// are never connected to, trusted or enrolled automatically.
import (
	"context"
	"errors"
	"net"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/miekg/dns"
	"node-herder/internal/deviceproviders"
)

const (
	discoveryService   = "_esphomelib._tcp.local."
	discoveryGroupIPv4 = "224.0.0.251"
	discoveryPort      = 5353
)

// Candidate is an untrusted mDNS advertisement, not an enrolled or verified device.
type Candidate struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type serviceAddress struct {
	host string
	port int
}
type discoveryRecords struct {
	instances map[string]bool
	services  map[string]serviceAddress
	hosts     map[string][]net.IP
}

func newDiscoveryRecords() *discoveryRecords {
	return &discoveryRecords{instances: make(map[string]bool), services: make(map[string]serviceAddress), hosts: make(map[string][]net.IP)}
}

// Discover performs one bounded IPv4 multicast DNS browse and owns its socket.
// It never connects to or enrolls a result; cancellation closes the socket.
func Discover(ctx context.Context) ([]Candidate, error) {
	if ctx == nil {
		return nil, errors.New("discovery context is required")
	}
	ctx, cancel := context.WithTimeout(ctx, discoveryTimeout)
	defer cancel()
	group := &net.UDPAddr{IP: net.ParseIP(discoveryGroupIPv4), Port: discoveryPort}
	conn, err := net.ListenMulticastUDP("udp4", nil, group)
	if err != nil {
		return nil, errors.New("multicast discovery unavailable")
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, err
	}
	query := new(dns.Msg)
	query.SetQuestion(discoveryService, dns.TypePTR)
	query.Id, query.RecursionDesired = 0, false
	packet, err := query.Pack()
	if err != nil {
		return nil, err
	}
	if _, err := conn.WriteToUDP(packet, group); err != nil {
		return nil, errors.New("multicast discovery unavailable")
	}
	records := newDiscoveryRecords()
	queried := make(map[string]bool)
	buffer := make([]byte, discoveryPacketBytes)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, ctx.Err()
			}
			if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
				return records.candidates(), nil
			}
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return records.candidates(), nil
			}
			return nil, errors.New("multicast discovery unavailable")
		}
		message := new(dns.Msg)
		if message.Unpack(buffer[:n]) != nil || !message.Response {
			continue
		}
		records.add(message)
		questions := records.missingQuestions(queried)
		for len(questions) > 0 {
			count := min(len(questions), maxDiscoveryQuestions)
			followup := &dns.Msg{Question: questions[:count]}
			encoded, err := followup.Pack()
			if err != nil {
				return nil, err
			}
			if _, err := conn.WriteToUDP(encoded, group); err != nil {
				return nil, errors.New("multicast discovery unavailable")
			}
			questions = questions[count:]
		}
	}
}

// missingQuestions requests records not included in a multicast response. Every
// name is queried at most once and the maps bound the total work per scan.
func (r *discoveryRecords) missingQuestions(queried map[string]bool) []dns.Question {
	questions := make([]dns.Question, 0)
	add := func(name string, kind uint16) {
		key := name + string(rune(kind))
		if queried[key] {
			return
		}
		queried[key] = true
		questions = append(questions, dns.Question{Name: name, Qtype: kind, Qclass: dns.ClassINET})
	}
	for name := range r.instances {
		if _, ok := r.services[name]; !ok {
			add(name, dns.TypeSRV)
		}
	}
	for _, service := range r.services {
		if len(r.hosts[service.host]) == 0 && strings.HasSuffix(service.host, ".local.") {
			add(service.host, dns.TypeA)
			add(service.host, dns.TypeAAAA)
		}
	}
	return questions
}

// add accepts records in any packet order and bounds all attacker-controlled maps.
func (r *discoveryRecords) add(message *dns.Msg) {
	all := append(append(append([]dns.RR{}, message.Answer...), message.Ns...), message.Extra...)
	if len(all) > maxDiscoveryRecords {
		return
	}
	for _, rr := range all {
		if rr.Header().Ttl == 0 {
			continue
		}
		switch value := rr.(type) {
		case *dns.PTR:
			if strings.EqualFold(value.Hdr.Name, discoveryService) && validDiscoveryName(value.Ptr) && strings.HasSuffix(strings.ToLower(value.Ptr), "."+discoveryService) && len(r.instances) < maxDiscoveryRecords {
				r.instances[strings.ToLower(value.Ptr)] = true
			}
		case *dns.SRV:
			if strings.HasSuffix(strings.ToLower(value.Hdr.Name), "."+discoveryService) && validDiscoveryName(value.Hdr.Name) && validDiscoveryName(value.Target) && len(r.services) < maxDiscoveryRecords {
				r.services[strings.ToLower(value.Hdr.Name)] = serviceAddress{strings.ToLower(value.Target), int(value.Port)}
			}
		case *dns.A:
			r.addHost(value.Hdr.Name, value.A)
		case *dns.AAAA:
			r.addHost(value.Hdr.Name, value.AAAA)
		}
	}
}

func (r *discoveryRecords) addHost(name string, address net.IP) {
	name = strings.ToLower(name)
	if !validDiscoveryName(name) || len(r.hosts) >= maxDiscoveryRecords || len(r.hosts[name]) >= maxDiscoveryHosts {
		return
	}
	r.hosts[name] = append(r.hosts[name], append(net.IP(nil), address...))
}

func validDiscoveryName(name string) bool {
	return len(name) > 0 && len(name) <= maxEntityNameBytes && utf8.ValidString(name) && strings.IndexFunc(name, unicode.IsControl) < 0
}

func (r *discoveryRecords) candidates() []Candidate {
	result := make([]Candidate, 0)
	seen := make(map[string]bool)
	for instance := range r.instances {
		service, ok := r.services[instance]
		if !ok {
			continue
		}
		for _, ip := range r.hosts[service.host] {
			endpoint, err := deviceproviders.NewEndpoint(ip.String(), service.port)
			if err != nil || seen[endpoint.NetworkAddress()] {
				continue
			}
			address, _, err := net.SplitHostPort(endpoint.NetworkAddress())
			if err != nil {
				continue
			}
			seen[endpoint.NetworkAddress()] = true
			name := strings.TrimSuffix(instance, "."+discoveryService)
			result = append(result, Candidate{Name: name, Address: address, Port: service.port})
			if len(result) == maxDiscoveryCandidates {
				break
			}
		}
		if len(result) == maxDiscoveryCandidates {
			break
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Address < result[j].Address })
	return result
}

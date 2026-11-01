package connection

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/mycontroller-org/esphome_api/pkg/api"
)

func TestPlaintextRejectsOversizedFrame(t *testing.T) {
	local, peer := net.Pipe()
	defer local.Close()
	defer peer.Close()
	c, err := NewPlaintextConnection(local, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	header := []byte{0}
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, ^uint64(0))
	header = append(header, buf[:n]...)
	header = append(header, 1)
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("oversized peer frame panicked: %v", recovered)
		}
	}()
	if _, err = c.Read(bufio.NewReader(bytes.NewReader(header))); err == nil {
		t.Fatal("oversized frame accepted")
	}
}

func TestPlaintextWriteLargeMessage(t *testing.T) {
	local, peer := net.Pipe()
	defer local.Close()
	defer peer.Close()
	sender, _ := NewPlaintextConnection(local, time.Second)
	receiver, _ := NewPlaintextConnection(peer, time.Second)
	result := make(chan error, 1)
	go func() { result <- sender.Write(&api.HelloRequest{ClientInfo: strings.Repeat("a", 20000)}) }()
	message, err := receiver.Read(bufio.NewReader(peer))
	if err != nil {
		t.Fatal(err)
	}
	if len(message.(*api.HelloRequest).ClientInfo) != 20000 {
		t.Fatal("message was truncated")
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if err := sender.Write(&api.HelloRequest{ClientInfo: strings.Repeat("a", 65537)}); err == nil {
		t.Fatal("accepted oversized message")
	}
}

func FuzzPlaintextRead(f *testing.F) {
	f.Add([]byte{0, 0, 1})
	f.Add([]byte{0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 1})
	f.Fuzz(func(t *testing.T, frame []byte) {
		c := &PlaintextConnection{}
		_, _ = c.Read(bufio.NewReader(bytes.NewReader(frame)))
	})
}

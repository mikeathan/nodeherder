package client

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	"github.com/mycontroller-org/esphome_api/pkg/connection"
	"google.golang.org/protobuf/proto"
)

func TestCloseContextCancelsAndConcurrentCloseCompletes(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		peer, err := l.Accept()
		if err == nil {
			accepted <- peer
		}
	}()
	c, err := GetClient("test", l.Addr().String(), "", time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	peer := <-accepted
	defer peer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- c.CloseContext(ctx) }()
	if _, err := peer.Read(make([]byte, 32)); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = c.Close() }()
	}
	wg.Wait()
}

func TestWaiterRejectsConcurrentSameType(t *testing.T) {
	c := &Client{waitMap: make(map[uint64]chan proto.Message)}
	first := make(chan proto.Message, 1)
	second := make(chan proto.Message, 1)
	if !c.waitFor(1, first) || c.waitFor(1, second) {
		t.Fatal("concurrent waiter replaced the first waiter")
	}
	c.waitDone(1, second)
	if c.waitMap[1] != first {
		t.Fatal("unrelated cleanup removed the active waiter")
	}
	c.waitDone(1, first)
	if !c.waitFor(1, second) {
		t.Fatal("waiter was not released")
	}
}

func TestResponseWaitEndsWhenPeerDisconnects(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := l.Accept()
		if err == nil {
			accepted <- c
		}
	}()
	c, err := GetClient("test", l.Addr().String(), "", time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	peer := <-accepted
	result := make(chan error, 1)
	go func() { _, err := c.Hello(); result <- err }()
	if _, err := peer.Read(make([]byte, 64)); err != nil {
		t.Fatal(err)
	}
	_ = peer.Close()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("expected connection failure")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("response waiter did not stop after disconnect")
	}
}

func TestDialContextCancelsStalledEncryptedHandshake(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := l.Accept()
		if err == nil {
			accepted <- c
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := DialContext(ctx, "test", l.Addr().String(), base64.StdEncoding.EncodeToString(make([]byte, 32)), time.Second, nil)
		result <- err
	}()
	peer := <-accepted
	defer peer.Close()
	// Consume the client's opening frame so cancellation occurs during the handshake read.
	if _, err := io.ReadFull(peer, make([]byte, 3)); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("handshake did not stop after cancellation")
	}
	if err := peer.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := peer.Read(make([]byte, 1)); err == nil {
		t.Fatal("connection remained open")
	} else if e, ok := err.(net.Error); ok && e.Timeout() {
		t.Fatal("connection was not closed")
	}
}

func TestInvalidEncryptionKeyClosesConnection(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := l.Accept()
		if err == nil {
			accepted <- c
		}
	}()
	_, err = GetClient("test", l.Addr().String(), "invalid-key", time.Second, nil)
	if err == nil {
		t.Fatal("invalid key accepted")
	}
	peer := <-accepted
	defer peer.Close()
	if err = peer.SetReadDeadline(time.Now().Add(150 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	_, err = peer.Read(make([]byte, 1))
	if e, ok := err.(net.Error); ok && e.Timeout() {
		t.Fatal("connection remains open after setup failed")
	}
	if err == nil {
		t.Fatal("unexpected data from failed handshake")
	}
}

func TestLastMessageAtConcurrentWithReports(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := l.Accept()
		if err == nil {
			accepted <- c
		}
	}()
	c, err := GetClient("test", l.Addr().String(), "", 100*time.Millisecond, nil)
	if err != nil {
		t.Fatal(err)
	}
	peer := <-accepted
	defer peer.Close()
	defer c.Close()
	wire, err := connection.NewPlaintextConnection(peer, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			if wire.Write(&api.SensorStateResponse{Key: 1, State: float32(i)}) != nil {
				return
			}
		}
	}()
	for i := 0; i < 100000; i++ {
		_ = c.LastMessageAt()
	}
	<-done
}

package connection

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/flynn/noise"
	"github.com/mycontroller-org/esphome_api/pkg/api"
	"google.golang.org/protobuf/proto"
)

func TestEncryptedConcurrentWritesKeepNonceOrder(t *testing.T) {
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	key := make([]byte, 32)
	c, err := NewEncryptedConnection(client, time.Second, base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- noisePeer(peer, key, 40, false) }()
	if err := c.Handshake(); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- c.Write(&api.PingRequest{}) }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}

func TestHandshakePreservesPipelinedApplicationFrame(t *testing.T) {
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	key := make([]byte, 32)
	c, err := NewEncryptedConnection(client, time.Second, base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- noisePeer(peer, key, 0, true) }()
	if err := c.Handshake(); err != nil {
		t.Fatal(err)
	}
	message, err := c.Read(bufio.NewReader(client))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := message.(*api.PingResponse); !ok {
		t.Fatalf("unexpected message %T", message)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}

func TestHandshakeRejectsUnknownProtocol(t *testing.T) {
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	c, err := NewEncryptedConnection(client, time.Second, base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = io.ReadFull(peer, make([]byte, 3))
		_, _ = peer.Write([]byte{1, 0, 1, 2})
	}()
	if err := c.Handshake(); !errors.Is(err, ErrHandshakeFailed) {
		t.Fatalf("got %v", err)
	}
}

// noisePeer independently implements the responder side of the Noise NNpsk0 exchange.
func noisePeer(conn net.Conn, key []byte, count int, pipeline bool) error {
	readFrame := func() ([]byte, error) {
		var header [3]byte
		if _, err := io.ReadFull(conn, header[:]); err != nil {
			return nil, err
		}
		if header[0] != 1 {
			return nil, fmt.Errorf("unexpected preamble %d", header[0])
		}
		body := make([]byte, binary.BigEndian.Uint16(header[1:]))
		_, err := io.ReadFull(conn, body)
		return body, err
	}
	writeFrame := func(body []byte) error {
		frame := []byte{1, byte(len(body) >> 8), byte(len(body))}
		_, err := conn.Write(append(frame, body...))
		return err
	}
	if _, err := readFrame(); err != nil {
		return err
	}
	if err := writeFrame([]byte{1}); err != nil {
		return err
	}
	state, err := noise.NewHandshakeState(noise.Config{Pattern: noise.HandshakeNN, CipherSuite: noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256), Prologue: []byte("NoiseAPIInit\x00\x00"), PresharedKey: key, PresharedKeyPlacement: 0})
	if err != nil {
		return err
	}
	message, err := readFrame()
	if err != nil {
		return err
	}
	if len(message) == 0 || message[0] != 0 {
		return fmt.Errorf("invalid handshake")
	}
	if _, _, _, err = state.ReadMessage(nil, message[1:]); err != nil {
		return err
	}
	response, decrypt, encrypt, err := state.WriteMessage(nil, nil)
	if err != nil {
		return err
	}
	handshake := append([]byte{0}, response...)
	if pipeline {
		plain := []byte{0, byte(api.PingResponseTypeID), 0, 0}
		ciphertext, err := encrypt.Encrypt(nil, nil, plain)
		if err != nil {
			return err
		}
		combined := append([]byte{1, byte(len(handshake) >> 8), byte(len(handshake))}, handshake...)
		combined = append(combined, 1, byte(len(ciphertext)>>8), byte(len(ciphertext)))
		combined = append(combined, ciphertext...)
		if _, err := conn.Write(combined); err != nil {
			return err
		}
	} else if err = writeFrame(handshake); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		frame, err := readFrame()
		if err != nil {
			return err
		}
		plain, err := decrypt.Decrypt(nil, nil, frame)
		if err != nil {
			return err
		}
		decoded, err := decodeEncryptedMessage(plain)
		if err != nil {
			return err
		}
		if _, ok := decoded.(*api.PingRequest); !ok {
			return fmt.Errorf("unexpected message %T", decoded)
		}
	}
	return nil
}

func TestDecodeEncryptedMessageValidatesHeaderAndLength(t *testing.T) {
	for _, payload := range [][]byte{nil, {0}, {0, 1, 0}, {0, 1, 0, 1}} {
		if _, err := decodeEncryptedMessage(payload); err == nil {
			t.Fatalf("accepted malformed payload %v", payload)
		}
	}
	want := &api.HelloResponse{Name: "fixture"}
	body, err := proto.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	payload := append([]byte{0, byte(api.HelloResponseTypeID), 0, byte(len(body))}, body...)
	got, err := decodeEncryptedMessage(payload)
	if err != nil || !proto.Equal(got, want) {
		t.Fatalf("valid message: got %v, error %v", got, err)
	}
}

func FuzzDecodeEncryptedMessage(f *testing.F) {
	f.Add([]byte{0, 1, 0, 0})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, frame []byte) {
		_, _ = decodeEncryptedMessage(frame)
	})
}

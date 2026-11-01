package connection

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	types "github.com/mycontroller-org/esphome_api/pkg/types"
	"google.golang.org/protobuf/proto"
)

type PlaintextConnection struct {
	conn                 net.Conn
	CommunicationTimeout time.Duration
	writeMutex           sync.Mutex
	readMutex            sync.Mutex
}

// maxPlaintextFrameBytes limits allocations from untrusted native API frame lengths.
const maxPlaintextFrameBytes = 64 * 1024

func NewPlaintextConnection(conn net.Conn, communicationTimeout time.Duration) (ApiConnection, error) {
	ptc := &PlaintextConnection{
		conn:                 conn,
		CommunicationTimeout: communicationTimeout,
	}
	return ptc, nil
}

func (ptc *PlaintextConnection) Handshake() error {
	return nil
}

func (ptc *PlaintextConnection) Write(message proto.Message) error {
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	if len(messageBytes) > maxPlaintextFrameBytes {
		return errors.New("esphome_api: plaintext frame exceeds size limit")
	}
	packed := binary.AppendUvarint([]byte{0}, uint64(len(messageBytes)))
	packed = binary.AppendUvarint(packed, api.TypeID(message))
	packed = append(packed, messageBytes...)

	// set write lock
	ptc.writeMutex.Lock()
	defer ptc.writeMutex.Unlock()

	err = ptc.conn.SetWriteDeadline(time.Now().Add(ptc.CommunicationTimeout))
	if err != nil {
		return err
	}
	n, err := ptc.conn.Write(packed)
	if err == nil && n != len(packed) {
		return io.ErrShortWrite
	}
	return err
}

func (ptc *PlaintextConnection) Read(reader *bufio.Reader) (proto.Message, error) {
	// set read lock
	ptc.readMutex.Lock()
	defer ptc.readMutex.Unlock()

	preamble, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if preamble != 0x00 {
		if preamble == 0x01 {
			return nil, types.ErrConnRequireEncryption
		}
		return nil, errors.New("esphome_api: invalid preamble. should starts with 0x00 byte")
	}

	// get message bytes length
	messageSize, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}
	if messageSize > maxPlaintextFrameBytes {
		return nil, errors.New("esphome_api: plaintext frame exceeds size limit")
	}

	// get message type id
	messageTypeID, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}

	// get message bytes
	messageBytes := make([]byte, messageSize)
	_, err = io.ReadFull(reader, messageBytes)
	if err != nil {
		return nil, err
	}

	message := api.NewMessageByTypeID(messageTypeID)
	if message == nil {
		return nil, fmt.Errorf("esphome_api: protocol error: unknown message type %#x", messageTypeID)
	}

	err = proto.Unmarshal(messageBytes, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

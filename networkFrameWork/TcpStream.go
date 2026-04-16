package networkFrameWork

import (
	"bnfs_p2p/network"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"io"
	"net"
	"sync"
)

type TcpStream struct {
	nodeId       string
	connection   net.Conn
	lock         sync.Mutex
	connectionId string
}

func (t *TcpStream) Close() error {
	return t.connection.Close()
}

func (t *TcpStream) NextMessage() (*network.Message, error) {
	headerRead := make([]byte, network.HeaderLength)
	_, err := io.ReadFull(t.connection, headerRead)
	if err != nil {
		return nil, err
	}
	header, err := network.ParseHeader(headerRead)
	if err != nil {
		return nil, err
	}
	payLoad := make([]byte, header.PayLoadLength)
	_, err = io.ReadFull(t.connection, payLoad)
	if err != nil {
		return nil, err
	}
	return &network.Message{Header: header, Payload: payLoad}, nil
}

func (t *TcpStream) SendMessage(ctx context.Context, message *network.Message) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	bytes, err := message.ParseToBytes()
	if err != nil {
		return err
	}
	_, err = t.connection.Write(bytes)
	return err
}

func (t *TcpStream) NodeId() string {
	return t.nodeId
}
func (t *TcpStream) ConnectionId() string {
	return t.connectionId
}
func TryConnectTCPStream(addr, targetNodeId, originalPubkeyHex string) (network.Stream, string, error) {
	connectionId := uuid.New().String()
	header := &network.Header{
		RouteName:     "",
		NodeId:        targetNodeId,
		NodeIdVersion: 1,
		PayLoadLength: 0,
		ConnectionId:  connectionId,
		OriginData:    nil,
	}
	body := &network.Message{
		Header:  header,
		Payload: []byte(originalPubkeyHex),
	}
	stream, err := clientStream(body, addr, targetNodeId)
	return stream, connectionId, err
}

func TryRegisterRelayStream(pubKey, relayAddress string) (network.Stream, error) {
	hash := sha256.Sum256([]byte(pubKey))
	originalNodeId := hex.EncodeToString(hash[:])

	header := &network.Header{
		RouteName:     "",
		NodeId:        originalNodeId,
		NodeIdVersion: 1,
		PayLoadLength: 0,
		ConnectionId:  "",
		OriginData:    nil,
	}
	body := &network.Message{
		Header:  header,
		Payload: []byte(pubKey),
	}
	return clientStream(body, relayAddress, originalNodeId)
}

func clientStream(FirstMessage *network.Message, tcpAddr, originalNodeId string) (network.Stream, error) {
	conn, err := net.Dial("tcp4", tcpAddr)
	if err != nil {
		return nil, err
	}
	bytes, err := FirstMessage.ParseToBytes()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_, err = conn.Write(bytes)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &TcpStream{
		nodeId:     originalNodeId,
		connection: conn,
	}, nil
}

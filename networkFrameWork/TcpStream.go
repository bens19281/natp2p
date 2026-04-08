package networkFrameWork

import (
	"bnfs_p2p/network"
	"context"
	"github.com/google/uuid"
	"io"
	"net"
	"sync"
)

type TcpStream struct {
	nodeId     string
	connection net.Conn
	lock       sync.Mutex
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
func TryConnectTCPStream(addr, targetNodeId, originalNodeId string) (network.Stream, string, error) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, "", err
	}
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
		Payload: []byte(originalNodeId),
	}
	bytes, err := body.ParseToBytes()
	if err != nil {
		conn.Close()
		return nil, "", err
	}
	_, err = conn.Write(bytes)
	if err != nil {
		conn.Close()
		return nil, "", err
	}

	return &TcpStream{
		nodeId:     originalNodeId,
		connection: conn,
	}, connectionId, nil
}

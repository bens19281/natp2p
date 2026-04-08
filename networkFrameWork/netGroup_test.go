package networkFrameWork

import (
	"bnfs_p2p/crypoto"
	"bnfs_p2p/network"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func NodeEstablish(address, originalNodeIdSource string, t *testing.T, wg *sync.WaitGroup, continueEstablish bool) (Id string) {
	dial, err := net.Dial("tcp", address)
	if err != nil {
		return ""
	}
	//originalNodeIdSource := "test-node-id-source-string"
	pair, err := crypoto.MakeKeyPair()
	if err != nil {
		t.Fatalf("Failed to make key pair: %v", err)
		return ""
	}
	pairId := crypoto.GetPubKeyStr(&pair.PublicKey)
	hash := sha256.Sum256([]byte(pairId))
	originalNodeId := hex.EncodeToString(hash[:])

	go func() {
		defer wg.Done()
		//time.Sleep(1 * time.Second)
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
			Payload: []byte(pairId),
		}
		bytes, err := body.ParseToBytes()
		if err != nil {
			t.Errorf("ERROR:" + err.Error())
			return
		}
		// 第一次建立连接
		_, err = dial.Write(bytes)
		if err != nil {
			t.Errorf("ERROR:" + err.Error())
			return
		}
		for {
			// 接下来接收响应 写入
			ClientFirstMessage, _ := tryReadMessageFromConnection(dial)
			marshal, err := json.Marshal(ClientFirstMessage)
			if err != nil {
				return
			}
			t.Logf("CurruentRelayId %s", originalNodeId)
			t.Log("ClientFirstMessage:", string(marshal))
			t.Logf("NodeId : %s", string(ClientFirstMessage.Payload))
			//ClientFirstMessage.Header.NodeId = originalNodeId
			ResponseMessage := &network.Message{
				Header:  ClientFirstMessage.Header,
				Payload: []byte("hello world"),
			}
			bytes, err = ResponseMessage.ParseToBytes()
			if err != nil {
				t.Errorf("ERROR:" + err.Error())
				return
			}
			// 发送响应
			_, err = dial.Write(bytes)
			if !continueEstablish {
				dial.Close()
				return
			}
		}

	}()
	return originalNodeId
}

func ClientTestWithStream(clientSteam network.Stream, connectionId, targetNodeId string, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
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
		Payload: []byte("hello server"),
	}
	clientSteam.SendMessage(context.Background(), body)
	ClientFirstMessage, _ := clientSteam.NextMessage()

	t.Log("RelayServerFirstMessage:", string(ClientFirstMessage.Payload))
}

func ClientSendTestMessage(TargetNodeId, RelayServerAddr, originalNodeIdSource string, t *testing.T, wg *sync.WaitGroup) string {
	dial, err := net.Dial("tcp", RelayServerAddr)
	if err != nil {
		return ""
	}
	//originalNodeIdSource := "test-node-id1-source-string"
	pair, err := crypoto.MakeKeyPair()
	if err != nil {
		t.Fatalf("Failed to make key pair: %v", err)
		return ""
	}
	pairId := crypoto.GetPubKeyStr(&pair.PublicKey)
	hash := sha256.Sum256([]byte(pairId))
	originalNodeId := hex.EncodeToString(hash[:])

	go func() {
		defer wg.Done()
		//time.Sleep(1 * time.Second)
		connectionId := uuid.New().String()
		header := &network.Header{
			RouteName:     "",
			NodeId:        TargetNodeId,
			NodeIdVersion: 1,
			PayLoadLength: 0,
			ConnectionId:  connectionId,
			OriginData:    nil,
		}
		body := &network.Message{
			Header:  header,
			Payload: []byte(pairId),
		}
		bytes, err := body.ParseToBytes()
		if err != nil {
			t.Errorf("ERROR:" + err.Error())
			return
		}
		// 第一次建立连接
		_, err = dial.Write(bytes)
		if err != nil {
			t.Errorf("ERROR:" + err.Error())
			return
		}
		time.Sleep(2 * time.Second)
		dial.Write(bytes)
		ClientFirstMessage, _ := tryReadMessageFromConnection(dial)
		t.Log("RelayServerFirstMessage:", string(ClientFirstMessage.Payload))
		dial.Close()
	}()
	return originalNodeId
}

func TestNewStreamGroup(t *testing.T) {
	tcpListener, err := net.Listen("tcp", ":9000")
	if err != nil {
		t.Fatalf("Failed to start relay server: %v", err)
	}
	t.Log("Relay Server started on :9000")
	var group *StreamGroup
	wg := &sync.WaitGroup{}
	index := 0
	establishId := NodeEstablish("127.0.0.1:9000", "test-node-id-source-string", t, wg, false)
	t.Log("establishId:", establishId)
	ClientSendTestMessage(establishId, "127.0.0.1:9000", "test-node-id1-source-string", t, wg)
	for index <= 1 {
		accept, err := tcpListener.Accept()
		if err != nil {
			t.Logf("Failed to accept connection: %v", err)
			return
		}
		if index%2 == 0 {

			firstMessage, _ := tryReadMessageFromConnection(accept)
			stream, err := TrySetupRelayStream(accept, firstMessage)
			if err != nil {
				t.Logf("Failed to setup relay stream: %v", err)
				return
			}
			group = NewStreamGroup(stream, defaultHookfunc)
			go func() {
				group.StartListen()
			}()
		} else {
			firstMessage, _ := tryReadMessageFromConnection(accept)
			stream, err := TrySetupRelayStream(accept, firstMessage)
			if err != nil {
				t.Logf("Failed to setup relay stream: %v", err)
				return
			}
			go func() {
				group.StreamOn(stream, firstMessage)
			}()
		}
		index++
		wg.Add(1)
	}
	wg.Wait()

}

func TestNewTransportCover(t *testing.T) {
	tcpListener, err := net.Listen("tcp", ":9000")
	if err != nil {
		t.Fatalf("Failed to start relay server: %v", err)
	}
	wg := &sync.WaitGroup{}

	transport := NewTransportCover()
	establishId := NodeEstablish("127.0.0.1:9000", "test-node-id-source-string", t, wg, true)
	//wg.Add(1)
	// 开始注册一个relay Connection
	accept, err := tcpListener.Accept()
	t.Logf("accept: %v", -1)
	if err != nil {
		t.Logf("Failed to accept connection: %v", err)
	}
	err = transport.ListenTCPConnection(accept)
	if err != nil {
		t.Fatalf("Failed to transport connection: %v", err)
		return
	}
	for i := 0; i < 9; i++ {
		wg.Add(1)
		ClientId := ClientSendTestMessage(establishId, "127.0.0.1:9000", fmt.Sprintf("test-node-id%d-source-string", i), t, wg)
		t.Logf("ClientId: %s ,Client Index: %d ", ClientId, i)
		accept, err = tcpListener.Accept()
		t.Logf("accept: %v", i)
		if err != nil {
			t.Logf("Failed to accept connection: %v", err)
		}
		err = transport.ListenTCPConnection(accept)
		if err != nil {
			t.Fatalf("Failed to transport connection: %v", err)
			return
		}

	}
	wg.Wait()

}

func TestRandomRelayClientInteraction(t *testing.T) {
	tcpListener, err := net.Listen("tcp", ":9000")
	if err != nil {
		t.Fatalf("Failed to start relay server: %v", err)
	}
	defer tcpListener.Close() // 确保测试结束后关闭监听器

	wg := &sync.WaitGroup{}
	transport := NewTransportCover()

	// 配置区域：简单控制 Relay 和 Client 的数量或比例
	baseRelayCount := 3
	baseClientCount := 10

	t.Logf("Starting test with %d Relays and %d Clients", baseRelayCount, baseClientCount)
	// 用于存储 Relay 节点的 ID，供 Client 随机选择
	var relayNodeIds []string
	// 1. 建立指定数量的 Relay 持久连接
	for i := 0; i < baseRelayCount; i++ {
		// 每个 Relay 使用不同的 NodeID Source 以区分
		relayNodeIdSource := fmt.Sprintf("relay-node-id-source-%d", i)
		establishId := NodeEstablish("127.0.0.1:9000", relayNodeIdSource, t, wg, true)
		t.Logf("Relay %d established with ID: %s", i, establishId)

		// 将建立的 Relay ID 存入列表
		relayNodeIds = append(relayNodeIds, establishId)

		accept, err := tcpListener.Accept()
		if err != nil {
			t.Logf("Failed to accept relay connection %d: %v", i, err)
			continue
		}

		err = transport.ListenTCPConnection(accept)
		if err != nil {
			t.Fatalf("Failed to transport relay connection %d: %v", i, err)
		}
	}

	// 2. 模拟指定数量的 Client 节点与 Relay 交互
	for i := 0; i < baseClientCount; i++ {
		wg.Add(1)

		// 随机选择一个 Relay ID 作为目标
		targetRelayId := ""
		if len(relayNodeIds) > 0 {
			randomIndex := rand.Intn(len(relayNodeIds))
			targetRelayId = relayNodeIds[randomIndex]
		} else {
			// 如果没有可用的 Relay，跳过或报错，这里选择跳过并减少 wg
			t.Logf("No available relays for client %d", i)
			wg.Done()
			continue
		}

		clientID := fmt.Sprintf("client-node-id-%d", i)
		hash := sha256.Sum256([]byte(clientID))
		originalNodeId := hex.EncodeToString(hash[:])
		go func() {
			stream, connectionId, err := TryConnectTCPStream("127.0.0.1:9000", targetRelayId, originalNodeId)
			if err != nil {
				t.Logf("Failed to connect to relay: %v", err)
				return
			}
			ClientTestWithStream(stream, connectionId, targetRelayId, t, wg)
		}()

		// 使用随机选择的 targetRelayId 替换原来的固定 targetRelaySource

		t.Logf("Client %d started with ID: %s targeting Relay(%s)", i, originalNodeId, targetRelayId)
		//clientSource := fmt.Sprintf("client-node-id-source-%d", i)
		//ClientId := ClientSendTestMessage(targetRelayId, "127.0.0.1:9000", clientSource, t, wg)
		//t.Logf("Client %d started with ID: %s targeting Relay(%s)", i, ClientId, targetRelayId)
		accept, err := tcpListener.Accept()
		if err != nil {
			t.Logf("Failed to accept client connection %d: %v", i, err)
			wg.Done() // 平衡 wg.Add
			continue
		}

		err = transport.ListenTCPConnection(accept)
		if err != nil {
			t.Logf("Failed to transport client connection %d: %v", i, err)
			// 这里不直接 fatal，因为可能是单个客户端失败，允许其他继续
		}
	}

	wg.Wait()

	t.Log("Test finished: All clients completed.")
}

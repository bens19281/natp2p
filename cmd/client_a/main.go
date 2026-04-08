package main

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"github.com/pion/ice/v3"
	"log"
	"net"
)

func main() {
	// 生成 256 位 (32 字节) 的 Node ID，这里用 UUID 模拟唯一标识，实际可用 sha256(rand)
	nodeID := uuid.New().String()
	fmt.Printf("Generated NodeID: %s\n", nodeID)

	// 连接中继服务器
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatalf("Failed to connect to relay server: %v", err)
	}
	defer conn.Close()

	// 注册
	_, err = conn.Write([]byte(fmt.Sprintf("REGISTER %s\n", nodeID)))
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')
	fmt.Print("Server response: ", response)

	if response[:2] != "OK" {
		log.Fatal("Registration failed")
	}

	fmt.Println("Waiting for messages from ClientB... (Press Ctrl+C to exit)")

	// 启动 ICE Agent (此处仅为框架展示，实际穿透需交换 Candidate)
	// 在简单中继模式下，我们直接通过 TCP 连接接收转发来的消息
	// 如果需要真正的 P2P，需在此处实现 OnConnectionStateChange 和 Candidate 交换

	agentConfig := &ice.AgentConfig{
		NetworkTypes: []ice.NetworkType{ice.NetworkTypeTCP4}, // 测试用
	}
	agent, err := ice.NewAgent(agentConfig)
	if err != nil {
		log.Printf("Warning: Failed to create ICE agent (normal in loopback test): %v", err)
	} else {
		defer agent.Close()
		fmt.Println("ICE Agent initialized (Ready for P2P negotiation)")
	}

	// 循环读取服务器转发的消息
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Connection lost: %v", err)
			break
		}
		fmt.Printf("Received: %s", msg)
	}
}

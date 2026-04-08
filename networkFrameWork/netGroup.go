package networkFrameWork

import (
	"bnfs_p2p/network"
	"sync"
)

// NetGroup 中转服务器: 分为3种模式: TURN STURN Relay中转服务器,优先实现Relay中转
type NetGroup struct {
	SessionMap map[string]network.Stream
	lock       sync.RWMutex
}

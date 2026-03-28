package relay

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

// Manager controls the lifecycle of the embedded tunnel client.
type Manager struct {
	mu sync.Mutex

	parentCtx context.Context
	cancel    context.CancelFunc

	deploymentMode string
	serverURL      string
	localAPI       string
	token          string
	redisClient    *redis.Client
}

func NewManager(deploymentMode, serverURL, localAPI, token string) *Manager {
	return &Manager{
		deploymentMode: deploymentMode,
		serverURL:      serverURL,
		localAPI:       localAPI,
		token:          token,
	}
}

func (m *Manager) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.parentCtx = ctx
	m.startLocked()
}

func (m *Manager) SetRedisClient(client *redis.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.redisClient = client
}

func (m *Manager) UpdateToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = token
	m.deploymentMode = "tunnel"
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.startLocked()
}

func (m *Manager) startLocked() {
	if m.parentCtx == nil {
		return
	}
	if m.deploymentMode != "tunnel" || m.token == "" {
		return
	}
	runCtx, cancel := context.WithCancel(m.parentCtx)
	m.cancel = cancel
	c := NewClientWithRedis(m.serverURL, m.token, m.localAPI, m.redisClient)
	go c.Start(runCtx)
}

type State struct {
	Mode     string `json:"mode"`
	HasToken bool   `json:"has_token"`
	Running  bool   `json:"running"`
}

func (m *Manager) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return State{
		Mode:     m.deploymentMode,
		HasToken: m.token != "",
		Running:  m.cancel != nil,
	}
}

var (
	globalManager   *Manager
	globalManagerMu sync.RWMutex
)

func SetGlobalManager(m *Manager) {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()
	globalManager = m
}

func GetGlobalManager() *Manager {
	globalManagerMu.RLock()
	defer globalManagerMu.RUnlock()
	return globalManager
}

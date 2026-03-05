package relay

import (
	"context"
	"sync"
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
	c := NewClient(m.serverURL, m.token, m.localAPI)
	go c.Start(runCtx)
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

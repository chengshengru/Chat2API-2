package session

import (
	"sync"
	"time"

	"chat2api-wails/internal/logger"
	"chat2api-wails/internal/store"
	"chat2api-wails/internal/types"
)

// Manager 会话管理器
type Manager struct {
	mu           sync.RWMutex
	logger       *logger.Logger
	storeManager *store.StoreManager
}

// NewManager 创建新的会话管理器
func NewManager(sm *store.StoreManager, l *logger.Logger) *Manager {
	m := &Manager{
		logger:       l,
		storeManager: sm,
	}

	l.Info("Session manager initialized")
	return m
}

// GetActiveSessions 获取活跃会话
func (m *Manager) GetActiveSessions() []types.SessionRecord {
	return m.storeManager.GetActiveSessions()
}

// CreateSession 创建新会话
func (m *Manager) CreateSession(providerId, accountId, providerType string, credentials map[string]string) (*types.SessionRecord, error) {
	session := &types.SessionRecord{
		ID:            store.GenerateId(),
		ProviderID:    providerId,
		AccountID:     accountId,
		ProviderType:  providerType,
		SessionKey:    store.GenerateId() + "-" + store.GenerateId(),
		CreatedAt:     time.Now().Unix(),
		LastActiveAt:  time.Now().Unix() * 1000,
		Status:        "active",
		Credentials:   credentials,
	}

	err := m.storeManager.AddSession(session)
	if err != nil {
		m.logger.Error("Failed to create session", logger.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	m.logger.Info("Session created",
		logger.Field{Key: "sessionId", Value: session.ID},
		logger.Field{Key: "providerId", Value: providerId},
	)

	return session, nil
}

// RefreshSession 刷新会话
func (m *Manager) RefreshSession(sessionId string) (*types.SessionRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessions := m.storeManager.GetAllSessions()

	for i, s := range sessions {
		if s.ID == sessionId {
			sessions[i].LastActiveAt = time.Now().Unix() * 1000
			return &sessions[i], nil
		}
	}

	return nil, nil
}

// DeleteSession 删除会话
func (m *Manager) DeleteSession(sessionId string) error {
	return m.storeManager.DeleteSession(sessionId)
}

// ClearAllSessions 清除所有会话
func (m *Manager) ClearAllSessions() error {
	return m.storeManager.ClearAllSessions()
}

// GetSessionById 根据 ID 获取会话
func (m *Manager) GetSessionById(sessionId string) *types.SessionRecord {
	sessions := m.storeManager.GetAllSessions()

	for _, s := range sessions {
		if s.ID == sessionId {
			return &s
		}
	}

	return nil
}

// GetSessionByKey 根据 SessionKey 获取会话
func (m *Manager) GetSessionByKey(key string) *types.SessionRecord {
	sessions := m.storeManager.GetAllSessions()

	for _, s := range sessions {
		if s.SessionKey == key {
			return &s
		}
	}

	return nil
}

// CountActiveSessions 统计活跃会话数量
func (m *Manager) CountActiveSessions() int {
	return len(m.storeManager.GetActiveSessions())
}

// UpdateSessionStatus 更新会话状态
func (m *Manager) UpdateSessionStatus(sessionId string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessions := m.storeManager.GetAllSessions()

	for _, s := range sessions {
		if s.ID == sessionId {
			s.Status = status
			break
		}
	}

	return nil
}

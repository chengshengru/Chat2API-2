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
	sessions     map[string]*types.SessionRecord
}

// NewManager 创建新的会话管理器
func NewManager(sm *store.StoreManager, l *logger.Logger) *Manager {
	m := &Manager{
		logger:       l,
		storeManager: sm,
		sessions:     make(map[string]*types.SessionRecord),
	}

	l.Info("Session manager initialized")
	return m
}

// GetActiveSessions 获取活跃会话
func (m *Manager) GetActiveSessions() []*types.SessionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*types.SessionRecord, 0)
	for _, s := range m.sessions {
		if s.ID != "" {
			active = append(active, s)
		}
	}
	return active
}

// CreateSession 创建新会话
func (m *Manager) CreateSession(model string, messages []types.ChatMessage) *types.SessionRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().Unix()
	session := &types.SessionRecord{
		ID:        store.GenerateSessionID(),
		Model:     model,
		Messages:  messages,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m.sessions[session.ID] = session

	// 同时保存到 store
	m.storeManager.CreateSession(model, messages)

	m.logger.Info("Session created",
		logger.Field{Key: "sessionId", Value: session.ID},
		logger.Field{Key: "model", Value: model},
	)

	return session
}

// GetSession 获取会话
func (m *Manager) GetSession(sessionId string) (*types.SessionRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if session, ok := m.sessions[sessionId]; ok {
		return session, nil
	}

	// 尝试从 store 获取
	return m.storeManager.GetSession(sessionId)
}

// UpdateSession 更新会话
func (m *Manager) UpdateSession(sessionId string, messages []types.ChatMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[sessionId]; ok {
		session.Messages = messages
		session.UpdatedAt = time.Now().Unix()
	}

	// 同时更新 store
	return m.storeManager.UpdateSession(sessionId, messages)
}

// DeleteSession 删除会话
func (m *Manager) DeleteSession(sessionId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionId)

	// 同时删除 store 中的记录
	return m.storeManager.DeleteSession(sessionId)
}

// ClearAllSessions 清除所有会话
func (m *Manager) ClearAllSessions() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions = make(map[string]*types.SessionRecord)
	return nil
}

// CountActiveSessions 统计活跃会话数量
func (m *Manager) CountActiveSessions() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// AddMessagesToSession 向会话添加消息
func (m *Manager) AddMessagesToSession(sessionId string, messages []types.ChatMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[sessionId]; ok {
		session.Messages = append(session.Messages, messages...)
		session.UpdatedAt = time.Now().Unix()
	}

	return nil
}

// GetSessionMessages 获取会话消息
func (m *Manager) GetSessionMessages(sessionId string) ([]types.ChatMessage, error) {
	session, err := m.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	return session.Messages, nil
}

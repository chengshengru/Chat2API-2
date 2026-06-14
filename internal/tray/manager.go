package tray

import (
	"runtime"
	"sync"

	"chat2api-wails/internal/logger"
)

// Manager 托盘管理器
type Manager struct {
	mu      sync.RWMutex
	logger  *logger.Logger
	visible bool
	enabled bool
}

// NewManager 创建新的托盘管理器
func NewManager(l *logger.Logger) *Manager {
	m := &Manager{
		logger:  l,
		visible: false,
		enabled: true,
	}

	// 在 Linux 下，默认禁用托盘以避免缺少依赖库
	if runtime.GOOS == "linux" {
		m.enabled = false
	}

	l.Info("Tray manager initialized")
	return m
}

// Show 显示托盘图标
func (m *Manager) Show() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		m.logger.Info("Tray disabled on this platform")
		return false
	}

	m.visible = true
	m.logger.Info("Tray shown")
	return true
}

// Hide 隐藏托盘图标
func (m *Manager) Hide() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.visible = false
	m.logger.Info("Tray hidden")
	return true
}

// IsVisible 是否可见
func (m *Manager) IsVisible() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.visible
}

// IsEnabled 是否启用
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.enabled
}

// SetTooltip 设置提示
func (m *Manager) SetTooltip(tooltip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.visible {
		return false
	}

	m.logger.Info("Tray tooltip set", logger.Field{Key: "tooltip", Value: tooltip})
	return true
}

// ShowMessage 显示消息
func (m *Manager) ShowMessage(title, message string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.visible {
		return false
	}

	m.logger.Info("Tray message",
		logger.Field{Key: "title", Value: title},
		logger.Field{Key: "message", Value: message},
	)

	return true
}

// Destroy 销毁托盘
func (m *Manager) Destroy() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.visible = false
	m.logger.Info("Tray destroyed")
}

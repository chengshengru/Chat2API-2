package window

import (
	"fmt"
	"sync"

	"chat2api-wails/internal/logger"
)

// Manager 窗口管理器
type Manager struct {
	mu       sync.RWMutex
	logger   *logger.Logger
	isOpen   bool
	window   interface{}
}

// NewManager 创建新的窗口管理器
func NewManager(l *logger.Logger) *Manager {
	m := &Manager{
		logger: l,
		isOpen: false,
	}

	l.Info("Window manager initialized")
	return m
}

// SetWindow 设置 Wails 主窗口引用
func (m *Manager) SetWindow(w interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.window = w
	m.isOpen = true
}

// Show 显示窗口
func (m *Manager) Show() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isOpen = true
	m.logger.Info("Window shown")
	return true
}

// Hide 隐藏窗口
func (m *Manager) Hide() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isOpen = false
	m.logger.Info("Window hidden")
	return true
}

// Minimize 最小化窗口
func (m *Manager) Minimize() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Window minimized")
	return true
}

// Maximize 最大化窗口
func (m *Manager) Maximize() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Window maximized")
	return true
}

// Toggle 切换窗口显示状态
func (m *Manager) Toggle() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isOpen = !m.isOpen
	m.logger.Info("Window toggled", logger.Field{Key: "isOpen", Value: fmt.Sprintf("%t", m.isOpen)})
	return true
}

// IsOpen 是否打开
func (m *Manager) IsOpen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.isOpen
}

// SetTitle 设置标题
func (m *Manager) SetTitle(title string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Window title set", logger.Field{Key: "title", Value: title})
	return true
}

// Navigate 导航到指定路径
func (m *Manager) Navigate(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Window navigate", logger.Field{Key: "path", Value: path})
	return true
}

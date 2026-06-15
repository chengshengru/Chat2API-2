package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"chat2api-wails/internal/logger"
	"chat2api-wails/internal/proxy"
	"chat2api-wails/internal/store"
	"chat2api-wails/internal/types"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// App 应用主结构
type App struct {
	ctx              context.Context
	logger           *logger.Logger
	store            *store.StoreManager
	proxyServer      *proxy.ProxyServer
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{}
}

// startup 应用启动回调
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 获取数据目录
	userDataDir, err := getUserDataDir()
	if err != nil {
		log.Printf("Failed to get user data dir: %v", err)
		userDataDir = "./data"
	}

	// 确保目录存在
	os.MkdirAll(userDataDir, 0755)

	// 初始化日志系统
	logDir := filepath.Join(userDataDir, "logs")
	os.MkdirAll(logDir, 0755)
	a.logger = logger.NewWithDir(logDir)

	a.logger.Info("Application starting",
		logger.Field{Key: "version", Value: "1.4.0"},
		logger.Field{Key: "platform", Value: runtime.GOOS},
		logger.Field{Key: "arch", Value: runtime.GOARCH},
	)

	// 初始化存储管理器
	dataDir := filepath.Join(userDataDir, "store")
	os.MkdirAll(dataDir, 0755)
	a.store, err = store.NewStoreManager(dataDir, a.logger)
	if err != nil {
		a.logger.Error("Failed to initialize store manager", logger.Field{Key: "error", Value: err.Error()})
		log.Printf("Failed to initialize store: %v", err)
		return
	}

	// 初始化代理服务器
	a.proxyServer = proxy.NewServer(a.store, a.logger)

	// 启动代理服务器
	config := a.store.GetConfig()
	if config.ProxyConfig.Enabled {
		go func() {
			time.Sleep(500 * time.Millisecond)
			a.proxyServer.Start(config.ProxyConfig.Port, config.ProxyConfig.Host)
		}()
	}

	a.logger.Info("Application started successfully")
}

// domReady DOM准备就绪
func (a *App) domReady(ctx context.Context) {
	a.logger.Info("DOM ready")
}

// beforeClose 窗口关闭前
func (a *App) beforeClose(ctx context.Context) bool {
	a.logger.Info("Application closing")
	if a.proxyServer != nil {
		a.proxyServer.Stop()
	}
	return false
}

// shutdown 应用关闭
func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("Application shutdown")
}

// ==================== Wails 绑定方法 ====================

// GetConfig 获取配置
func (a *App) GetConfig() types.AppConfig {
	if a.store == nil {
		return types.AppConfig{}
	}
	return a.store.GetConfig()
}

// UpdateConfig 更新配置
func (a *App) UpdateConfig(config types.AppConfig) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}

	// 如果代理配置变更，重启代理服务器
	oldConfig := a.store.GetConfig()
	if oldConfig.ProxyConfig.Port != config.ProxyConfig.Port ||
		oldConfig.ProxyConfig.Host != config.ProxyConfig.Host ||
		oldConfig.ProxyConfig.Enabled != config.ProxyConfig.Enabled {

		if a.proxyServer != nil {
			a.proxyServer.Stop()
		}

		if config.ProxyConfig.Enabled {
			go func() {
				time.Sleep(500 * time.Millisecond)
				a.proxyServer.Start(config.ProxyConfig.Port, config.ProxyConfig.Host)
			}()
		}
	}

	return a.store.UpdateConfig(config)
}

// GetAllProviders 获取所有提供商
func (a *App) GetAllProviders() []types.Provider {
	if a.store == nil {
		return []types.Provider{}
	}
	return a.store.GetAllProviders()
}

// GetProvider 获取提供商
func (a *App) GetProvider(id string) (*types.Provider, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetProviderByID(id)
}

// CreateProvider 创建提供商
func (a *App) CreateProvider(provider types.Provider) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.CreateProvider(provider)
}

// UpdateProvider 更新提供商
func (a *App) UpdateProvider(id string, updates map[string]interface{}) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.UpdateProvider(id, updates)
}

// DeleteProvider 删除提供商
func (a *App) DeleteProvider(id string) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.DeleteProvider(id)
}

// GetAllAccounts 获取所有账户
func (a *App) GetAllAccounts() []types.Account {
	if a.store == nil {
		return []types.Account{}
	}
	return a.store.GetAllAccounts()
}

// GetAccountsByProvider 获取提供商账户
func (a *App) GetAccountsByProvider(providerID string) []types.Account {
	if a.store == nil {
		return []types.Account{}
	}
	return a.store.GetAccountsByProviderID(providerID, false)
}

// CreateAccount 创建账户
func (a *App) CreateAccount(providerID, name string, credentials map[string]string) (*types.Account, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.CreateAccount(providerID, name, credentials)
}

// UpdateAccount 更新账户
func (a *App) UpdateAccount(id string, updates map[string]interface{}) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.UpdateAccount(id, updates)
}

// DeleteAccount 删除账户
func (a *App) DeleteAccount(id string) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.DeleteAccount(id)
}

// GetProxyStatus 获取代理状态
func (a *App) GetProxyStatus() *types.ProxyStatus {
	if a.proxyServer == nil {
		return &types.ProxyStatus{}
	}
	return a.proxyServer.GetStatus()
}

// GetStatistics 获取统计信息
func (a *App) GetStatistics() types.Statistics {
	if a.store == nil {
		return types.Statistics{}
	}
	return a.store.GetStatistics()
}

// StartProxy 启动代理
func (a *App) StartProxy(port int, host string) bool {
	if a.proxyServer == nil {
		return false
	}
	return a.proxyServer.Start(port, host)
}

// StopProxy 停止代理
func (a *App) StopProxy() bool {
	if a.proxyServer == nil {
		return false
	}
	return a.proxyServer.Stop()
}

// GetRequestLogs 获取请求日志
func (a *App) GetRequestLogs(limit, offset int) []*types.RequestLogEntry {
	if a.store == nil {
		return []*types.RequestLogEntry{}
	}
	return a.store.GetRequestLogs(limit, offset)
}

// GetModelMappings 获取模型映射
func (a *App) GetModelMappings() []types.ModelMapping {
	if a.store == nil {
		return []types.ModelMapping{}
	}
	return a.store.GetAllModelMappings()
}

// CreateModelMapping 创建模型映射
func (a *App) CreateModelMapping(mapping types.ModelMapping) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.CreateModelMapping(mapping)
}

// UpdateModelMapping 更新模型映射
func (a *App) UpdateModelMapping(id string, mappings []types.ModelMappingEntry) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.UpdateModelMapping(id, mappings)
}

// GetSessions 获取会话
func (a *App) GetSessions() []types.SessionRecord {
	return []types.SessionRecord{}
}

// CreateSession 创建会话
func (a *App) CreateSession(model string, messages []types.ChatMessage) *types.SessionRecord {
	if a.store == nil {
		return nil
	}
	return a.store.CreateSession(model, messages)
}

// GetSession 获取会话
func (a *App) GetSession(id string) (*types.SessionRecord, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetSession(id)
}

// UpdateSession 更新会话
func (a *App) UpdateSession(id string, messages []types.ChatMessage) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.UpdateSession(id, messages)
}

// DeleteSession 删除会话
func (a *App) DeleteSession(id string) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.DeleteSession(id)
}

// GetLogs 获取日志
func (a *App) GetLogs(options logger.LogOptions) []logger.LogEntry {
	if a.logger == nil {
		return []logger.LogEntry{}
	}
	return a.logger.GetLogs(logger.LogFilter{
		Level:  logger.LogLevel(options.Level),
		Limit:  options.Limit,
		Offset: options.Offset,
	})
}

// ClearLogs 清除日志
func (a *App) ClearLogs() {
	if a.logger != nil {
		a.logger.Clear()
	}
}

// ==================== 辅助函数 ====================

func getUserDataDir() (string, error) {
	platform := runtime.GOOS
	switch platform {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Chat2API"), nil
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Chat2API"), nil
	default: // linux
		return filepath.Join(os.Getenv("HOME"), ".config", "Chat2API"), nil
	}
}

// ==================== 静态资源 ====================

//go:embed frontend/dist
var assets embed.FS

// ==================== 应用入口 ====================

func main() {
	// 创建应用实例
	app := NewApp()

	// 创建 Wails 应用
	wa := &options.App{
		Title:            "Chat2API",
		Width:            1200,
		Height:           800,
		MinWidth:         800,
		MinHeight:        600,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			app,
		},
		// macOS 特定配置
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
		},
		// Windows 特定配置
		Windows: &windows.Options{
			WebviewIsTransparent: false,
		},
	}

	// 运行应用
	if err := wails.Run(wa); err != nil {
		log.Fatal(err)
	}
}

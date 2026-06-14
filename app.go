package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"chat2api-wails/internal/logger"
	"chat2api-wails/internal/oauth"
	"chat2api-wails/internal/proxy"
	"chat2api-wails/internal/session"
	"chat2api-wails/internal/store"
	"chat2api-wails/internal/tray"
	"chat2api-wails/internal/types"
	"chat2api-wails/internal/window"
)

// App 应用程序核心
type App struct {
	mu             sync.RWMutex
	ctx            context.Context
	logger         *logger.Logger
	storeManager   *store.StoreManager
	proxyServer    *proxy.ProxyServer
	oauthManager   *oauth.Manager
	sessionManager *session.Manager
	trayManager    *tray.Manager
	windowManager  *window.Manager
}

// NewApp 创建新的应用
func NewApp() *App {
	appLogger := logger.New()

	return &App{
		logger: appLogger,
	}
}

// startup 启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger = logger.New()

	// 初始化管理器
	a.storeManager = store.NewManager(a.logger)
	a.proxyServer = proxy.NewServer(a.storeManager, a.logger)
	a.oauthManager = oauth.NewManager(a.logger)
	a.sessionManager = session.NewManager(a.storeManager, a.logger)
	a.trayManager = tray.NewManager(a.logger)
	a.windowManager = window.NewManager(a.logger)

	// 自动启动代理服务器
	config := a.storeManager.GetConfig()
	if config.AutoStartProxy {
		a.proxyServer.Start(config.ProxyPort, config.ProxyHost)
	}

	a.logger.Info("Application started")
}

// domReady DOM 准备好时调用
func (a *App) domReady(ctx context.Context) {
	a.logger.Info("DOM ready")
}

// beforeClose 关闭前调用
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.logger.Info("Application closing")

	// 停止代理服务器
	a.proxyServer.Stop()

	return false
}

// shutdown 关机时调用
func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("Application shutdown complete")
}

// ==================== 日志相关 API ====================

// GetLogs 获取日志
func (a *App) GetLogs(options logger.LogOptions) *logger.PaginatedResult {
	return a.logger.GetLogsPaginated(options)
}

// GetLogsByLevel 根据级别获取日志
func (a *App) GetLogsByLevel(level string, limit int) []logger.LogEntry {
	return a.logger.GetLogsByLevel(level, limit)
}

// GetErrorLogs 获取错误日志
func (a *App) GetErrorLogs(limit int) []logger.LogEntry {
	return a.logger.GetErrorLogs(limit)
}

// GetRecentLogs 获取最近日志
func (a *App) GetRecentLogs(limit int) []logger.LogEntry {
	return a.logger.GetRecentLogs(limit)
}

// ClearLogs 清除日志
func (a *App) ClearLogs() bool {
	return a.logger.ClearLogs()
}

// ExportLogs 导出日志
func (a *App) ExportLogs(filename string) string {
	return a.logger.ExportLogs("txt")
}

// GetStatistics 获取统计
func (a *App) GetStatistics() logger.PersistentStatistics {
	return a.storeManager.GetStatistics()
}

// ==================== 提供商 API ====================

// GetProviders 获取所有提供商
func (a *App) GetProviders() []types.Provider {
	return a.storeManager.GetProviders()
}

// GetProviderById 根据 ID 获取提供商
func (a *App) GetProviderById(id string) *types.Provider {
	return a.storeManager.GetProviderById(id)
}

// AddProvider 添加提供商
func (a *App) AddProvider(provider *types.Provider) (string, error) {
	if provider.ID == "" {
		provider.ID = store.GenerateId()
	}
	err := a.storeManager.AddProvider(provider)
	if err != nil {
		return "", err
	}
	a.logger.Info("Provider added", logger.Field{Key: "id", Value: provider.ID}, logger.Field{Key: "name", Value: provider.Name})
	return provider.ID, nil
}

// UpdateProvider 更新提供商
func (a *App) UpdateProvider(id string, updates *types.Provider) error {
	_, err := a.storeManager.UpdateProvider(id, updates)
	if err != nil {
		return err
	}
	a.logger.Info("Provider updated", logger.Field{Key: "id", Value: id})
	return nil
}

// DeleteProvider 删除提供商
func (a *App) DeleteProvider(id string) error {
	err := a.storeManager.DeleteProvider(id)
	if err != nil {
		return err
	}
	a.logger.Info("Provider deleted", logger.Field{Key: "id", Value: id})
	return nil
}

// ==================== 账户 API ====================

// GetAccounts 获取所有账户
func (a *App) GetAccounts(includeCredentials bool) []types.Account {
	return a.storeManager.GetAccounts(includeCredentials)
}

// GetAccountById 根据 ID 获取账户
func (a *App) GetAccountById(id string, includeCredentials bool) *types.Account {
	return a.storeManager.GetAccountById(id, includeCredentials)
}

// GetAccountsByProviderId 获取指定提供商的账户
func (a *App) GetAccountsByProviderId(providerId string, includeCredentials bool) []types.Account {
	return a.storeManager.GetAccountsByProviderId(providerId, includeCredentials)
}

// AddAccount 添加账户
func (a *App) AddAccount(account *types.Account) (string, error) {
	if account.ID == "" {
		account.ID = store.GenerateId()
	}
	err := a.storeManager.AddAccount(account)
	if err != nil {
		return "", err
	}
	a.logger.Info("Account added", logger.Field{Key: "id", Value: account.ID}, logger.Field{Key: "providerId", Value: account.ProviderID})
	return account.ID, nil
}

// UpdateAccount 更新账户
func (a *App) UpdateAccount(id string, updates *types.Account) error {
	_, err := a.storeManager.UpdateAccount(id, updates)
	if err != nil {
		return err
	}
	a.logger.Info("Account updated", logger.Field{Key: "id", Value: id})
	return nil
}

// DeleteAccount 删除账户
func (a *App) DeleteAccount(id string) error {
	err := a.storeManager.DeleteAccount(id)
	if err != nil {
		return err
	}
	a.logger.Info("Account deleted", logger.Field{Key: "id", Value: id})
	return nil
}

// UpdateAccountCredentials 更新账户凭证
func (a *App) UpdateAccountCredentials(id string, credentials map[string]string) error {
	return a.storeManager.UpdateAccountCredentials(id, credentials)
}

// ==================== 配置 API ====================

// GetConfig 获取配置
func (a *App) GetConfig() *types.AppConfig {
	return a.storeManager.GetConfig()
}

// UpdateConfig 更新配置
func (a *App) UpdateConfig(updates map[string]interface{}) error {
	_, err := a.storeManager.UpdateConfig(updates)
	if err != nil {
		return err
	}

	// 如果端口或主机改变，重新启动代理
	if _, hasPort := updates["serverPort"]; hasPort {
		if status := a.proxyServer.GetStatus(); status.IsRunning {
			config := a.storeManager.GetConfig()
			a.proxyServer.Stop()
			a.proxyServer.Start(config.ProxyPort, config.ProxyHost)
		}
	}

	a.logger.Info("Config updated")
	return nil
}

// UpdateConfigStruct 更新完整配置
func (a *App) UpdateConfigStruct(config types.AppConfig) error {
	return a.storeManager.UpdateConfigStruct(config)
}

// GenerateApiKey 生成新的 API Key
func (a *App) GenerateApiKey(name string) types.ApiKey {
	return a.storeManager.GenerateApiKey(name)
}

// DeleteApiKey 删除 API Key
func (a *App) DeleteApiKey(id string) error {
	return a.storeManager.DeleteApiKey(id)
}

// ==================== 代理服务器 API ====================

// StartProxy 启动代理
func (a *App) StartProxy(port int, host string) bool {
	return a.proxyServer.Start(port, host)
}

// StopProxy 停止代理
func (a *App) StopProxy() bool {
	return a.proxyServer.Stop()
}

// GetProxyStatus 获取代理状态
func (a *App) GetProxyStatus() *proxy.ProxyStatus {
	return a.proxyServer.GetStatus()
}

// GetProxyStatistics 获取代理统计
func (a *App) GetProxyStatistics() *proxy.Statistics {
	return a.proxyServer.GetStatistics()
}

// ==================== OAuth API ====================

// GetOAuthAuthURL 获取认证 URL
func (a *App) GetOAuthAuthURL(providerType string) (string, error) {
	return a.oauthManager.GetAuthURL(providerType)
}

// StartOAuthLogin 开始 OAuth 登录
func (a *App) StartOAuthLogin(providerId string, providerType string) *oauth.OAuthResult {
	return a.oauthManager.StartLogin(providerId, providerType)
}

// OAuthLoginWithToken 使用 Token 登录
func (a *App) OAuthLoginWithToken(providerType string, token string) *oauth.OAuthResult {
	return a.oauthManager.LoginWithToken(providerType, token)
}

// ValidateOAuthToken 验证 Token
func (a *App) ValidateOAuthToken(providerType string, credentials map[string]string) *oauth.TokenValidationResult {
	return a.oauthManager.ValidateToken(providerType, credentials)
}

// RefreshOAuthToken 刷新 Token
func (a *App) RefreshOAuthToken(providerType string, credentials map[string]string) *oauth.TokenValidationResult {
	return a.oauthManager.RefreshToken(providerType, credentials)
}

// HandleOAuthCallback 处理回调
func (a *App) HandleOAuthCallback(providerType string, callbackURL string) *oauth.OAuthResult {
	return a.oauthManager.HandleCallback(providerType, callbackURL)
}

// GetSupportedOAuthProviders 获取支持的提供商
func (a *App) GetSupportedOAuthProviders() []string {
	return a.oauthManager.GetSupportedProviders()
}

// ==================== 会话 API ====================

// GetActiveSessions 获取活跃会话
func (a *App) GetActiveSessions() []types.SessionRecord {
	return a.sessionManager.GetActiveSessions()
}

// CreateSession 创建会话
func (a *App) CreateSession(providerId, accountId, providerType string, credentials map[string]string) (*types.SessionRecord, error) {
	return a.sessionManager.CreateSession(providerId, accountId, providerType, credentials)
}

// DeleteSession 删除会话
func (a *App) DeleteSession(sessionId string) error {
	return a.sessionManager.DeleteSession(sessionId)
}

// ClearAllSessions 清除所有会话
func (a *App) ClearAllSessions() error {
	return a.sessionManager.ClearAllSessions()
}

// ==================== 窗口控制 API ====================

// ShowWindow 显示窗口
func (a *App) ShowWindow() bool {
	return a.windowManager.Show()
}

// HideWindow 隐藏窗口
func (a *App) HideWindow() bool {
	return a.windowManager.Hide()
}

// MinimizeWindow 最小化
func (a *App) MinimizeWindow() bool {
	return a.windowManager.Minimize()
}

// ToggleWindow 切换显示
func (a *App) ToggleWindow() bool {
	return a.windowManager.Toggle()
}

// SetWindowTitle 设置标题
func (a *App) SetWindowTitle(title string) bool {
	return a.windowManager.SetTitle(title)
}

// NavigateWindow 导航
func (a *App) NavigateWindow(path string) bool {
	return a.windowManager.Navigate(path)
}

// ==================== 系统对话框 API ====================

// ShowOpenDialog 显示打开对话框
func (a *App) ShowOpenDialog(title string, filter string) (string, error) {
	options := runtime.OpenDialogOptions{
		Title: title,
	}

	if filter != "" {
		options.Filters = []runtime.FileFilter{
			{DisplayName: "Files", Pattern: filter},
		}
	}

	return runtime.OpenFileDialog(a.ctx, options)
}

// ShowSaveDialog 显示保存对话框
func (a *App) ShowSaveDialog(title string, defaultFilename string) (string, error) {
	options := runtime.SaveDialogOptions{
		Title:                      title,
		DefaultFilename:            defaultFilename,
	}

	return runtime.SaveFileDialog(a.ctx, options)
}

// ShowMessageDialog 显示消息对话框
func (a *App) ShowMessageDialog(messageType string, title string, message string) string {
	var dialogType runtime.DialogType
	switch messageType {
	case "info":
		dialogType = runtime.InfoDialog
	case "warning":
		dialogType = runtime.WarningDialog
	case "error":
		dialogType = runtime.ErrorDialog
	case "question":
		dialogType = runtime.QuestionDialog
	default:
		dialogType = runtime.InfoDialog
	}

	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    dialogType,
		Title:   title,
		Message: message,
	})

	if err != nil {
		return "error"
	}

	return result
}

// ==================== 其他实用 API ====================

// OpenExternal 在浏览器中打开 URL
func (a *App) OpenExternal(url string) bool {
	runtime.BrowserOpenURL(a.ctx, url)
	a.logger.Info("Open external URL", logger.Field{Key: "url", Value: url})
	return true
}

// GetAppInfo 获取应用信息
func (a *App) GetAppInfo() map[string]interface{} {
	config := a.storeManager.GetConfig()
	return map[string]interface{}{
		"name":         "Chat2API",
		"version":      "1.0.0",
		"description":  "Chat2API Wails 版本",
		"config":       config,
		"proxyStatus":  a.proxyServer.GetStatus(),
	}
}

// GetStorePath 获取存储路径
func (a *App) GetStorePath() string {
	return a.storeManager.GetBasePath()
}

// LogCustomMessage 自定义日志记录
func (a *App) LogCustomMessage(level string, message string, data map[string]interface{}) {
	fields := make([]logger.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, logger.Field{Key: k, Value: fmt.Sprintf("%v", v)})
	}

	switch level {
	case "debug":
		a.logger.Debug(message, fields...)
	case "warn":
		a.logger.Warn(message, fields...)
	case "error":
		a.logger.Error(message, fields...)
	case "info":
		fallthrough
	default:
		a.logger.Info(message, fields...)
	}
}

// DebugDumpState 调试：转储当前状态
func (a *App) DebugDumpState() string {
	data := map[string]interface{}{
		"config":       a.storeManager.GetConfig(),
		"providers":    a.storeManager.GetProviders(),
		"accounts":     a.storeManager.GetAccounts(false),
		"proxyStatus":  a.proxyServer.GetStatus(),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	return string(jsonBytes)
}

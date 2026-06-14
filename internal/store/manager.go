package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"chat2api-wails/internal/logger"
	"chat2api-wails/internal/types"
)

// ErrNotInitialized 表示存储未初始化
var ErrNotInitialized = errors.New("store not initialized")

// ErrProviderNotFound 表示提供商未找到
var ErrProviderNotFound = errors.New("provider not found")

// ErrAccountNotFound 表示账户未找到
var ErrAccountNotFound = errors.New("account not found")

// StoreManager 管理应用数据存储
type StoreManager struct {
	mu            sync.RWMutex
	basePath      string
	providersFile string
	accountsFile  string
	configFile    string
	sessionFile   string
	statsFile     string
	logger        *logger.Logger

	providers []types.Provider
	accounts  []types.Account
	config    types.AppConfig
	sessions  []types.SessionRecord
	stats     logger.PersistentStatistics
}

// NewManager 创建新的存储管理器
func NewManager(l *logger.Logger) *StoreManager {
	homeDir, _ := os.UserHomeDir()
	basePath := filepath.Join(homeDir, ".chat2api")
	os.MkdirAll(basePath, 0755)

	sm := &StoreManager{
		basePath:      basePath,
		providersFile: filepath.Join(basePath, "providers.json"),
		accountsFile:  filepath.Join(basePath, "accounts.json"),
		configFile:    filepath.Join(basePath, "config.json"),
		sessionFile:   filepath.Join(basePath, "sessions.json"),
		statsFile:     filepath.Join(basePath, "statistics.json"),
		logger:        l,
	}

	// 加载现有数据
	sm.loadConfig()
	sm.loadProviders()
	sm.loadAccounts()
	sm.loadSessions()
	sm.loadStatistics()

	l.Info("Store manager initialized", logger.Field{Key: "path", Value: basePath})

	return sm
}

// GetBasePath 返回存储基础路径
func (sm *StoreManager) GetBasePath() string {
	return sm.basePath
}

// ==================== Provider 操作 ====================

// GetAllProviders 返回所有提供商
func (sm *StoreManager) GetProviders() []types.Provider {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]types.Provider, len(sm.providers))
	copy(result, sm.providers)
	return result
}

// GetProviderById 根据 ID 获取提供商
func (sm *StoreManager) GetProviderById(id string) *types.Provider {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, provider := range sm.providers {
		if provider.ID == id {
			return &provider
		}
	}
	return nil
}

// AddProvider 添加提供商
func (sm *StoreManager) AddProvider(provider *types.Provider) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().Unix()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	if provider.ID == "" {
		provider.ID = GenerateId()
	}

	sm.providers = append(sm.providers, *provider)
	return sm.saveProviders()
}

// UpdateProvider 更新提供商
func (sm *StoreManager) UpdateProvider(id string, updates *types.Provider) (*types.Provider, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	found := false
	for i := range sm.providers {
		if sm.providers[i].ID == id {
			updates.ID = id
			updates.CreatedAt = sm.providers[i].CreatedAt
			updates.UpdatedAt = time.Now().Unix()
			sm.providers[i] = *updates
			found = true
			break
		}
	}

	if !found {
		return nil, ErrProviderNotFound
	}

	return updates, sm.saveProviders()
}

// DeleteProvider 删除提供商
func (sm *StoreManager) DeleteProvider(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	newProviders := make([]types.Provider, 0)
	for _, p := range sm.providers {
		if p.ID != id {
			newProviders = append(newProviders, p)
		}
	}
	sm.providers = newProviders

	// 同时删除相关账户
	newAccounts := make([]types.Account, 0)
	for _, a := range sm.accounts {
		if a.ProviderID != id {
			newAccounts = append(newAccounts, a)
		}
	}
	sm.accounts = newAccounts

	if err := sm.saveProviders(); err != nil {
		return err
	}
	return sm.saveAccounts()
}

// ==================== Account 操作 ====================

// GetAllAccounts 返回所有账户
func (sm *StoreManager) GetAccounts(includeCredentials bool) []types.Account {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]types.Account, len(sm.accounts))
	for i, a := range sm.accounts {
		result[i] = a
		if includeCredentials {
			result[i].Credentials = DecryptCredentials(a.Credentials)
		} else {
			result[i].Credentials = nil
		}
	}
	return result
}

// GetAccountById 根据 ID 获取账户
func (sm *StoreManager) GetAccountById(id string, includeCredentials bool) *types.Account {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, a := range sm.accounts {
		if a.ID == id {
			result := a
			if includeCredentials {
				result.Credentials = DecryptCredentials(a.Credentials)
			} else {
				result.Credentials = nil
			}
			return &result
		}
	}
	return nil
}

// GetAccountsByProviderId 获取指定提供商的账户
func (sm *StoreManager) GetAccountsByProviderId(providerId string, includeCredentials bool) []types.Account {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]types.Account, 0)
	for _, a := range sm.accounts {
		if a.ProviderID == providerId {
			account := a
			if includeCredentials {
				account.Credentials = DecryptCredentials(a.Credentials)
			} else {
				account.Credentials = nil
			}
			result = append(result, account)
		}
	}
	return result
}

// GetActiveAccounts 获取活跃账户
func (sm *StoreManager) GetActiveAccounts(includeCredentials bool) []types.Account {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]types.Account, 0)
	for _, a := range sm.accounts {
		if a.Status == types.AccountStatusActive {
			account := a
			if includeCredentials {
				account.Credentials = DecryptCredentials(a.Credentials)
			} else {
				account.Credentials = nil
			}
			result = append(result, account)
		}
	}
	return result
}

// AddAccount 添加账户（自动加密凭证）
func (sm *StoreManager) AddAccount(account *types.Account) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().Unix()
	account.CreatedAt = now
	account.UpdatedAt = now
	if account.ID == "" {
		account.ID = GenerateId()
	}

	// 加密凭证
	if account.Credentials != nil {
		account.Credentials = EncryptCredentials(account.Credentials)
	}

	sm.accounts = append(sm.accounts, *account)
	return sm.saveAccounts()
}

// UpdateAccount 更新账户
func (sm *StoreManager) UpdateAccount(id string, updates *types.Account) (*types.Account, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	found := false
	for i := range sm.accounts {
		if sm.accounts[i].ID == id {
			// 保留创建时间
			updates.CreatedAt = sm.accounts[i].CreatedAt
			updates.UpdatedAt = time.Now().Unix()

			// 如果有凭证，加密后存储
			if updates.Credentials != nil {
				updates.Credentials = EncryptCredentials(updates.Credentials)
			}

			updates.ID = id
			sm.accounts[i] = *updates
			found = true
			break
		}
	}

	if !found {
		return nil, ErrAccountNotFound
	}

	return updates, sm.saveAccounts()
}

// DeleteAccount 删除账户
func (sm *StoreManager) DeleteAccount(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	newAccounts := make([]types.Account, 0)
	for _, a := range sm.accounts {
		if a.ID != id {
			newAccounts = append(newAccounts, a)
		}
	}
	sm.accounts = newAccounts
	return sm.saveAccounts()
}

// UpdateAccountCredentials 更新账户凭证（用于 OAuth 回调）
func (sm *StoreManager) UpdateAccountCredentials(id string, credentials map[string]string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i := range sm.accounts {
		if sm.accounts[i].ID == id {
			sm.accounts[i].Credentials = EncryptCredentials(credentials)
			sm.accounts[i].UpdatedAt = time.Now().Unix()
			return sm.saveAccounts()
		}
	}
	return ErrAccountNotFound
}

// ==================== Config 操作 ====================

// GetConfig 获取应用配置
func (sm *StoreManager) GetConfig() *types.AppConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	config := sm.config
	return &config
}

// SetConfig 设置应用配置
func (sm *StoreManager) SetConfig(config *types.AppConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.config = *config
	return sm.saveConfig()
}

// UpdateConfig 更新部分配置
func (sm *StoreManager) UpdateConfig(updates map[string]interface{}) (*types.AppConfig, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := json.Marshal(updates)
	if err != nil {
		return nil, err
	}

	// 部分更新：将 updates 合并到当前配置
	configMap := make(map[string]interface{})
	if err := json.Unmarshal(data, &configMap); err != nil {
		return nil, err
	}

	// 将当前配置转换为 map 后合并
	currentData, _ := json.Marshal(sm.config)
	currentMap := make(map[string]interface{})
	json.Unmarshal(currentData, &currentMap)

	for k, v := range configMap {
		currentMap[k] = v
	}

	merged, _ := json.Marshal(currentMap)
	if err := json.Unmarshal(merged, &sm.config); err != nil {
		return nil, err
	}

	return &sm.config, sm.saveConfig()
}

// UpdateConfigStruct 更新完整配置对象
func (sm *StoreManager) UpdateConfigStruct(config types.AppConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.config = config
	return sm.saveConfig()
}

// GenerateApiKey 生成新的 API Key
func (sm *StoreManager) GenerateApiKey(name string) types.ApiKey {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := GenerateId() + "-" + GenerateId()

	apiKey := types.ApiKey{
		ID:         GenerateId(),
		Name:       name,
		Key:        key,
		Enabled:    true,
		CreatedAt:  time.Now().Unix(),
		UsageCount: 0,
	}

	sm.config.ApiKeys = append(sm.config.ApiKeys, apiKey)
	sm.saveConfig()
	return apiKey
}

// UpdateApiKeyUsage 更新 API Key 使用统计
func (sm *StoreManager) UpdateApiKeyUsage(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, k := range sm.config.ApiKeys {
		if k.Key == key && k.Enabled {
			sm.config.ApiKeys[i].LastUsedAt = time.Now().Unix()
			sm.config.ApiKeys[i].UsageCount++
			break
		}
	}
	sm.saveConfig()
}

// DeleteApiKey 删除 API Key
func (sm *StoreManager) DeleteApiKey(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	newKeys := make([]types.ApiKey, 0)
	for _, k := range sm.config.ApiKeys {
		if k.ID != id {
			newKeys = append(newKeys, k)
		}
	}
	sm.config.ApiKeys = newKeys
	return sm.saveConfig()
}

// ==================== Session 操作 ====================

// GetAllSessions 获取所有会话
func (sm *StoreManager) GetAllSessions() []types.SessionRecord {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]types.SessionRecord, len(sm.sessions))
	copy(result, sm.sessions)
	return result
}

// GetActiveSessions 获取活跃会话
func (sm *StoreManager) GetActiveSessions() []types.SessionRecord {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	timeoutSec := sm.config.SessionConfig.SessionTimeout
	if timeoutSec == 0 {
		timeoutSec = 3600 // 默认 1 小时
	}
	timeoutMs := timeoutSec * 1000
	now := time.Now().Unix() * 1000

	result := make([]types.SessionRecord, 0)
	for _, s := range sm.sessions {
		if s.Status == "active" && (now-s.LastActiveAt) < int64(timeoutMs) {
			result = append(result, s)
		}
	}
	return result
}

// AddSession 添加会话
func (sm *StoreManager) AddSession(session *types.SessionRecord) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session.ID == "" {
		session.ID = GenerateId()
	}

	sm.sessions = append(sm.sessions, *session)
	return sm.saveSessions()
}

// DeleteSession 删除会话
func (sm *StoreManager) DeleteSession(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	newSessions := make([]types.SessionRecord, 0)
	for _, s := range sm.sessions {
		if s.ID != id {
			newSessions = append(newSessions, s)
		}
	}
	sm.sessions = newSessions
	return sm.saveSessions()
}

// ClearAllSessions 清空所有会话
func (sm *StoreManager) ClearAllSessions() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions = nil
	return sm.saveSessions()
}

// ==================== Statistics 操作 ====================

// GetStatistics 获取统计信息
func (sm *StoreManager) GetStatistics() logger.PersistentStatistics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := sm.stats
	stats.LastUpdated = time.Now().Unix()
	return stats
}

// RecordRequest 记录一次请求
func (sm *StoreManager) RecordRequest(success bool, latency int64, model, providerId, accountId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.stats.ModelUsage == nil {
		sm.stats.ModelUsage = make(map[string]string)
	}
	if sm.stats.ProviderUsage == nil {
		sm.stats.ProviderUsage = make(map[string]string)
	}
	if sm.stats.AccountUsage == nil {
		sm.stats.AccountUsage = make(map[string]string)
	}
	if sm.stats.DailyStats == nil {
		sm.stats.DailyStats = make(map[string]string)
	}

	sm.stats.TotalRequests++
	if success {
		sm.stats.SuccessRequests++
	} else {
		sm.stats.FailedRequests++
	}
	sm.stats.TotalLatency += latency

	if model != "" {
		curr, _ := strconv.ParseInt(sm.stats.ModelUsage[model], 10, 64)
		sm.stats.ModelUsage[model] = strconv.FormatInt(curr+1, 10)
	}
	if providerId != "" {
		curr, _ := strconv.ParseInt(sm.stats.ProviderUsage[providerId], 10, 64)
		sm.stats.ProviderUsage[providerId] = strconv.FormatInt(curr+1, 10)
	}
	if accountId != "" {
		curr, _ := strconv.ParseInt(sm.stats.AccountUsage[accountId], 10, 64)
		sm.stats.AccountUsage[accountId] = strconv.FormatInt(curr+1, 10)
	}

	// 每日统计
	today := time.Now().Format("2006-01-02")
	if sm.stats.DailyStats[today] == "" {
		data, _ := json.Marshal(map[string]int64{
			"total":   0,
			"success": 0,
			"failed":  0,
			"latency": 0,
		})
		sm.stats.DailyStats[today] = string(data)
	}

	// 解析并更新每日统计
	dailyMap := map[string]int64{}
	json.Unmarshal([]byte(sm.stats.DailyStats[today]), &dailyMap)
	dailyMap["total"]++
	if success {
		dailyMap["success"]++
	} else {
		dailyMap["failed"]++
	}
	dailyMap["latency"] += latency
	newData, _ := json.Marshal(dailyMap)
	sm.stats.DailyStats[today] = string(newData)

	sm.stats.LastUpdated = time.Now().Unix()

	// 定期清理旧数据
	cutoff := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	for date := range sm.stats.DailyStats {
		if date < cutoff {
			delete(sm.stats.DailyStats, date)
		}
	}

	sm.saveStatistics()
}

// ==================== 持久化操作 ====================

func (sm *StoreManager) loadConfig() {
	data, err := os.ReadFile(sm.configFile)
	if err != nil {
		sm.config = types.DefaultAppConfig()
		return
	}

	if err := json.Unmarshal(data, &sm.config); err != nil {
		sm.logger.Error("Failed to parse config", logger.Field{Key: "error", Value: err.Error()})
		sm.config = types.DefaultAppConfig()
	}

	// 确保 ApiKeys 切片已初始化
	if sm.config.ApiKeys == nil {
		sm.config.ApiKeys = make([]types.ApiKey, 0)
	}
}

func (sm *StoreManager) saveConfig() error {
	data, err := json.MarshalIndent(sm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(sm.configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

func (sm *StoreManager) loadProviders() {
	data, err := os.ReadFile(sm.providersFile)
	if err != nil {
		sm.providers = make([]types.Provider, 0)
		return
	}

	if err := json.Unmarshal(data, &sm.providers); err != nil {
		sm.logger.Error("Failed to parse providers", logger.Field{Key: "error", Value: err.Error()})
		sm.providers = make([]types.Provider, 0)
	}
}

func (sm *StoreManager) saveProviders() error {
	data, err := json.MarshalIndent(sm.providers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal providers: %w", err)
	}

	if err := os.WriteFile(sm.providersFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write providers: %w", err)
	}
	return nil
}

func (sm *StoreManager) loadAccounts() {
	data, err := os.ReadFile(sm.accountsFile)
	if err != nil {
		sm.accounts = make([]types.Account, 0)
		return
	}

	if err := json.Unmarshal(data, &sm.accounts); err != nil {
		sm.logger.Error("Failed to parse accounts", logger.Field{Key: "error", Value: err.Error()})
		sm.accounts = make([]types.Account, 0)
	}
}

func (sm *StoreManager) saveAccounts() error {
	data, err := json.MarshalIndent(sm.accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	if err := os.WriteFile(sm.accountsFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write accounts: %w", err)
	}
	return nil
}

func (sm *StoreManager) loadSessions() {
	data, err := os.ReadFile(sm.sessionFile)
	if err != nil {
		sm.sessions = make([]types.SessionRecord, 0)
		return
	}

	if err := json.Unmarshal(data, &sm.sessions); err != nil {
		sm.logger.Error("Failed to parse sessions", logger.Field{Key: "error", Value: err.Error()})
		sm.sessions = make([]types.SessionRecord, 0)
	}
}

func (sm *StoreManager) saveSessions() error {
	data, err := json.MarshalIndent(sm.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	if err := os.WriteFile(sm.sessionFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write sessions: %w", err)
	}
	return nil
}

func (sm *StoreManager) loadStatistics() {
	data, err := os.ReadFile(sm.statsFile)
	if err != nil {
		sm.stats = logger.PersistentStatistics{
			LastUpdated:   time.Now().Unix(),
			ModelUsage:    make(map[string]string),
			ProviderUsage: make(map[string]string),
			AccountUsage:  make(map[string]string),
			DailyStats:    make(map[string]string),
		}
		return
	}

	if err := json.Unmarshal(data, &sm.stats); err != nil {
		sm.logger.Error("Failed to parse statistics", logger.Field{Key: "error", Value: err.Error()})
		sm.stats = logger.PersistentStatistics{
			LastUpdated:   time.Now().Unix(),
			ModelUsage:    make(map[string]string),
			ProviderUsage: make(map[string]string),
			AccountUsage:  make(map[string]string),
			DailyStats:    make(map[string]string),
		}
	}

	// 确保所有 map 不为 nil
	if sm.stats.ModelUsage == nil {
		sm.stats.ModelUsage = make(map[string]string)
	}
	if sm.stats.ProviderUsage == nil {
		sm.stats.ProviderUsage = make(map[string]string)
	}
	if sm.stats.AccountUsage == nil {
		sm.stats.AccountUsage = make(map[string]string)
	}
	if sm.stats.DailyStats == nil {
		sm.stats.DailyStats = make(map[string]string)
	}
}

func (sm *StoreManager) saveStatistics() error {
	data, err := json.MarshalIndent(sm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	if err := os.WriteFile(sm.statsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write statistics: %w", err)
	}
	return nil
}

// FlushPendingWrites 确保所有待写操作已刷新到磁盘
// 在 Wails 中，所有操作都是同步的，此函数仅作占位符
func (sm *StoreManager) FlushPendingWrites() {
	// 所有操作都是同步写入，无需额外操作
}

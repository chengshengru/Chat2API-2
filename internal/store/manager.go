package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
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

// ErrProviderNotFound 提供商未找到
var ErrProviderNotFound = errors.New("provider not found")

// ErrAccountNotFound 账户未找到
var ErrAccountNotFound = errors.New("account not found")

// StoreManager 存储管理器
type StoreManager struct {
	mu         sync.RWMutex
	dataDir    string
	encryption []byte

	// 内存缓存
	config         *types.AppConfig
	providers      map[string]*types.Provider
	accounts       map[string]*types.Account
	modelMappings  map[string]*types.ModelMapping
	sessions       map[string]*types.SessionRecord
	requestLogs    []*types.RequestLogEntry

	// 统计信息
	stats types.Statistics

	// 日志器
	logger *logger.Logger

	// 文件路径
	configFile      string
	providersFile    string
	accountsFile     string
	modelMappingsFile string
	sessionsFile     string
	statsFile        string
	requestLogsFile   string
}

// NewStoreManager 创建存储管理器
func NewStoreManager(dataDir string, log *logger.Logger) (*StoreManager, error) {
	sm := &StoreManager{
		dataDir:    dataDir,
		encryption: make([]byte, 32), // 32 bytes for AES-256
		providers:  make(map[string]*types.Provider),
		accounts:   make(map[string]*types.Account),
		modelMappings: make(map[string]*types.ModelMapping),
		sessions:  make(map[string]*types.SessionRecord),
		logger:    log,
	}

	// 生成或加载加密密钥
	if err := sm.loadOrCreateEncryptionKey(); err != nil {
		return nil, fmt.Errorf("failed to load encryption key: %w", err)
	}

	// 设置文件路径
	sm.configFile = filepath.Join(dataDir, "config.json")
	sm.providersFile = filepath.Join(dataDir, "providers.json")
	sm.accountsFile = filepath.Join(dataDir, "accounts.json")
	sm.modelMappingsFile = filepath.Join(dataDir, "model_mappings.json")
	sm.sessionsFile = filepath.Join(dataDir, "sessions.json")
	sm.statsFile = filepath.Join(dataDir, "statistics.json")
	sm.requestLogsFile = filepath.Join(dataDir, "request_logs.json")

	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 加载所有数据
	sm.loadConfig()
	sm.loadProviders()
	sm.loadAccounts()
	sm.loadModelMappings()
	sm.loadSessions()
	sm.loadStatistics()
	sm.loadRequestLogs()

	// 初始化默认配置
	if sm.config == nil {
		sm.config = sm.getDefaultConfig()
	}

	// 初始化内置提供商
	sm.initializeBuiltinProviders()

	sm.logger.Info("Store manager initialized", logger.Field{Key: "dataDir", Value: dataDir})

	return sm, nil
}

// getDefaultConfig 获取默认配置
func (sm *StoreManager) getDefaultConfig() *types.AppConfig {
	return &types.AppConfig{
		Theme: types.ThemeLight,
		ProxyConfig: types.ProxyConfig{
			Enabled:  true,
			Port:    8080,
			Host:    "127.0.0.1",
			EnableApiKey: false,
			SessionConfig: types.SessionConfig{
				Enabled:          true,
				MaxHistoryLength: 100,
				DeleteAfterChat:  false,
			},
			ToolCallingConfig: types.ToolCallingConfig{
				Enabled:            true,
				DiagnosticsEnabled: false,
			},
			ContextConfig: types.ContextManagementConfig{
				Enabled:              true,
				MaxContextMessages:    40,
				EnableAutoSummary:     false,
			},
			LoadBalanceConfig: types.LoadBalanceConfig{
				Strategy:      types.LoadBalanceRoundRobin,
				ExcludeFailed: true,
				RecoveryTime:  60000,
			},
		},
		ManagementConfig: types.ManagementApiConfig{
			Enabled: true,
			Port:    8081,
			Host:    "127.0.0.1",
		},
		StartMinimized: false,
		AutoStart:      false,
	}
}

// ==================== 加密相关 ====================

// loadOrCreateEncryptionKey 加载或创建加密密钥
func (sm *StoreManager) loadOrCreateEncryptionKey() error {
	keyFile := filepath.Join(sm.dataDir, ".key")

	// 尝试加载现有密钥
	data, err := os.ReadFile(keyFile)
	if err == nil && len(data) == 32 {
		copy(sm.encryption, data)
		return nil
	}

	// 生成新密钥
	if _, err := rand.Read(sm.encryption); err != nil {
		return err
	}

	// 保存密钥
	return os.WriteFile(keyFile, sm.encryption, 0600)
}

// encrypt 加密数据
func (sm *StoreManager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(sm.encryption)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt 解密数据
func (sm *StoreManager) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(sm.encryption)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ==================== 配置管理 ====================

// GetConfig 获取配置
func (sm *StoreManager) GetConfig() types.AppConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.config == nil {
		return *sm.getDefaultConfig()
	}
	return *sm.config
}

// UpdateConfig 更新配置
func (sm *StoreManager) UpdateConfig(config types.AppConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.config = &config
	return sm.saveConfig()
}

func (sm *StoreManager) loadConfig() {
	data, err := os.ReadFile(sm.configFile)
	if err != nil {
		sm.config = sm.getDefaultConfig()
		return
	}

	if err := json.Unmarshal(data, sm.config); err != nil {
		sm.logger.Error("Failed to parse config", logger.Field{Key: "error", Value: err.Error()})
		sm.config = sm.getDefaultConfig()
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

// ==================== 提供商管理 ====================

// GetAllProviders 获取所有提供商
func (sm *StoreManager) GetAllProviders() []types.Provider {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	providers := make([]types.Provider, 0, len(sm.providers))
	for _, p := range sm.providers {
		providers = append(providers, *p)
	}
	return providers
}

// GetProviderByID 获取提供商
func (sm *StoreManager) GetProviderByID(id string) (*types.Provider, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	provider, ok := sm.providers[id]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// CreateProvider 创建提供商
func (sm *StoreManager) CreateProvider(provider types.Provider) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	provider.ID = generateID()
	provider.CreatedAt = time.Now().Unix()
	provider.UpdatedAt = provider.CreatedAt

	sm.providers[provider.ID] = &provider
	return sm.saveProviders()
}

// UpdateProvider 更新提供商
func (sm *StoreManager) UpdateProvider(id string, updates map[string]interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	provider, ok := sm.providers[id]
	if !ok {
		return ErrProviderNotFound
	}

	// 应用更新
	if name, ok := updates["name"].(string); ok {
		provider.Name = name
	}
	if apiEndpoint, ok := updates["apiEndpoint"].(string); ok {
		provider.APIEndpoint = apiEndpoint
	}
	if chatPath, ok := updates["chatPath"].(string); ok {
		provider.ChatPath = chatPath
	}
	if headers, ok := updates["headers"].(map[string]interface{}); ok {
		h := make(map[string]string)
		for k, v := range headers {
			if vs, ok := v.(string); ok {
				h[k] = vs
			}
		}
		provider.Headers = h
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		provider.Enabled = enabled
	}

	provider.UpdatedAt = time.Now().Unix()
	sm.providers[id] = provider

	return sm.saveProviders()
}

// DeleteProvider 删除提供商
func (sm *StoreManager) DeleteProvider(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.providers[id]; !ok {
		return ErrProviderNotFound
	}

	delete(sm.providers, id)

	// 删除关联的账户
	for aid, account := range sm.accounts {
		if account.ProviderID == id {
			delete(sm.accounts, aid)
		}
	}

	return sm.saveProviders()
}

func (sm *StoreManager) loadProviders() {
	data, err := os.ReadFile(sm.providersFile)
	if err != nil {
		return
	}

	var providers []types.Provider
	if err := json.Unmarshal(data, &providers); err != nil {
		sm.logger.Error("Failed to parse providers", logger.Field{Key: "error", Value: err.Error()})
		return
	}

	for i := range providers {
		sm.providers[providers[i].ID] = &providers[i]
	}
}

func (sm *StoreManager) saveProviders() error {
	providers := make([]types.Provider, 0, len(sm.providers))
	for _, p := range sm.providers {
		providers = append(providers, *p)
	}

	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal providers: %w", err)
	}

	if err := os.WriteFile(sm.providersFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write providers: %w", err)
	}

	// 保存账户
	return sm.saveAccounts()
}

// initializeBuiltinProviders 初始化内置提供商
func (sm *StoreManager) initializeBuiltinProviders() {
	builtinProviders := []types.Provider{
		{
			ID:           "deepseek",
			Name:         "DeepSeek",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeUserToken,
			APIEndpoint:  "https://chat.deepseek.com/api",
			ChatPath:     "/chat/completions",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "DeepSeek AI Assistant",
		},
		{
			ID:           "glm",
			Name:         "GLM",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeRefreshToken,
			APIEndpoint:  "https://open.bigmodel.cn/api/paas/v4",
			ChatPath:     "/chat/completions",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "Zhipu AI ChatGLM",
		},
		{
			ID:           "kimi",
			Name:         "Kimi",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeJWT,
			APIEndpoint:  "https://api.moonshot.cn/v1",
			ChatPath:     "/chat/completions",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "Moonshot AI Kimi",
		},
		{
			ID:           "qwen",
			Name:         "Qwen",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeTongyiSSOTicket,
			APIEndpoint:  "https://dashscope.aliyuncs.com/compatible-mode/v1",
			ChatPath:     "/chat/completions",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "Alibaba Qwen",
		},
		{
			ID:           "qwen-ai",
			Name:         "Qwen AI",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeToken,
			APIEndpoint:  "https://qwen.ai/api/chat",
			ChatPath:     "",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "Qwen AI Chat",
		},
		{
			ID:           "minimax",
			Name:         "MiniMax",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeRealUserID,
			APIEndpoint:  "https://api.minimax.chat/v",
			ChatPath:     "/text/chatcompletion_v2",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "MiniMax AI",
		},
		{
			ID:           "mimo",
			Name:         "Mimo",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeToken,
			APIEndpoint:  "https://api.mimo.ai/v1",
			ChatPath:     "/chat/completions",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "Mimo AI Assistant",
		},
		{
			ID:           "perplexity",
			Name:         "Perplexity",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeToken,
			APIEndpoint:  "https://api.perplexity.ai",
			ChatPath:     "/chat/completions",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "Perplexity AI Search",
		},
		{
			ID:           "zai",
			Name:         "Zai",
			Type:         types.ProviderTypeBuiltin,
			AuthType:     types.AuthTypeToken,
			APIEndpoint:  "https://api.zenvest.ai/v1",
			ChatPath:     "/chat/completions",
			Enabled:      true,
			Headers:      map[string]string{"Content-Type": "application/json"},
			Description:  "Zai AI",
		},
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, p := range builtinProviders {
		if _, exists := sm.providers[p.ID]; !exists {
			p.CreatedAt = time.Now().Unix()
			p.UpdatedAt = p.CreatedAt
			sm.providers[p.ID] = &p
		}
	}

	sm.saveProviders()
}

// ==================== 账户管理 ====================

// GetAllAccounts 获取所有账户
func (sm *StoreManager) GetAllAccounts() []types.Account {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	accounts := make([]types.Account, 0, len(sm.accounts))
	for _, a := range sm.accounts {
		accounts = append(accounts, *a)
	}
	return accounts
}

// GetAccountsByProviderID 根据提供商ID获取账户
func (sm *StoreManager) GetAccountsByProviderID(providerID string, activeOnly bool) []types.Account {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	accounts := make([]types.Account, 0)
	for _, a := range sm.accounts {
		if a.ProviderID == providerID {
			if !activeOnly || a.Status == types.AccountStatusActive {
				accounts = append(accounts, *a)
			}
		}
	}
	return accounts
}

// GetAccountByID 获取账户
func (sm *StoreManager) GetAccountByID(id string) (*types.Account, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	account, ok := sm.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

// CreateAccount 创建账户
func (sm *StoreManager) CreateAccount(providerID string, name string, credentials map[string]string) (*types.Account, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 验证提供商存在
	if _, ok := sm.providers[providerID]; !ok {
		return nil, ErrProviderNotFound
	}

	// 加密凭证
	encryptedCreds := make(map[string]string)
	for k, v := range credentials {
		enc, err := sm.encrypt(v)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt credential: %w", err)
		}
		encryptedCreds[k] = enc
	}

	account := &types.Account{
		ID:          generateID(),
		ProviderID:  providerID,
		Name:        name,
		Credentials: encryptedCreds,
		Status:      types.AccountStatusActive,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	sm.accounts[account.ID] = account

	if err := sm.saveAccounts(); err != nil {
		delete(sm.accounts, account.ID)
		return nil, err
	}

	return account, nil
}

// UpdateAccount 更新账户
func (sm *StoreManager) UpdateAccount(id string, updates map[string]interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	account, ok := sm.accounts[id]
	if !ok {
		return ErrAccountNotFound
	}

	// 更新名称
	if name, ok := updates["name"].(string); ok {
		account.Name = name
	}

	// 更新凭证
	if creds, ok := updates["credentials"].(map[string]interface{}); ok {
		for k, v := range creds {
			if vs, ok := v.(string); ok {
				enc, err := sm.encrypt(vs)
				if err != nil {
					return fmt.Errorf("failed to encrypt credential: %w", err)
				}
				account.Credentials[k] = enc
			}
		}
	}

	// 更新状态
	if status, ok := updates["status"].(string); ok {
		account.Status = types.AccountStatus(status)
	}

	account.UpdatedAt = time.Now().Unix()
	sm.accounts[id] = account

	return sm.saveAccounts()
}

// DeleteAccount 删除账户
func (sm *StoreManager) DeleteAccount(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.accounts[id]; !ok {
		return ErrAccountNotFound
	}

	delete(sm.accounts, id)
	return sm.saveAccounts()
}

// GetDecryptedCredentials 获取解密的凭证
func (sm *StoreManager) GetDecryptedCredentials(accountID string) (map[string]string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	account, ok := sm.accounts[accountID]
	if !ok {
		return nil, ErrAccountNotFound
	}

	decrypted := make(map[string]string)
	for k, v := range account.Credentials {
		plain, err := sm.decrypt(v)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt credential %s: %w", k, err)
		}
		decrypted[k] = plain
	}

	return decrypted, nil
}

// MarkAccountUsed 标记账户已使用
func (sm *StoreManager) MarkAccountUsed(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if account, ok := sm.accounts[id]; ok {
		account.LastUsed = time.Now().Unix()
		account.RequestCount++
		account.TodayUsed++
		account.UpdatedAt = time.Now().Unix()
	}
}

func (sm *StoreManager) loadAccounts() {
	data, err := os.ReadFile(sm.accountsFile)
	if err != nil {
		return
	}

	var accounts []types.Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		sm.logger.Error("Failed to parse accounts", logger.Field{Key: "error", Value: err.Error()})
		return
	}

	for i := range accounts {
		sm.accounts[accounts[i].ID] = &accounts[i]
	}
}

func (sm *StoreManager) saveAccounts() error {
	accounts := make([]types.Account, 0, len(sm.accounts))
	for _, a := range sm.accounts {
		accounts = append(accounts, *a)
	}

	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	if err := os.WriteFile(sm.accountsFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write accounts: %w", err)
	}

	return nil
}

// ==================== 模型映射管理 ====================

// GetAllModelMappings 获取所有模型映射
func (sm *StoreManager) GetAllModelMappings() []types.ModelMapping {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	mappings := make([]types.ModelMapping, 0, len(sm.modelMappings))
	for _, m := range sm.modelMappings {
		mappings = append(mappings, *m)
	}
	return mappings
}

// GetModelMapping 获取模型映射
func (sm *StoreManager) GetModelMapping(model string) (*types.ModelMappingEntry, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, m := range sm.modelMappings {
		for i := range m.Mappings {
			if m.Mappings[i].ClientModel == model && m.Mappings[i].Enabled {
				return &m.Mappings[i], nil
			}
		}
	}

	return nil, errors.New("model mapping not found")
}

// CreateModelMapping 创建模型映射
func (sm *StoreManager) CreateModelMapping(mapping types.ModelMapping) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	mapping.ID = generateID()
	mapping.CreatedAt = time.Now().Unix()
	mapping.UpdatedAt = mapping.CreatedAt

	sm.modelMappings[mapping.ID] = &mapping
	return sm.saveModelMappings()
}

// UpdateModelMapping 更新模型映射
func (sm *StoreManager) UpdateModelMapping(id string, mappings []types.ModelMappingEntry) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	mapping, ok := sm.modelMappings[id]
	if !ok {
		return errors.New("model mapping not found")
	}

	mapping.Mappings = mappings
	mapping.UpdatedAt = time.Now().Unix()
	sm.modelMappings[id] = mapping

	return sm.saveModelMappings()
}

func (sm *StoreManager) loadModelMappings() {
	data, err := os.ReadFile(sm.modelMappingsFile)
	if err != nil {
		return
	}

	var mappings []types.ModelMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		sm.logger.Error("Failed to parse model mappings", logger.Field{Key: "error", Value: err.Error()})
		return
	}

	for i := range mappings {
		sm.modelMappings[mappings[i].ID] = &mappings[i]
	}
}

func (sm *StoreManager) saveModelMappings() error {
	mappings := make([]types.ModelMapping, 0, len(sm.modelMappings))
	for _, m := range sm.modelMappings {
		mappings = append(mappings, *m)
	}

	data, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal model mappings: %w", err)
	}

	if err := os.WriteFile(sm.modelMappingsFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write model mappings: %w", err)
	}

	return nil
}

// ==================== 会话管理 ====================

// GetSession 获取会话
func (sm *StoreManager) GetSession(id string) (*types.SessionRecord, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[id]
	if !ok {
		return nil, errors.New("session not found")
	}
	return session, nil
}

// CreateSession 创建会话
func (sm *StoreManager) CreateSession(model string, messages []types.ChatMessage) *types.SessionRecord {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().Unix()
	session := &types.SessionRecord{
		ID:        generateID(),
		Model:     model,
		Messages:  messages,
		CreatedAt: now,
		UpdatedAt: now,
	}

	sm.sessions[session.ID] = session
	sm.saveSessions()

	return session
}

// UpdateSession 更新会话
func (sm *StoreManager) UpdateSession(id string, messages []types.ChatMessage) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[id]
	if !ok {
		return errors.New("session not found")
	}

	session.Messages = messages
	session.UpdatedAt = time.Now().Unix()
	sm.sessions[id] = session

	return sm.saveSessions()
}

// DeleteSession 删除会话
func (sm *StoreManager) DeleteSession(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, id)
	return sm.saveSessions()
}

func (sm *StoreManager) loadSessions() {
	data, err := os.ReadFile(sm.sessionsFile)
	if err != nil {
		return
	}

	var sessions []types.SessionRecord
	if err := json.Unmarshal(data, &sessions); err != nil {
		sm.logger.Error("Failed to parse sessions", logger.Field{Key: "error", Value: err.Error()})
		return
	}

	for i := range sessions {
		sm.sessions[sessions[i].ID] = &sessions[i]
	}
}

func (sm *StoreManager) saveSessions() error {
	sessions := make([]types.SessionRecord, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		sessions = append(sessions, *s)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	if err := os.WriteFile(sm.sessionsFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write sessions: %w", err)
	}

	return nil
}

// ==================== 统计管理 ====================

// RecordRequest 记录请求
func (sm *StoreManager) RecordRequest(success bool, latency int64, model, providerId, accountId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 初始化 maps
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

	// 更新模型使用
	if model != "" {
		curr, _ := strconv.ParseInt(sm.stats.ModelUsage[model], 10, 64)
		sm.stats.ModelUsage[model] = strconv.FormatInt(curr+1, 10)
	}

	// 更新提供商使用
	if providerId != "" {
		curr, _ := strconv.ParseInt(sm.stats.ProviderUsage[providerId], 10, 64)
		sm.stats.ProviderUsage[providerId] = strconv.FormatInt(curr+1, 10)
	}

	// 更新账户使用
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

	// 清理30天前的数据
	cutoff := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	for date := range sm.stats.DailyStats {
		if date < cutoff {
			delete(sm.stats.DailyStats, date)
		}
	}

	sm.saveStatistics()
}

// GetStatistics 获取统计信息
func (sm *StoreManager) GetStatistics() types.Statistics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.stats
}

func (sm *StoreManager) loadStatistics() {
	data, err := os.ReadFile(sm.statsFile)
	if err != nil {
		sm.stats = types.Statistics{
			TotalRequests:   0,
			SuccessRequests: 0,
			FailedRequests:  0,
			TotalLatency:   0,
			ModelUsage:     make(map[string]string),
			ProviderUsage:  make(map[string]string),
			AccountUsage:   make(map[string]string),
			DailyStats:     make(map[string]string),
			LastUpdated:    time.Now().Unix(),
		}
		return
	}

	if err := json.Unmarshal(data, &sm.stats); err != nil {
		sm.logger.Error("Failed to parse statistics", logger.Field{Key: "error", Value: err.Error()})
	}

	// 确保 maps 不为 nil
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

// ==================== 请求日志管理 ====================

// GetRequestLogs 获取请求日志
func (sm *StoreManager) GetRequestLogs(limit, offset int) []*types.RequestLogEntry {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	total := len(sm.requestLogs)
	if offset >= total {
		return []*types.RequestLogEntry{}
	}

	end := offset + limit
	if end > total {
		end = total
	}

	logs := make([]*types.RequestLogEntry, end-offset)
	copy(logs, sm.requestLogs[offset:end])

	return logs
}

// AddRequestLog 添加请求日志
func (sm *StoreManager) AddRequestLog(entry *types.RequestLogEntry) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	entry.ID = generateID()
	entry.Timestamp = time.Now().Unix()

	sm.requestLogs = append([]*types.RequestLogEntry{entry}, sm.requestLogs...)

	// 限制日志数量（最多保留1000条）
	const maxLogEntries = 1000

	if len(sm.requestLogs) > maxLogEntries {
		sm.requestLogs = sm.requestLogs[:maxLogEntries]
	}

	sm.saveRequestLogs()
}

func (sm *StoreManager) loadRequestLogs() {
	data, err := os.ReadFile(sm.requestLogsFile)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &sm.requestLogs); err != nil {
		sm.logger.Error("Failed to parse request logs", logger.Field{Key: "error", Value: err.Error()})
	}
}

func (sm *StoreManager) saveRequestLogs() error {
	data, err := json.MarshalIndent(sm.requestLogs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal request logs: %w", err)
	}

	if err := os.WriteFile(sm.requestLogsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write request logs: %w", err)
	}

	return nil
}

// ==================== 辅助函数 ====================

// generateID 生成唯一ID
func generateID() string {
	timestamp := time.Now().UnixNano()
	random := make([]byte, 8)
	rand.Read(random)
	return fmt.Sprintf("%d%s", timestamp, base64.RawURLEncoding.EncodeToString(random))
}

// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
	return generateID()
}

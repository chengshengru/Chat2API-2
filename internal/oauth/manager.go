package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"chat2api-wails/internal/logger"
	"chat2api-wails/internal/store"
	"chat2api-wails/internal/types"
)

// OAuthResult OAuth 流程结果
type OAuthResult struct {
	Success      bool              `json:"success"`
	ProviderID   string            `json:"providerId,omitempty"`
	ProviderType string            `json:"providerType,omitempty"`
	Credentials  map[string]string `json:"credentials,omitempty"`
	Account      *types.Account    `json:"account,omitempty"`
	AccountInfo  *AccountInfo      `json:"accountInfo,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// AccountInfo 账户信息
type AccountInfo struct {
	UserID  string `json:"userId,omitempty"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	Quota   int64  `json:"quota,omitempty"`
	Used    int64  `json:"used,omitempty"`
}

// TokenValidationResult 令牌验证结果
type TokenValidationResult struct {
	Valid     bool   `json:"valid"`
	TokenType string `json:"tokenType,omitempty"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
	AccountInfo *AccountInfo `json:"accountInfo,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ProviderAdapter 提供商适配器接口
type ProviderAdapter interface {
	GetProviderType() string
	GetAuthURL() (string, error)
	HandleCallback(callbackURL string) (*OAuthResult, error)
	ValidateToken(credentials map[string]string) (*TokenValidationResult, error)
	RefreshToken(credentials map[string]string) (*TokenValidationResult, error)
	LoginWithToken(token string) (*OAuthResult, error)
}

// Manager OAuth 管理器
type Manager struct {
	mu          sync.RWMutex
	adapters    map[string]ProviderAdapter
	logger      *logger.Logger
	storeManager *store.StoreManager
	pendingAuth map[string]string // providerId -> auth url
}

// NewManager 创建新的 OAuth 管理器
func NewManager(l *logger.Logger) *Manager {
	m := &Manager{
		logger:       l,
		adapters:     make(map[string]ProviderAdapter),
		pendingAuth: make(map[string]string),
	}

	// 注册内置提供商适配器
	m.registerAdapter(NewDeepSeekAdapter(l))
	m.registerAdapter(NewGLMAdapter(l))
	m.registerAdapter(NewKimiAdapter(l))
	m.registerAdapter(NewMiniMaxAdapter(l))
	m.registerAdapter(NewPerplexityAdapter(l))
	m.registerAdapter(NewQwenAdapter(l))

	l.Info("OAuth manager initialized")
	return m
}

func (m *Manager) registerAdapter(adapter ProviderAdapter) {
	m.adapters[adapter.GetProviderType()] = adapter
}

// GetAuthURL 获取认证 URL
func (m *Manager) GetAuthURL(providerType string) (string, error) {
	m.mu.RLock()
	adapter := m.adapters[providerType]
	m.mu.RUnlock()

	if adapter == nil {
		return "", fmt.Errorf("unsupported provider type: %s", providerType)
	}

	url, err := adapter.GetAuthURL()
	if err != nil {
		return "", err
	}

	m.pendingAuth[providerType] = url
	return url, nil
}

// StartLogin 开始认证流程
func (m *Manager) StartLogin(providerId string, providerType string) *OAuthResult {
	m.logger.Info("Starting OAuth login",
		logger.Field{Key: "providerId", Value: providerId},
		logger.Field{Key: "providerType", Value: providerType},
	)

	m.mu.RLock()
	adapter := m.adapters[providerType]
	m.mu.RUnlock()

	if adapter == nil {
		return &OAuthResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported provider type: %s", providerType),
		}
	}

	// 返回待处理的登录信息
	authURL, err := adapter.GetAuthURL()
	if err != nil {
		return &OAuthResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return &OAuthResult{
		Success:      false,
		ProviderID:   providerId,
		ProviderType: providerType,
		Error:        "Waiting for user authentication: " + authURL,
	}
}

// LoginWithToken 使用 Token 直接登录
func (m *Manager) LoginWithToken(providerType string, token string) *OAuthResult {
	m.logger.Info("Login with token", logger.Field{Key: "providerType", Value: providerType})

	m.mu.RLock()
	adapter := m.adapters[providerType]
	m.mu.RUnlock()

	if adapter == nil {
		return &OAuthResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported provider type: %s", providerType),
		}
	}

	result, err := adapter.LoginWithToken(token)
	if err != nil {
		return &OAuthResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return result
}

// ValidateToken 验证 Token 是否有效
func (m *Manager) ValidateToken(providerType string, credentials map[string]string) *TokenValidationResult {
	m.mu.RLock()
	adapter := m.adapters[providerType]
	m.mu.RUnlock()

	if adapter == nil {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Sprintf("unsupported provider type: %s", providerType),
		}
	}

	result, err := adapter.ValidateToken(credentials)
	if err != nil {
		return &TokenValidationResult{
			Valid: false,
			Error: err.Error(),
		}
	}

	return result
}

// RefreshToken 刷新 Token
func (m *Manager) RefreshToken(providerType string, credentials map[string]string) *TokenValidationResult {
	m.mu.RLock()
	adapter := m.adapters[providerType]
	m.mu.RUnlock()

	if adapter == nil {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Sprintf("unsupported provider type: %s", providerType),
		}
	}

	result, err := adapter.RefreshToken(credentials)
	if err != nil {
		return &TokenValidationResult{
			Valid: false,
			Error: err.Error(),
		}
	}

	return result
}

// HandleCallback 处理 OAuth 回调
func (m *Manager) HandleCallback(providerType string, callbackURL string) *OAuthResult {
	m.logger.Info("Handling OAuth callback", logger.Field{Key: "providerType", Value: providerType})

	m.mu.RLock()
	adapter := m.adapters[providerType]
	m.mu.RUnlock()

	if adapter == nil {
		return &OAuthResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported provider type: %s", providerType),
		}
	}

	result, err := adapter.HandleCallback(callbackURL)
	if err != nil {
		return &OAuthResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return result
}

// GetSupportedProviders 获取支持的提供商类型
func (m *Manager) GetSupportedProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, 0, len(m.adapters))
	for t := range m.adapters {
		result = append(result, t)
	}
	return result
}

// ==================== 内置适配器基类 ====================

type baseAdapter struct {
	name   string
	logger *logger.Logger
}

func (b *baseAdapter) GetProviderType() string {
	return b.name
}

// ==================== DeepSeek 适配器 ====================

type DeepSeekAdapter struct {
	baseAdapter
}

func NewDeepSeekAdapter(l *logger.Logger) *DeepSeekAdapter {
	return &DeepSeekAdapter{baseAdapter: baseAdapter{name: "deepseek", logger: l}}
}

func (a *DeepSeekAdapter) GetAuthURL() (string, error) {
	return "https://chat.deepseek.com/", nil
}

func (a *DeepSeekAdapter) HandleCallback(callbackURL string) (*OAuthResult, error) {
	// 尝试从 URL 中提取 token
	if callbackURL == "" {
		return &OAuthResult{
			Success: false,
			Error:   "callback URL is empty",
		}, nil
	}

	parsed, err := url.Parse(callbackURL)
	if err == nil {
		token := parsed.Query().Get("token")
		if token != "" {
			return &OAuthResult{
				Success: true,
				Credentials: map[string]string{
					"token": token,
				},
			}, nil
		}

		if parsed.Fragment != "" {
			fragmentMap := parseFragment(parsed.Fragment)
			if token, ok := fragmentMap["access_token"]; ok {
				return &OAuthResult{
					Success: true,
					Credentials: map[string]string{
						"token": token,
					},
				}, nil
			}
		}
	}

	return &OAuthResult{
		Success: false,
		Error:   "failed to extract token from callback URL",
	}, nil
}

func (a *DeepSeekAdapter) ValidateToken(credentials map[string]string) (*TokenValidationResult, error) {
	token := credentials["token"]
	if token == "" {
		return &TokenValidationResult{Valid: false, Error: "token is empty"}, nil
	}

	req, err := http.NewRequest("GET", "https://api.deepseek.com/user/balance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &TokenValidationResult{Valid: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	return &TokenValidationResult{
		Valid:     resp.StatusCode == http.StatusOK,
		TokenType: "Bearer",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

func (a *DeepSeekAdapter) RefreshToken(credentials map[string]string) (*TokenValidationResult, error) {
	// DeepSeek 通常使用长期有效的 API Key，不需要刷新
	return a.ValidateToken(credentials)
}

func (a *DeepSeekAdapter) LoginWithToken(token string) (*OAuthResult, error) {
	if token == "" {
		return &OAuthResult{Success: false, Error: "token is empty"}, nil
	}

	result := &OAuthResult{
		Success: true,
		Credentials: map[string]string{
			"token": token,
		},
	}

	validation, _ := a.ValidateToken(result.Credentials)
	if !validation.Valid {
		return &OAuthResult{
			Success: false,
			Error:   "invalid token",
		}, nil
	}

	return result, nil
}

// ==================== GLM 适配器 ====================

type GLMAdapter struct {
	baseAdapter
}

func NewGLMAdapter(l *logger.Logger) *GLMAdapter {
	return &GLMAdapter{baseAdapter: baseAdapter{name: "glm", logger: l}}
}

func (a *GLMAdapter) GetAuthURL() (string, error) {
	return "https://open.bigmodel.cn/", nil
}

func (a *GLMAdapter) HandleCallback(callbackURL string) (*OAuthResult, error) {
	return &OAuthResult{
		Success: false,
		Error:   "GLM OAuth requires manual API key input",
	}, nil
}

func (a *GLMAdapter) ValidateToken(credentials map[string]string) (*TokenValidationResult, error) {
	token := credentials["api_key"]
	if token == "" {
		token = credentials["token"]
	}
	if token == "" {
		return &TokenValidationResult{Valid: false, Error: "api key is empty"}, nil
	}

	req, err := http.NewRequest("GET", "https://open.bigmodel.cn/api/paas/v4/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &TokenValidationResult{Valid: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	return &TokenValidationResult{
		Valid:     resp.StatusCode == http.StatusOK,
		TokenType: "Bearer",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

func (a *GLMAdapter) RefreshToken(credentials map[string]string) (*TokenValidationResult, error) {
	return a.ValidateToken(credentials)
}

func (a *GLMAdapter) LoginWithToken(token string) (*OAuthResult, error) {
	if token == "" {
		return &OAuthResult{Success: false, Error: "token is empty"}, nil
	}

	result := &OAuthResult{
		Success: true,
		Credentials: map[string]string{
			"api_key": token,
		},
	}

	validation, _ := a.ValidateToken(result.Credentials)
	if !validation.Valid {
		return &OAuthResult{
			Success: false,
			Error:   "invalid api key",
		}, nil
	}

	return result, nil
}

// ==================== Kimi 适配器 ====================

type KimiAdapter struct {
	baseAdapter
}

func NewKimiAdapter(l *logger.Logger) *KimiAdapter {
	return &KimiAdapter{baseAdapter: baseAdapter{name: "kimi", logger: l}}
}

func (a *KimiAdapter) GetAuthURL() (string, error) {
	return "https://kimi.moonshot.cn/", nil
}

func (a *KimiAdapter) HandleCallback(callbackURL string) (*OAuthResult, error) {
	return &OAuthResult{
		Success: false,
		Error:   "Kimi OAuth requires manual API key input",
	}, nil
}

func (a *KimiAdapter) ValidateToken(credentials map[string]string) (*TokenValidationResult, error) {
	token := credentials["api_key"]
	if token == "" {
		token = credentials["token"]
	}
	if token == "" {
		return &TokenValidationResult{Valid: false, Error: "api key is empty"}, nil
	}

	req, err := http.NewRequest("GET", "https://api.moonshot.cn/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &TokenValidationResult{Valid: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	return &TokenValidationResult{
		Valid:     resp.StatusCode == http.StatusOK,
		TokenType: "Bearer",
	}, nil
}

func (a *KimiAdapter) RefreshToken(credentials map[string]string) (*TokenValidationResult, error) {
	return a.ValidateToken(credentials)
}

func (a *KimiAdapter) LoginWithToken(token string) (*OAuthResult, error) {
	if token == "" {
		return &OAuthResult{Success: false, Error: "token is empty"}, nil
	}

	result := &OAuthResult{
		Success: true,
		Credentials: map[string]string{
			"api_key": token,
		},
	}

	validation, _ := a.ValidateToken(result.Credentials)
	if !validation.Valid {
		return &OAuthResult{
			Success: false,
			Error:   "invalid api key",
		}, nil
	}

	return result, nil
}

// ==================== MiniMax 适配器 ====================

type MiniMaxAdapter struct {
	baseAdapter
}

func NewMiniMaxAdapter(l *logger.Logger) *MiniMaxAdapter {
	return &MiniMaxAdapter{baseAdapter: baseAdapter{name: "minimax", logger: l}}
}

func (a *MiniMaxAdapter) GetAuthURL() (string, error) {
	return "https://api.minimax.chat/", nil
}

func (a *MiniMaxAdapter) HandleCallback(callbackURL string) (*OAuthResult, error) {
	return &OAuthResult{
		Success: false,
		Error:   "MiniMax OAuth requires manual API key input",
	}, nil
}

func (a *MiniMaxAdapter) ValidateToken(credentials map[string]string) (*TokenValidationResult, error) {
	token := credentials["api_key"]
	if token == "" {
		token = credentials["token"]
	}
	if token == "" {
		return &TokenValidationResult{Valid: false, Error: "api key is empty"}, nil
	}

	req, err := http.NewRequest("GET", "https://api.minimax.chat/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &TokenValidationResult{Valid: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	return &TokenValidationResult{
		Valid:     resp.StatusCode == http.StatusOK,
		TokenType: "Bearer",
	}, nil
}

func (a *MiniMaxAdapter) RefreshToken(credentials map[string]string) (*TokenValidationResult, error) {
	return a.ValidateToken(credentials)
}

func (a *MiniMaxAdapter) LoginWithToken(token string) (*OAuthResult, error) {
	if token == "" {
		return &OAuthResult{Success: false, Error: "token is empty"}, nil
	}

	result := &OAuthResult{
		Success: true,
		Credentials: map[string]string{
			"api_key": token,
		},
	}

	validation, _ := a.ValidateToken(result.Credentials)
	if !validation.Valid {
		return &OAuthResult{
			Success: false,
			Error:   "invalid api key",
		}, nil
	}

	return result, nil
}

// ==================== Perplexity 适配器 ====================

type PerplexityAdapter struct {
	baseAdapter
}

func NewPerplexityAdapter(l *logger.Logger) *PerplexityAdapter {
	return &PerplexityAdapter{baseAdapter: baseAdapter{name: "perplexity", logger: l}}
}

func (a *PerplexityAdapter) GetAuthURL() (string, error) {
	return "https://labs.perplexity.ai/", nil
}

func (a *PerplexityAdapter) HandleCallback(callbackURL string) (*OAuthResult, error) {
	return &OAuthResult{
		Success: false,
		Error:   "Perplexity OAuth requires manual API key input",
	}, nil
}

func (a *PerplexityAdapter) ValidateToken(credentials map[string]string) (*TokenValidationResult, error) {
	token := credentials["api_key"]
	if token == "" {
		token = credentials["token"]
	}
	if token == "" {
		return &TokenValidationResult{Valid: false, Error: "api key is empty"}, nil
	}

	req, err := http.NewRequest("GET", "https://api.perplexity.ai/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &TokenValidationResult{Valid: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	return &TokenValidationResult{
		Valid:     resp.StatusCode == http.StatusOK,
		TokenType: "Bearer",
	}, nil
}

func (a *PerplexityAdapter) RefreshToken(credentials map[string]string) (*TokenValidationResult, error) {
	return a.ValidateToken(credentials)
}

func (a *PerplexityAdapter) LoginWithToken(token string) (*OAuthResult, error) {
	if token == "" {
		return &OAuthResult{Success: false, Error: "token is empty"}, nil
	}

	result := &OAuthResult{
		Success: true,
		Credentials: map[string]string{
			"api_key": token,
		},
	}

	validation, _ := a.ValidateToken(result.Credentials)
	if !validation.Valid {
		return &OAuthResult{
			Success: false,
			Error:   "invalid api key",
		}, nil
	}

	return result, nil
}

// ==================== Qwen 适配器 ====================

type QwenAdapter struct {
	baseAdapter
}

func NewQwenAdapter(l *logger.Logger) *QwenAdapter {
	return &QwenAdapter{baseAdapter: baseAdapter{name: "qwen", logger: l}}
}

func (a *QwenAdapter) GetAuthURL() (string, error) {
	return "https://dashscope.aliyun.com/", nil
}

func (a *QwenAdapter) HandleCallback(callbackURL string) (*OAuthResult, error) {
	return &OAuthResult{
		Success: false,
		Error:   "Qwen OAuth requires manual API key input",
	}, nil
}

func (a *QwenAdapter) ValidateToken(credentials map[string]string) (*TokenValidationResult, error) {
	token := credentials["api_key"]
	if token == "" {
		token = credentials["token"]
	}
	if token == "" {
		return &TokenValidationResult{Valid: false, Error: "api key is empty"}, nil
	}

	body := map[string]interface{}{
		"model": "qwen-turbo",
		"input": map[string]interface{}{
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "ping",
				},
			},
		},
	}

	bodyJSON, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation",
		bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &TokenValidationResult{Valid: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	return &TokenValidationResult{
		Valid:     resp.StatusCode == http.StatusOK,
		TokenType: "Bearer",
	}, nil
}

func (a *QwenAdapter) RefreshToken(credentials map[string]string) (*TokenValidationResult, error) {
	return a.ValidateToken(credentials)
}

func (a *QwenAdapter) LoginWithToken(token string) (*OAuthResult, error) {
	if token == "" {
		return &OAuthResult{Success: false, Error: "token is empty"}, nil
	}

	result := &OAuthResult{
		Success: true,
		Credentials: map[string]string{
			"api_key": token,
		},
	}

	validation, _ := a.ValidateToken(result.Credentials)
	if !validation.Valid {
		return &OAuthResult{
			Success: false,
			Error:   "invalid api key",
		}, nil
	}

	return result, nil
}

// ==================== 工具函数 ====================

// parseFragment 解析 URL 片段参数
func parseFragment(fragment string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(fragment, "&")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			if value, err := url.QueryUnescape(kv[1]); err == nil {
				result[kv[0]] = value
			} else {
				result[kv[0]] = kv[1]
			}
		}
	}
	return result
}

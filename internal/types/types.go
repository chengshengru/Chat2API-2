package types

// AccountStatus 账户状态类型
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusExpired  AccountStatus = "expired"
	AccountStatusError    AccountStatus = "error"
)

// ProviderStatus 提供商状态类型
type ProviderStatus string

const (
	ProviderStatusOnline  ProviderStatus = "online"
	ProviderStatusOffline ProviderStatus = "offline"
	ProviderStatusUnknown ProviderStatus = "unknown"
)

// AuthType 认证类型
type AuthType string

const (
	AuthTypeOAuth       AuthType = "oauth"
	AuthTypeToken       AuthType = "token"
	AuthTypeCookie      AuthType = "cookie"
	AuthTypeUserToken   AuthType = "userToken"
	AuthTypeRefresh     AuthType = "refresh_token"
	AuthTypeJWT         AuthType = "jwt"
	AuthTypeRealUserID  AuthType = "realUserID_token"
	AuthTypeTongyiSSO   AuthType = "tongyi_sso_ticket"
)

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	StrategyRoundRobin LoadBalanceStrategy = "round-robin"
	StrategyFillFirst  LoadBalanceStrategy = "fill-first"
	StrategyFailover   LoadBalanceStrategy = "failover"
)

// Theme 主题类型
type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

// Provider 提供商定义
type Provider struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Type             string            `json:"type"` // "builtin" | "custom"
	AuthType         AuthType          `json:"authType"`
	APIEndpoint      string            `json:"apiEndpoint"`
	ChatPath         string            `json:"chatPath,omitempty"`
	Headers          map[string]string `json:"headers"`
	Enabled          bool              `json:"enabled"`
	CreatedAt        int64             `json:"createdAt"`
	UpdatedAt        int64             `json:"updatedAt"`
	Description      string            `json:"description,omitempty"`
	Icon             string            `json:"icon,omitempty"`
	SupportedModels  []string          `json:"supportedModels,omitempty"`
	ModelMappings    map[string]string `json:"modelMappings,omitempty"`
	Status           ProviderStatus    `json:"status,omitempty"`
	LastStatusCheck  int64             `json:"lastStatusCheck,omitempty"`
}

// Account 账户定义
type Account struct {
	ID           string            `json:"id"`
	ProviderID   string            `json:"providerId"`
	Name         string            `json:"name"`
	Email        string            `json:"email,omitempty"`
	Credentials  map[string]string `json:"credentials,omitempty"`
	Status       AccountStatus     `json:"status"`
	LastUsed     int64             `json:"lastUsed,omitempty"`
	CreatedAt    int64             `json:"createdAt"`
	UpdatedAt    int64             `json:"updatedAt"`
	ErrorMessage string            `json:"errorMessage,omitempty"`
	RequestCount int64             `json:"requestCount,omitempty"`
	DailyLimit   int64             `json:"dailyLimit,omitempty"`
	TodayUsed    int64             `json:"todayUsed,omitempty"`
}

// ApiKey API密钥定义
type ApiKey struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   int64  `json:"createdAt"`
	LastUsedAt  int64  `json:"lastUsedAt,omitempty"`
	UsageCount  int64  `json:"usageCount"`
	Description string `json:"description,omitempty"`
}

// ModelMapping 模型映射定义
type ModelMapping struct {
	RequestModel       string `json:"requestModel"`
	ActualModel        string `json:"actualModel"`
	PreferredProviderID string `json:"preferredProviderId,omitempty"`
	PreferredAccountID  string `json:"preferredAccountId,omitempty"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	SessionTimeout        int64 `json:"sessionTimeout"`
	MaxMessagesPerSession int   `json:"maxMessagesPerSession"`
	DeleteAfterTimeout    bool  `json:"deleteAfterTimeout"`
	MaxSessionsPerAccount int   `json:"maxSessionsPerAccount"`
}

// RequestLogConfig 请求日志配置
type RequestLogConfig struct {
	Enabled            bool   `json:"enabled"`
	MaxEntries         int    `json:"maxEntries"`
	IncludeBodies      bool   `json:"includeBodies"`
	MaxBodyChars       int    `json:"maxBodyChars"`
	RedactSensitiveData bool  `json:"redactSensitiveData"`
}

// ManagementApiConfig 管理API配置
type ManagementApiConfig struct {
	EnableManagementApi   bool   `json:"enableManagementApi"`
	ManagementApiSecret   string `json:"managementApiSecret"`
	ManagementApiPort     int    `json:"managementApiPort,omitempty"`
}

// ToolCallingConfig 工具调用配置
type ToolCallingConfig struct {
	Enabled            bool   `json:"enabled"`
	PromptMode         string `json:"promptMode"` // "standard" | "custom"
	CustomTemplate     string `json:"customTemplate,omitempty"`
	MaxToolsPerRequest int    `json:"maxToolsPerRequest"`
	MaxRetries         int    `json:"maxRetries"`
	TimeoutSeconds     int    `json:"timeoutSeconds"`
}

// AppConfig 应用配置
type AppConfig struct {
	ProxyPort           int                  `json:"proxyPort"`
	ProxyHost           string               `json:"proxyHost"`
	LoadBalanceStrategy LoadBalanceStrategy    `json:"loadBalanceStrategy"`
	ModelMappings       map[string]ModelMapping `json:"modelMappings"`
	Theme               Theme                `json:"theme"`
	AutoStart           bool                 `json:"autoStart"`
	AutoStartProxy      bool                 `json:"autoStartProxy"`
	MinimizeToTray      bool                 `json:"minimizeToTray"`
	LogLevel            string               `json:"logLevel"` // "debug" | "info" | "warn" | "error"
	LogRetentionDays    int                  `json:"logRetentionDays"`
	RequestLogConfig    RequestLogConfig     `json:"requestLogConfig"`
	RequestTimeout      int                  `json:"requestTimeout"`
	RetryCount          int                  `json:"retryCount"`
	ApiKeys             []ApiKey             `json:"apiKeys"`
	EnableApiKey        bool                 `json:"enableApiKey"`
	OauthProxyMode      string               `json:"oauthProxyMode"` // "system" | "none"
	SessionConfig       SessionConfig        `json:"sessionConfig"`
	ToolCallingConfig   ToolCallingConfig    `json:"toolCallingConfig"`
	ManagementApi       ManagementApiConfig  `json:"managementApi"`
	Language            string               `json:"language"` // "zh-CN" | "en-US"
}

// DefaultAppConfig 返回默认配置
func DefaultAppConfig() AppConfig {
	return AppConfig{
		ProxyPort:           8080,
		ProxyHost:           "127.0.0.1",
		LoadBalanceStrategy: StrategyRoundRobin,
		ModelMappings:       make(map[string]ModelMapping),
		Theme:               ThemeSystem,
		AutoStart:           false,
		AutoStartProxy:      false,
		MinimizeToTray:      true,
		LogLevel:            "info",
		LogRetentionDays:    7,
		RequestLogConfig:    RequestLogConfig{
			Enabled:              true,
			MaxEntries:           1000,
			IncludeBodies:       false,
			MaxBodyChars:         500,
			RedactSensitiveData: true,
		},
		RequestTimeout:    120,
		RetryCount:        1,
		ApiKeys:           []ApiKey{},
		EnableApiKey:      false,
		OauthProxyMode:    "system",
		SessionConfig: SessionConfig{
			SessionTimeout:        3600,
			MaxMessagesPerSession: 50,
			DeleteAfterTimeout:    false,
			MaxSessionsPerAccount: 10,
		},
		ToolCallingConfig: ToolCallingConfig{
			Enabled:            true,
			PromptMode:         "standard",
			MaxToolsPerRequest: 10,
			MaxRetries:         3,
			TimeoutSeconds:     120,
		},
		ManagementApi: ManagementApiConfig{
			EnableManagementApi: false,
			ManagementApiSecret: "",
		},
		Language:            "zh-CN",
	}
}

// SessionRecord 会话记录
type SessionRecord struct {
	ID                string                 `json:"id"`
	ProviderID        string                 `json:"providerId"`
	AccountID         string                 `json:"accountId"`
	ProviderType      string                 `json:"providerType,omitempty"`
	SessionKey        string                 `json:"sessionKey,omitempty"`
	ProviderSessionID string                 `json:"providerSessionId"`
	ParentMessageID   string                 `json:"parentMessageId,omitempty"`
	SessionType       string                 `json:"sessionType"` // "chat" | "agent"
	Messages          []map[string]interface{} `json:"messages"`
	Credentials       map[string]string      `json:"credentials,omitempty"`
	CreatedAt         int64                  `json:"createdAt"`
	LastActiveAt      int64                  `json:"lastActiveAt"`
	Status            string                 `json:"status"` // "active" | "expired" | "deleted"
	Model             string                 `json:"model,omitempty"`
}

// OAuthResult OAuth 结果
type OAuthResult struct {
	Success      bool              `json:"success"`
	ProviderID   string            `json:"providerId,omitempty"`
	ProviderType string            `json:"providerType,omitempty"`
	Credentials  map[string]string `json:"credentials,omitempty"`
	Account      *Account          `json:"account,omitempty"`
	AccountInfo  *AccountInfo      `json:"accountInfo,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// AccountInfo 账户信息
type AccountInfo struct {
	UserID   string `json:"userId,omitempty"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Quota    int64  `json:"quota,omitempty"`
	Used     int64  `json:"used,omitempty"`
	ExpiresAt int64 `json:"expiresAt,omitempty"`
}

// ProviderCheckResult 提供商检查结果
type ProviderCheckResult struct {
	ProviderID string          `json:"providerId"`
	Status     ProviderStatus  `json:"status"`
	Latency    int64           `json:"latency,omitempty"`
	Error      string          `json:"error,omitempty"`
}

// ProxyStatus 代理服务器状态
type ProxyStatus struct {
	IsRunning    bool  `json:"isRunning"`
	Port         int   `json:"port"`
	Host         string `json:"host"`
	Uptime       int64 `json:"uptime"`
	Connections  int64 `json:"connections"`
	StartedAt    int64 `json:"startedAt"`
}

// ProxyStatistics 代理服务器统计信息
type ProxyStatistics struct {
	TotalRequests     int64            `json:"totalRequests"`
	SuccessRequests   int64            `json:"successRequests"`
	FailedRequests    int64            `json:"failedRequests"`
	AvgLatency        int64            `json:"avgLatency"`
	RequestsPerMinute int64            `json:"requestsPerMinute"`
	ActiveConnections int64            `json:"activeConnections"`
	ModelUsage        map[string]int64 `json:"modelUsage"`
	ProviderUsage     map[string]int64 `json:"providerUsage"`
	AccountUsage      map[string]int64 `json:"accountUsage"`
	LastUpdated       int64            `json:"lastUpdated"`
}

// EffectiveModel 有效模型信息
type EffectiveModel struct {
	DisplayName   string `json:"displayName"`
	ActualModelID string `json:"actualModelId"`
	IsCustom      bool   `json:"isCustom"`
}

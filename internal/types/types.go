package types

import "time"

// ==================== 基础枚举类型 ====================

// AccountStatus 账户状态
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusExpired  AccountStatus = "expired"
	AccountStatusError    AccountStatus = "error"
)

// ProviderType 提供商类型
type ProviderType string

const (
	ProviderTypeBuiltin ProviderType = "builtin"
	ProviderTypeCustom  ProviderType = "custom"
)

// AuthType 认证类型
type AuthType string

const (
	AuthTypeOAuth          AuthType = "oauth"
	AuthTypeToken          AuthType = "token"
	AuthTypeCookie         AuthType = "cookie"
	AuthTypeUserToken      AuthType = "userToken"
	AuthTypeRefreshToken   AuthType = "refresh_token"
	AuthTypeJWT            AuthType = "jwt"
	AuthTypeRealUserID     AuthType = "realUserID_token"
	AuthTypeTongyiSSOTicket AuthType = "tongyi_sso_ticket"
)

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	LoadBalanceRoundRobin LoadBalanceStrategy = "round-robin"
	LoadBalanceFillFirst  LoadBalanceStrategy = "fill-first"
	LoadBalanceFailover   LoadBalanceStrategy = "failover"
)

// Theme 主题
type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

// ==================== 提供商相关类型 ====================

// CredentialField 凭证字段配置
type CredentialField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"` // text, password, textarea
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
	HelpText    string `json:"helpText,omitempty"`
}

// Provider 提供商配置
type Provider struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Type            ProviderType      `json:"type"`
	AuthType        AuthType          `json:"authType"`
	APIEndpoint     string            `json:"apiEndpoint"`
	ChatPath        string            `json:"chatPath,omitempty"`
	Headers         map[string]string `json:"headers"`
	Enabled         bool              `json:"enabled"`
	CreatedAt       int64             `json:"createdAt"`
	UpdatedAt       int64             `json:"updatedAt"`
	Description     string            `json:"description,omitempty"`
	Icon            string            `json:"icon,omitempty"`
	SupportedModels []string         `json:"supportedModels,omitempty"`
}

// BuiltinProviderConfig 内置提供商配置（包含凭证字段配置）
type BuiltinProviderConfig struct {
	Provider
	CredentialFields    []CredentialField   `json:"credentialFields"`
	TokenCheckEndpoint   string              `json:"tokenCheckEndpoint,omitempty"`
	TokenCheckMethod     string              `json:"tokenCheckMethod,omitempty"`
	ModelsApiEndpoint    string              `json:"modelsApiEndpoint,omitempty"`
	ModelsApiHeaders     map[string]string   `json:"modelsApiHeaders,omitempty"`
}

// ==================== 账户相关类型 ====================

// Account 账户配置
type Account struct {
	ID           string            `json:"id"`
	ProviderID   string            `json:"providerId"`
	Name         string            `json:"name"`
	Email        string            `json:"email,omitempty"`
	Credentials  map[string]string `json:"credentials"` // 加密存储
	Status       AccountStatus     `json:"status"`
	LastUsed     int64             `json:"lastUsed,omitempty"`
	CreatedAt    int64             `json:"createdAt"`
	UpdatedAt    int64             `json:"updatedAt"`
	ErrorMessage string            `json:"errorMessage,omitempty"`
	RequestCount int64             `json:"requestCount,omitempty"`
	DailyLimit   int64             `json:"dailyLimit,omitempty"`
	TodayUsed    int64             `json:"todayUsed,omitempty"`
}

// ==================== API Key 相关类型 ====================

// ApiKey API Key配置
type ApiKey struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

// ==================== 模型映射相关类型 ====================

// ModelMappingEntry 模型映射条目
type ModelMappingEntry struct {
	ClientModel  string `json:"clientModel"`  // 客户端请求的模型名
	ProviderID   string `json:"providerId"`   // 映射到的提供商ID
	AccountID    string `json:"accountId"`    // 映射到的账户ID
	ProviderModel string `json:"providerModel"` // 提供商实际的模型名
	Enabled      bool   `json:"enabled"`
}

// ModelMapping 模型映射配置
type ModelMapping struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Mappings    []ModelMappingEntry `json:"mappings"`
	CreatedAt   int64              `json:"createdAt"`
	UpdatedAt   int64              `json:"updatedAt"`
}

// ==================== 代理配置相关类型 ====================

// SessionConfig 会话配置
type SessionConfig struct {
	Enabled          bool   `json:"enabled"`
	MaxHistoryLength int    `json:"maxHistoryLength"`
	DeleteAfterChat  bool   `json:"deleteAfterChat"`
}

// ToolCallingConfig Tool Calling 配置
type ToolCallingConfig struct {
	Enabled            bool     `json:"enabled"`
	ForceAdapter       string   `json:"forceAdapter,omitempty"` // 强制使用的适配器
	AllowedProtocols    []string `json:"allowedProtocols,omitempty"`
	DisabledProtocols   []string `json:"disabledProtocols,omitempty"`
	CustomPromptTemplate string  `json:"customPromptTemplate,omitempty"`
	DiagnosticsEnabled  bool     `json:"diagnosticsEnabled"`
}

// ContextManagementConfig 上下文管理配置
type ContextManagementConfig struct {
	Enabled              bool    `json:"enabled"`
	MaxContextMessages    int     `json:"maxContextMessages"`
	SummaryModel          string  `json:"summaryModel,omitempty"`
	SummaryThreshold      int     `json:"summaryThreshold"`
	EnableAutoSummary     bool    `json:"enableAutoSummary"`
}

// LoadBalanceConfig 负载均衡配置
type LoadBalanceConfig struct {
	Strategy        LoadBalanceStrategy `json:"strategy"`
	ExcludeFailed   bool                `json:"excludeFailed"`
	RecoveryTime    int                 `json:"recoveryTime"` // 毫秒
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled            bool                    `json:"enabled"`
	Port               int                     `json:"port"`
	Host               string                  `json:"host"`
	EnableApiKey       bool                    `json:"enableApiKey"`
	ApiKeys            []ApiKey                `json:"apiKeys,omitempty"`
	SessionConfig      SessionConfig            `json:"sessionConfig"`
	ToolCallingConfig  ToolCallingConfig        `json:"toolCallingConfig"`
	ContextConfig      ContextManagementConfig  `json:"contextConfig"`
	LoadBalanceConfig  LoadBalanceConfig        `json:"loadBalanceConfig"`
}

// ==================== 应用配置相关类型 ====================

// ManagementApiConfig 管理API配置
type ManagementApiConfig struct {
	Enabled   bool   `json:"enabled"`
	Port      int    `json:"port"`
	Host      string `json:"host"`
	AuthToken string `json:"authToken,omitempty"`
}

// AppConfig 应用配置
type AppConfig struct {
	ProxyConfig     ProxyConfig      `json:"proxyConfig"`
	ManagementConfig ManagementApiConfig `json:"managementConfig"`
	Theme           Theme            `json:"theme"`
	StartMinimized  bool             `json:"startMinimized"`
	AutoStart       bool             `json:"autoStart"`
}

// ==================== 会话相关类型 ====================

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string        `json:"role"` // system, user, assistant, tool
	Content string        `json:"content"`
	Name    string        `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数
type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Type      string          `json:"type"`
	Function  ToolFuncDef     `json:"function"`
}

// ToolFuncDef 工具函数定义
type ToolFuncDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// SessionRecord 会话记录
type SessionRecord struct {
	ID        string         `json:"id"`
	Model     string         `json:"model"`
	Messages  []ChatMessage  `json:"messages"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
}

// ==================== 代理请求/响应类型 ====================

// ChatCompletionRequest 聊天补全请求（OpenAI兼容）
type ChatCompletionRequest struct {
	Model           string           `json:"model"`
	OriginalModel   string           `json:"originalModel,omitempty"`
	Messages        []ChatMessage    `json:"messages"`
	Temperature     *float64         `json:"temperature,omitempty"`
	TopP            *float64         `json:"top_p,omitempty"`
	N               *int             `json:"n,omitempty"`
	Stream          bool             `json:"stream,omitempty"`
	Stop            interface{}      `json:"stop,omitempty"`
	MaxTokens       *int             `json:"max_tokens,omitempty"`
	PresencePenalty *float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64       `json:"frequency_penalty,omitempty"`
	LogitBias       map[string]int   `json:"logit_bias,omitempty"`
	User            string           `json:"user,omitempty"`
	WebSearch       bool             `json:"web_search,omitempty"`
	ReasoningEffort string           `json:"reasoning_effort,omitempty"`
	DeepResearch    bool             `json:"deep_research,omitempty"`
	Tools           []ToolDefinition  `json:"tools,omitempty"`
	ToolChoice      interface{}      `json:"tool_choice,omitempty"`
	ToolFormat      string           `json:"tool_format,omitempty"`
}

// ChatCompletionResponse 聊天补全响应
type ChatCompletionResponse struct {
	ID      string              `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *UsageInfo           `json:"usage,omitempty"`
}

// ChatCompletionChoice 聊天补全选项
type ChatCompletionChoice struct {
	Index        int              `json:"index"`
	Message      *ChatMessage     `json:"message,omitempty"`
	FinishReason string           `json:"finish_reason,omitempty"`
}

// UsageInfo 使用量信息
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ==================== 流式响应类型 ====================

// SSEEvent SSE事件
type SSEEvent struct {
	Event string `json:"event,omitempty"`
	Data  string `json:"data"`
	ID    string `json:"id,omitempty"`
	Retry int    `json:"retry,omitempty"`
}

// StreamChunk 流式数据块
type StreamChunk struct {
	ID               string            `json:"id"`
	Object           string            `json:"object"`
	Created          int64             `json:"created"`
	Model            string            `json:"model"`
	Choices          []StreamChoice    `json:"choices"`
	XDelta           string            `json:"x_delta,omitempty"`
	FinishReason     string            `json:"finish_reason,omitempty"`
}

// StreamChoice 流式选项
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        interface{} `json:"delta,omitempty"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

// ==================== 代理上下文 ====================

// ProxyContext 代理请求上下文
type ProxyContext struct {
	RequestID   string `json:"requestId"`
	ProviderID  string `json:"providerId"`
	AccountID   string `json:"accountId"`
	Model       string `json:"model"`
	ActualModel string `json:"actualModel"`
	StartTime   int64  `json:"startTime"`
	IsStream    bool   `json:"isStream"`
	ClientIP    string `json:"clientIP"`
}

// ==================== 请求日志相关类型 ====================

// RequestLogConfig 请求日志配置
type RequestLogConfig struct {
	Enabled    bool     `json:"enabled"`
	Redact     []string `json:"redact,omitempty"`
	MaxEntries int      `json:"maxEntries"`
}

// RequestLogEntry 请求日志条目
type RequestLogEntry struct {
	ID           string                 `json:"id"`
	Timestamp    int64                  `json:"timestamp"`
	ProviderID   string                 `json:"providerId"`
	AccountID    string                 `json:"accountId"`
	Model        string                 `json:"model"`
	RequestBody  map[string]interface{}  `json:"requestBody"`
	ResponseBody map[string]interface{}  `json:"responseBody,omitempty"`
	StatusCode   int                    `json:"statusCode"`
	Latency      int64                  `json:"latency"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	Redacted     bool                   `json:"redacted"`
}

// ==================== OAuth 相关类型 ====================

// OAuthResult OAuth结果
type OAuthResult struct {
	Success      bool   `json:"success"`
	ProviderID   string `json:"providerId"`
	ProviderType string `json:"providerType"`
	Error        string `json:"error,omitempty"`
	Credentials  map[string]string `json:"credentials,omitempty"`
}

// OAuthProgressEvent OAuth进度事件
type OAuthProgressEvent struct {
	Step    string `json:"step"`
	Message string `json:"message"`
	Percent int    `json:"percent,omitempty"`
}

// TokenValidationResult Token验证结果
type TokenValidationResult struct {
	Valid         bool                   `json:"valid"`
	ExpiresAt    *time.Time             `json:"expiresAt,omitempty"`
	Scopes       []string               `json:"scopes,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// AccountInfo 账户信息
type AccountInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Provider string `json:"provider"`
}

// ==================== 代理状态和统计 ====================

// ProxyStatus 代理服务器状态
type ProxyStatus struct {
	IsRunning  bool   `json:"isRunning"`
	Port      int    `json:"port"`
	Host      string `json:"host"`
	StartedAt int64  `json:"startedAt"`
	Uptime    int64  `json:"uptime"`
	Addr      string `json:"addr"`
}

// Statistics 统计信息
type Statistics struct {
	TotalRequests   int64                `json:"totalRequests"`
	SuccessRequests int64                `json:"successRequests"`
	FailedRequests  int64                `json:"failedRequests"`
	TotalLatency    int64                `json:"totalLatency"`
	ModelUsage      map[string]string    `json:"modelUsage"`
	ProviderUsage   map[string]string    `json:"providerUsage"`
	AccountUsage    map[string]string    `json:"accountUsage"`
	DailyStats      map[string]string    `json:"dailyStats"`
	LastUpdated     int64                `json:"lastUpdated"`
}

// ==================== 管理API相关类型 ====================

// ManagementApiResponse 管理API响应
type ManagementApiResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data,omitempty"`
	Error   *ApiError      `json:"error,omitempty"`
}

// ApiError API错误
type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateProviderRequest 创建提供商请求
type CreateProviderRequest struct {
	Name        string            `json:"name"`
	Type        ProviderType      `json:"type"`
	AuthType    AuthType          `json:"authType"`
	APIEndpoint string            `json:"apiEndpoint"`
	ChatPath    string            `json:"chatPath,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// UpdateProviderRequest 更新提供商请求
type UpdateProviderRequest struct {
	Name        string            `json:"name,omitempty"`
	APIEndpoint string            `json:"apiEndpoint,omitempty"`
	ChatPath    string            `json:"chatPath,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
}

// CreateAccountRequest 创建账户请求
type CreateAccountRequest struct {
	ProviderID  string            `json:"providerId"`
	Name        string            `json:"name"`
	Credentials map[string]string `json:"credentials"`
}

// UpdateAccountRequest 更新账户请求
type UpdateAccountRequest struct {
	Name        string            `json:"name,omitempty"`
	Credentials map[string]string `json:"credentials,omitempty"`
	Status      *AccountStatus    `json:"status,omitempty"`
}

package logger

// LogLevel 定义日志级别
type LogLevel string

const (
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelDebug LogLevel = "debug"
)

// Field 用于结构化日志字段
type Field struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
}

// LogEntry 是单条日志记录
type LogEntry struct {
	ID         string                 `json:"id"`
	Timestamp  int64                  `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	AccountID  string                 `json:"accountId,omitempty"`
	ProviderID string                 `json:"providerId,omitempty"`
	RequestID  string                 `json:"requestId,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// LogStats 日志统计信息
type LogStats struct {
	Total int `json:"total"`
	Info  int `json:"info"`
	Warn  int `json:"warn"`
	Error int `json:"error"`
	Debug int `json:"debug"`
}

// LogFilter 用于过滤日志查询
type LogFilter struct {
	Level     LogLevel `json:"level,omitempty"`
	Keyword   string   `json:"keyword,omitempty"`
	StartTime int64    `json:"startTime,omitempty"`
	EndTime   int64    `json:"endTime,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}

// LogTrend 每日趋势
type LogTrend struct {
	Date  string `json:"date"`
	Total int    `json:"total"`
	Info  int    `json:"info"`
	Warn  int    `json:"warn"`
	Error int    `json:"error"`
}

// RequestLogEntry 是请求日志记录
type RequestLogEntry struct {
	ID           string                 `json:"id"`
	Timestamp    int64                  `json:"timestamp"`
	Status       string                 `json:"status"`
	StatusCode   int                    `json:"statusCode"`
	Method       string                 `json:"method"`
	URL          string                 `json:"url"`
	Model        string                 `json:"model"`
	ProviderID   string                 `json:"providerId,omitempty"`
	ProviderName string                 `json:"providerName,omitempty"`
	AccountID    string                 `json:"accountId,omitempty"`
	AccountName  string                 `json:"accountName,omitempty"`
	RequestBody  string                 `json:"requestBody,omitempty"`
	UserInput    string                 `json:"userInput,omitempty"`
	ResponsePreview string             `json:"responsePreview,omitempty"`
	ResponseBody string                 `json:"responseBody,omitempty"`
	Latency      int64                  `json:"latency"`
	IsStream     bool                   `json:"isStream"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	WebSearch    bool                   `json:"webSearch,omitempty"`
}

// PersistentStatistics 持久化的统计数据
type PersistentStatistics struct {
	TotalRequests    int64                `json:"totalRequests"`
	SuccessRequests  int64                `json:"successRequests"`
	FailedRequests   int64                `json:"failedRequests"`
	TotalLatency     int64                `json:"totalLatency"`
	LastUpdated      int64                `json:"lastUpdated"`
	ModelUsage       map[string]int       `json:"modelUsage"`
	ProviderUsage    map[string]int       `json:"providerUsage"`
	AccountUsage     map[string]int       `json:"accountUsage"`
	DailyStats       map[string]*DailyStats `json:"dailyStats"`
}

// DailyStats 每日统计数据
type DailyStats struct {
	Date            string         `json:"date"`
	TotalRequests   int64          `json:"totalRequests"`
	SuccessRequests int64          `json:"successRequests"`
	FailedRequests  int64          `json:"failedRequests"`
	TotalLatency    int64          `json:"totalLatency"`
	ModelUsage      map[string]int `json:"modelUsage"`
	ProviderUsage   map[string]int `json:"providerUsage"`
}

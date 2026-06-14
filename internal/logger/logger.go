package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Logger 提供结构化日志记录器
type Logger struct {
	mu             sync.RWMutex
	logs           []LogEntry
	logFile       string
	requestLogs   []RequestLogEntry
	requestFile    string
	stats         PersistentStatistics
	statsFile      string
	maxLogs       int
	retentionDays int
	initialized    bool
	baseDir       string
}

// New 创建一个新的 Logger 实例
func New() *Logger {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".chat2api", "logs")
	os.MkdirAll(baseDir, 0755)

	return &Logger{
		baseDir:       baseDir,
		logFile:       filepath.Join(baseDir, "app.log"),
		requestFile:  filepath.Join(baseDir, "request.log"),
		statsFile:     filepath.Join(baseDir, "stats.json"),
		maxLogs:       10000,
		retentionDays: 7,
		stats: PersistentStatistics{
			LastUpdated:   time.Now().Unix(),
			ModelUsage:    make(map[string]string),
			ProviderUsage: make(map[string]string),
			AccountUsage:  make(map[string]string),
			DailyStats:   make(map[string]string),
		},
	}
}

// NewWithDir 使用自定义目录创建 Logger
func NewWithDir(baseDir string) *Logger {
	os.MkdirAll(baseDir, 0755)

	return &Logger{
		baseDir:       baseDir,
		logFile:       filepath.Join(baseDir, "app.log"),
		requestFile:  filepath.Join(baseDir, "request.log"),
		statsFile:     filepath.Join(baseDir, "stats.json"),
		maxLogs:       10000,
		retentionDays: 7,
		stats: PersistentStatistics{
			LastUpdated:   time.Now().Unix(),
			ModelUsage:    make(map[string]string),
			ProviderUsage: make(map[string]string),
			AccountUsage:  make(map[string]string),
			DailyStats:    make(map[string]string),
		},
	}
}

// ==================== 核心日志方法 ====================

func (l *Logger) log(level LogLevel, message string, fields ...Field) LogEntry {
	l.mu.Lock()

	data := make(map[string]string)
	for _, f := range fields {
		data[f.Key] = f.Value
	}

	entry := LogEntry{
		ID:        fmt.Sprintf("%d-%d", time.Now().UnixNano(), len(l.logs)),
		Timestamp: time.Now().Unix(),
		Level:     level,
		Message:   message,
		Data:      data,
	}

	l.logs = append(l.logs, entry)

	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[len(l.logs)-l.maxLogs:]
	}

	l.mu.Unlock()

	go l.persistLog(entry)

	// 控制台输出
	l.consolePrint(entry)

	return entry
}

// Info 记录信息级日志
func (l *Logger) Info(message string, fields ...Field) LogEntry {
	return l.log(LevelInfo, message, fields...)
}

// Warn 记录警告级日志
func (l *Logger) Warn(message string, fields ...Field) LogEntry {
	return l.log(LevelWarn, message, fields...)
}

// Error 记录错误级日志
func (l *Logger) Error(message string, fields ...Field) LogEntry {
	return l.log(LevelError, message, fields...)
}

// Debug 记录调试级日志
func (l *Logger) Debug(message string, fields ...Field) LogEntry {
	return l.log(LevelDebug, message, fields...)
}

// ==================== 日志持久化 ====================

func (l *Logger) persistLog(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.Write(data)
	f.WriteString("\n")
}

// LoadLogs 从磁盘加载日志
func (l *Logger) LoadLogs() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(l.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		l.logs = append(l.logs, entry)
	}

	// 限制日志数量
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[len(l.logs)-l.maxLogs:]
	}

	l.initialized = true
	return nil
}

// ==================== 日志查询 ====================

// GetLogs 获取过滤后的日志
func (l *Logger) GetLogs(filter LogFilter) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]LogEntry, len(l.logs))
	copy(filtered, l.logs)

	if filter.Level != "" {
		levelFiltered := make([]LogEntry, 0)
		for _, log := range filtered {
			if log.Level == filter.Level {
				levelFiltered = append(levelFiltered, log)
			}
		}
		filtered = levelFiltered
	}

	if filter.Keyword != "" {
		keyword := strings.ToLower(filter.Keyword)
		keywordFiltered := make([]LogEntry, 0)
		for _, log := range filtered {
			if strings.Contains(strings.ToLower(log.Message), keyword) {
				keywordFiltered = append(keywordFiltered, log)
			}
		}
		filtered = keywordFiltered
	}

	if filter.StartTime > 0 {
		timeFiltered := make([]LogEntry, 0)
		for _, log := range filtered {
			if log.Timestamp >= filter.StartTime {
				timeFiltered = append(timeFiltered, log)
			}
		}
		filtered = timeFiltered
	}

	if filter.EndTime > 0 {
		timeFiltered := make([]LogEntry, 0)
		for _, log := range filtered {
			if log.Timestamp <= filter.EndTime {
				timeFiltered = append(timeFiltered, log)
			}
		}
		filtered = timeFiltered
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp > filtered[j].Timestamp
	})

	if filter.Offset > 0 {
		if filter.Offset < len(filtered) {
			filtered = filtered[filter.Offset:]
		} else {
			filtered = nil
		}
	}

	if filter.Limit > 0 && filter.Limit < len(filtered) {
		filtered = filtered[:filter.Limit]
	}

	return filtered
}

// GetStats 获取日志统计
func (l *Logger) GetStats() LogStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := LogStats{Total: len(l.logs)}
	for _, log := range l.logs {
		switch log.Level {
		case LevelInfo:
			stats.Info++
		case LevelWarn:
			stats.Warn++
		case LevelError:
			stats.Error++
		case LevelDebug:
			stats.Debug++
		}
	}
	return stats
}

// GetTrend 获取日志趋势（按天）
func (l *Logger) GetTrend(days int) []LogTrend {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if days <= 0 {
		days = 7
	}

	now := time.Now()
	trends := make([]LogTrend, days)

	for i := 0; i < days; i++ {
		dayStart := now.AddDate(0, 0, -days+i+1).Unix() * 1000
		dayEnd := now.AddDate(0, 0, -days+i).Unix() * 1000
		date := time.Unix(dayStart/1000, 0).Format("2006-01-02")

		var total, info, warn, errCount int
		for _, log := range l.logs {
			if log.Timestamp >= dayStart && log.Timestamp < dayEnd {
				total++
				switch log.Level {
				case LevelInfo:
					info++
				case LevelWarn:
					warn++
				case LevelError:
					errCount++
				}
			}
		}

		trends[i] = LogTrend{
			Date:  date,
			Total: total,
			Info:  info,
			Warn:  warn,
			Error: errCount,
		}
	}

	return trends
}

// ==================== 请求日志 ====================

// AddRequestLog 添加请求日志
func (l *Logger) AddRequestLog(entry RequestLogEntry) {
	l.mu.Lock()
	entry.ID = fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), len(l.requestLogs))
	entry.Timestamp = time.Now().Unix()
	l.requestLogs = append(l.requestLogs, entry)
	l.mu.Unlock()

	// 更新统计
	l.updateStatistics(entry)

	// 持久化
	go l.persistRequestLog(entry)
}

func (l *Logger) persistRequestLog(entry RequestLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(l.requestFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.Write(data)
	f.WriteString("\n")
}

// GetRequestLogs 获取请求日志
func (l *Logger) GetRequestLogs(status string, providerID string, limit int) []RequestLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]RequestLogEntry, 0)
	for _, log := range l.requestLogs {
		if status != "" && log.Status != status {
			continue
		}
		if providerID != "" && log.ProviderID != providerID {
			continue
		}
		filtered = append(filtered, log)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp > filtered[j].Timestamp
	})

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}

	return filtered
}

// ==================== 统计数据 ====================

func (l *Logger) updateStatistics(entry RequestLogEntry) {
	l.stats.TotalRequests++
	if entry.Status == "success" {
		l.stats.SuccessRequests++
	} else {
		l.stats.FailedRequests++
	}
	l.stats.TotalLatency += entry.Latency

	if entry.Model != "" {
		curr := l.stats.ModelUsage[entry.Model]
		n, _ := parseInt64(curr)
		n++
		l.stats.ModelUsage[entry.Model] = fmt.Sprintf("%d", n)
	}
	if entry.ProviderID != "" {
		curr := l.stats.ProviderUsage[entry.ProviderID]
		n, _ := parseInt64(curr)
		n++
		l.stats.ProviderUsage[entry.ProviderID] = fmt.Sprintf("%d", n)
	}
	if entry.AccountID != "" {
		curr := l.stats.AccountUsage[entry.AccountID]
		n, _ := parseInt64(curr)
		n++
		l.stats.AccountUsage[entry.AccountID] = fmt.Sprintf("%d", n)
	}

	// 更新每日统计（简化：存储 JSON 字符串）
	date := time.Unix(entry.Timestamp, 0).Format("2006-01-02")
	jsonStr, _ := json.Marshal(map[string]int64{
		"total":   l.stats.TotalRequests,
		"success": l.stats.SuccessRequests,
		"failed":  l.stats.FailedRequests,
	})
	l.stats.DailyStats[date] = string(jsonStr)

	l.stats.LastUpdated = time.Now().Unix()

	// 定期持久化统计
	go l.saveStatistics()
}

func (l *Logger) saveStatistics() {
	l.mu.RLock()
	defer l.mu.RUnlock()

	data, err := json.MarshalIndent(l.stats, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(l.statsFile, data, 0644)
}

// GetStatistics 获取统计数据
func (l *Logger) GetStatistics() PersistentStatistics {
	l.mu.RLock()
	defer l.mu.RUnlock()

	statsCopy := l.stats
	statsCopy.LastUpdated = time.Now().Unix()
	return statsCopy
}

// ==================== 清理操作 ====================

// Clear 清空所有日志
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logs = nil
	l.requestLogs = nil

	os.Remove(l.logFile)
	os.Remove(l.requestFile)
}

// CleanOldLogs 清理超过保留期的日志
func (l *Logger) CleanOldLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().Unix()
	retentionSec := int64(l.retentionDays) * 24 * 60 * 60

	filtered := make([]LogEntry, 0)
	for _, log := range l.logs {
		if now-log.Timestamp < retentionSec {
			filtered = append(filtered, log)
		}
	}
	l.logs = filtered

	requestFiltered := make([]RequestLogEntry, 0)
	for _, log := range l.requestLogs {
		if now-log.Timestamp < retentionSec {
			requestFiltered = append(requestFiltered, log)
		}
	}
	l.requestLogs = requestFiltered
}

// ==================== 配置方法 ====================

// SetMaxLogs 设置最大日志数量
func (l *Logger) SetMaxLogs(max int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxLogs = max
}

// SetRetentionDays 设置保留天数
func (l *Logger) SetRetentionDays(days int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.retentionDays = days
}

// GetBaseDir 获取基础目录
func (l *Logger) GetBaseDir() string {
	return l.baseDir
}

// ==================== 辅助 ====================

func (l *Logger) consolePrint(entry LogEntry) {
	level := strings.ToUpper(string(entry.Level))
	timeStr := time.Unix(entry.Timestamp, 0).Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] [%s] %s\n", timeStr, level, entry.Message)
}

// ExportLogs 导出日志为字符串
func (l *Logger) ExportLogs(format string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if format == "json" {
		data, _ := json.MarshalIndent(l.logs, "", "  ")
		return string(data)
	}

	// txt format
	var sb strings.Builder
	for _, log := range l.logs {
		timeStr := time.Unix(log.Timestamp, 0).Format("2006-01-02T15:04:05Z")
		level := strings.ToUpper(string(log.Level))
		line := fmt.Sprintf("[%s] [%5s] %s", timeStr, level, log.Message)
		if log.ProviderID != "" {
			line += fmt.Sprintf(" | Provider: %s", log.ProviderID)
		}
		if log.AccountID != "" {
			line += fmt.Sprintf(" | Account: %s", log.AccountID)
		}
		if log.RequestID != "" {
			line += fmt.Sprintf(" | Request: %s", log.RequestID)
		}
		if log.Data != nil && len(log.Data) > 0 {
			dataStr, _ := json.Marshal(log.Data)
			line += fmt.Sprintf(" | Data: %s", string(dataStr))
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// ==================== 便捷方法 ====================

// LogOptions 日志查询选项
type LogOptions struct {
	Level   string `json:"level"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	Keyword string `json:"keyword"`
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	Items []LogEntry `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

// GetLogsByLevel 根据级别获取日志
func (l *Logger) GetLogsByLevel(level string, limit int) []LogEntry {
	filter := LogFilter{
		Level: LogLevel(level),
		Limit: limit,
	}
	return l.GetLogs(filter)
}

// GetErrorLogs 获取错误日志
func (l *Logger) GetErrorLogs(limit int) []LogEntry {
	return l.GetLogsByLevel("error", limit)
}

// GetRecentLogs 获取最近日志
func (l *Logger) GetRecentLogs(limit int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]LogEntry, 0, limit)
	start := len(l.logs) - limit
	if start < 0 {
		start = 0
	}
	for i := start; i < len(l.logs); i++ {
		result = append(result, l.logs[i])
	}
	return result
}

// ClearLogs 清空日志
func (l *Logger) ClearLogs() bool {
	l.Clear()
	return true
}

// GetLogsPaginated 获取分页日志
func (l *Logger) GetLogsPaginated(options LogOptions) *PaginatedResult {
	filter := LogFilter{
		Level:   LogLevel(options.Level),
		Limit:   options.Limit,
		Offset:  options.Offset,
		Keyword: options.Keyword,
	}
	items := l.GetLogs(filter)
	l.mu.RLock()
	total := len(l.logs)
	l.mu.RUnlock()
	return &PaginatedResult{
		Items: items,
		Total: total,
		Page:  options.Offset/options.Limit + 1,
		Size:  len(items),
	}
}

// ==================== 辅助函数 ====================

func parseInt64(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

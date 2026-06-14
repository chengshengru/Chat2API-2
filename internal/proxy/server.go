package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"chat2api-wails/internal/logger"
	"chat2api-wails/internal/store"
	"chat2api-wails/internal/types"
)

// ProxyStatus 代理服务器状态
type ProxyStatus struct {
	IsRunning bool   `json:"isRunning"`
	Port      int    `json:"port"`
	Host      string `json:"host"`
	Uptime    int64  `json:"uptime"`
	StartedAt int64  `json:"startedAt"`
}

// Statistics 代理服务器统计
type Statistics struct {
	TotalRequests    int64          `json:"totalRequests"`
	SuccessRequests  int64          `json:"successRequests"`
	FailedRequests   int64          `json:"failedRequests"`
	ActiveConnections int64          `json:"activeConnections"`
	TotalLatency     int64          `json:"totalLatency"`
	LastUpdated      int64          `json:"lastUpdated"`
	ModelUsage       map[string]int `json:"modelUsage"`
	ProviderUsage    map[string]int `json:"providerUsage"`
}

// ForwardRequest 转发请求结构
type ForwardRequest struct {
	Method      string
	URL         string
	Headers     map[string]string
	Body        []byte
	Timeout     time.Duration
	IsStream    bool
	ProviderID  string
	AccountID   string
	Model       string
}

// ProxyServer 代理服务器核心
type ProxyServer struct {
	mu            sync.RWMutex
	server        *http.Server
	router        *http.ServeMux
	status        ProxyStatus
	stats         Statistics
	logger        *logger.Logger
	storeManager  *store.StoreManager
}

// NewServer 创建新的代理服务器
func NewServer(sm *store.StoreManager, l *logger.Logger) *ProxyServer {
	s := &ProxyServer{
		logger:       l,
		storeManager: sm,
		stats: Statistics{
			ModelUsage:    make(map[string]int),
			ProviderUsage: make(map[string]int),
		},
	}

	s.setupRoutes()
	return s
}

// Start 启动代理服务器
func (ps *ProxyServer) Start(port int, host string) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.status.IsRunning {
		return true
	}

	if port <= 0 {
		port = 8080
	}
	if host == "" {
		host = "127.0.0.1"
	}

	ps.status.Port = port
	ps.status.Host = host
	ps.status.StartedAt = time.Now().Unix()

	// 启动 HTTP 服务器
	go ps.startHTTPServer()

	time.Sleep(100 * time.Millisecond)
	ps.status.IsRunning = true

	ps.logger.Info("Proxy server started",
		logger.Field{Key: "host", Value: host},
		logger.Field{Key: "port", Value: fmt.Sprintf("%d", port)},
	)

	return true
}

// Stop 停止代理服务器
func (ps *ProxyServer) Stop() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.status.IsRunning {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ps.server.Shutdown(ctx); err != nil {
		ps.logger.Error("Proxy server shutdown error", logger.Field{Key: "error", Value: err.Error()})
		return false
	}

	ps.status.IsRunning = false
	ps.logger.Info("Proxy server stopped")

	return true
}

// GetStatus 获取服务器状态
func (ps *ProxyServer) GetStatus() *ProxyStatus {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	status := ps.status
	if status.IsRunning {
		status.Uptime = time.Now().Unix() - status.StartedAt
	}

	return &status
}

// GetStatistics 获取统计信息
func (ps *ProxyServer) GetStatistics() *Statistics {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	stats := ps.stats
	stats.LastUpdated = time.Now().Unix()
	return &stats
}

// ==================== HTTP 服务器实现 ====================

func (ps *ProxyServer) startHTTPServer() {
	ps.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", ps.status.Host, ps.status.Port),
		Handler:      ps.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := ps.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		ps.logger.Error("Proxy server error", logger.Field{Key: "error", Value: err.Error()})
		ps.mu.Lock()
		ps.status.IsRunning = false
		ps.mu.Unlock()
	}
}

// setupRoutes 设置路由
func (ps *ProxyServer) setupRoutes() {
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("/", ps.handleHealth)
	mux.HandleFunc("/health", ps.handleHealth)
	mux.HandleFunc("/stats", ps.handleStats)

	// OpenAI 兼容接口
	mux.HandleFunc("/v1/chat/completions", ps.handleChatCompletions)
	mux.HandleFunc("/v1/models", ps.handleModels)
	mux.HandleFunc("/v1/models/", ps.handleGetModel)

	ps.router = mux
}

// handleHealth 健康检查处理
func (ps *ProxyServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := ps.GetStatus()
	stats := ps.GetStatistics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     status,
		"statistics": stats,
	})
}

// handleStats 统计信息处理
func (ps *ProxyServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := ps.GetStatistics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleChatCompletions 聊天完成处理
func (ps *ProxyServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// API Key 认证
	if !ps.validateApiKey(r) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ps.logger.Error("Failed to read request body", logger.Field{Key: "error", Value: err.Error()})
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 解析请求以获取模型名称
	var req struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		ps.logger.Error("Failed to parse request", logger.Field{Key: "error", Value: err.Error()})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ps.logger.Info("Chat request",
		logger.Field{Key: "model", Value: req.Model},
		logger.Field{Key: "stream", Value: fmt.Sprintf("%t", req.Stream)},
	)

	// 选择提供商和账户
	provider, account, err := ps.selectProviderAndAccount(req.Model)
	if err != nil {
		ps.logger.Error("No provider available", logger.Field{Key: "error", Value: err.Error()})
		http.Error(w, "No provider available", http.StatusBadGateway)
		return
	}

	// 构建转发请求
	targetURL := provider.APIEndpoint
	if provider.ChatPath != "" {
		targetURL += provider.ChatPath
	} else {
		targetURL += "/chat/completions"
	}

	headers := ps.buildHeaders(provider, account, r)

	forwardReq := &ForwardRequest{
		Method:     http.MethodPost,
		URL:        targetURL,
		Headers:    headers,
		Body:       body,
		Timeout:    120 * time.Second,
		IsStream:   req.Stream,
		ProviderID: provider.ID,
		AccountID:  account.ID,
		Model:      req.Model,
	}

	ps.mu.Lock()
	ps.stats.TotalRequests++
	ps.stats.ActiveConnections++
	ps.mu.Unlock()

	// 执行转发
	if req.Stream {
		ps.forwardStream(forwardReq, w)
	} else {
		ps.forward(forwardReq, w)
	}

	ps.mu.Lock()
	ps.stats.ActiveConnections--
	ps.mu.Unlock()
}

// handleModels 获取所有模型
func (ps *ProxyServer) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !ps.validateApiKey(r) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	providers := ps.storeManager.GetProviders()
	models := make([]map[string]interface{}, 0)

	for _, p := range providers {
		if !p.Enabled {
			continue
		}
		for _, modelName := range p.SupportedModels {
			models = append(models, map[string]interface{}{
				"id":       modelName,
				"object":   "model",
				"created":  time.Now().Unix(),
				"owned_by": p.Name,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   models,
	})
}

// handleGetModel 获取单个模型
func (ps *ProxyServer) handleGetModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !ps.validateApiKey(r) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	modelName := r.URL.Path[len("/v1/models/"):]

	providers := ps.storeManager.GetProviders()
	for _, p := range providers {
		if !p.Enabled {
			continue
		}
		for _, m := range p.SupportedModels {
			if m == modelName {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":       modelName,
					"object":   "model",
					"created":  time.Now().Unix(),
					"owned_by": p.Name,
				})
				return
			}
		}
	}

	http.Error(w, "Model not found", http.StatusNotFound)
}

// ==================== 辅助方法 ====================

// validateApiKey 验证 API Key
func (ps *ProxyServer) validateApiKey(r *http.Request) bool {
	config := ps.storeManager.GetConfig()

	// 如果未启用 API Key 检查，允许通过
	if !config.EnableApiKey || len(config.ApiKeys) == 0 {
		return true
	}

	// 获取提供的 key
	providedKey := r.Header.Get("Authorization")
	if len(providedKey) > 7 && providedKey[:7] == "Bearer " {
		providedKey = providedKey[7:]
	} else if queryKey := r.URL.Query().Get("api_key"); queryKey != "" {
		providedKey = queryKey
	} else if headerKey := r.Header.Get("X-API-Key"); headerKey != "" {
		providedKey = headerKey
	} else {
		return false
	}

	// 检查 key 是否存在且启用
	for _, k := range config.ApiKeys {
		if k.Key == providedKey && k.Enabled {
			return true
		}
	}
	return false
}

// selectProviderAndAccount 选择提供商和账户
func (ps *ProxyServer) selectProviderAndAccount(model string) (*types.Provider, *types.Account, error) {
	providers := ps.storeManager.GetProviders()
	accounts := ps.storeManager.GetAccounts(false)

	// 简单策略：查找第一个启用的提供商及其第一个启用的账户
	for _, p := range providers {
		if !p.Enabled {
			continue
		}
		for _, a := range accounts {
			if a.ProviderID == p.ID && a.Status == types.AccountStatusActive {
				return &p, &a, nil
			}
		}
	}

	return nil, nil, errors.New("no available provider for model")
}

// buildHeaders 构建请求头
func (ps *ProxyServer) buildHeaders(provider *types.Provider, account *types.Account, r *http.Request) map[string]string {
	headers := make(map[string]string)

	// 复制原始请求头
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 提供商特定头
	for k, v := range provider.Headers {
		headers[k] = v
	}

	// 添加凭证
	if account.Credentials != nil {
		if token, ok := account.Credentials["token"]; ok && token != "" {
			headers["Authorization"] = "Bearer " + token
		}
		if cookie, ok := account.Credentials["cookie"]; ok && cookie != "" {
			headers["Cookie"] = cookie
		}
		if apiKey, ok := account.Credentials["api_key"]; ok && apiKey != "" {
			headers["api-key"] = apiKey
		}
		if authorization, ok := account.Credentials["Authorization"]; ok && authorization != "" {
			headers["Authorization"] = authorization
		}
	}

	return headers
}

// forward 非流式转发
func (ps *ProxyServer) forward(req *ForwardRequest, w http.ResponseWriter) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		ps.logger.Error("Failed to create request", logger.Field{Key: "error", Value: err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		ps.mu.Lock()
		ps.stats.FailedRequests++
		ps.mu.Unlock()
		return
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: req.Timeout,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		ps.logger.Error("Request failed", logger.Field{Key: "error", Value: err.Error()})
		http.Error(w, "Gateway error", http.StatusBadGateway)
		ps.mu.Lock()
		ps.stats.FailedRequests++
		ps.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for k, v := range resp.Header {
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	io.Copy(w, resp.Body)

	latency := time.Since(start).Milliseconds()

	ps.mu.Lock()
	ps.stats.SuccessRequests++
	ps.stats.TotalLatency += latency
	if req.Model != "" {
		ps.stats.ModelUsage[req.Model]++
	}
	if req.ProviderID != "" {
		ps.stats.ProviderUsage[req.ProviderID]++
	}
	ps.mu.Unlock()

	ps.storeManager.RecordRequest(true, latency, req.Model, req.ProviderID, req.AccountID)
}

// forwardStream 流式转发
func (ps *ProxyServer) forwardStream(req *ForwardRequest, w http.ResponseWriter) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		ps.logger.Error("Failed to create stream request", logger.Field{Key: "error", Value: err.Error()})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		ps.mu.Lock()
		ps.stats.FailedRequests++
		ps.mu.Unlock()
		return
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: req.Timeout,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		ps.logger.Error("Stream request failed", logger.Field{Key: "error", Value: err.Error()})
		http.Error(w, "Gateway error", http.StatusBadGateway)
		ps.mu.Lock()
		ps.stats.FailedRequests++
		ps.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for k, v := range resp.Header {
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}
	w.WriteHeader(resp.StatusCode)

	// 流式传输
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				ps.logger.Error("Stream write error", logger.Field{Key: "error", Value: writeErr.Error()})
				break
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			ps.logger.Error("Stream read error", logger.Field{Key: "error", Value: err.Error()})
			break
		}
	}

	latency := time.Since(start).Milliseconds()

	ps.mu.Lock()
	ps.stats.SuccessRequests++
	ps.stats.TotalLatency += latency
	if req.Model != "" {
		ps.stats.ModelUsage[req.Model]++
	}
	if req.ProviderID != "" {
		ps.stats.ProviderUsage[req.ProviderID]++
	}
	ps.mu.Unlock()

	ps.storeManager.RecordRequest(true, latency, req.Model, req.ProviderID, req.AccountID)
}

// Handler 返回 HTTP 处理器供内部使用
func (ps *ProxyServer) Handler() http.Handler {
	return ps.router
}

// Port 返回当前端口
func (ps *ProxyServer) Port() int {
	return ps.status.Port
}

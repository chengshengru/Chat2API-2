package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"chat2api-wails/internal/logger"
	"chat2api-wails/internal/store"
	"chat2api-wails/internal/types"
)

// ProxyServer HTTP代理服务器
type ProxyServer struct {
	mu      sync.RWMutex
	store   *store.StoreManager
	logger  *logger.Logger
	server  *http.Server
	router  *http.ServeMux
	status  types.ProxyStatus
	stats   types.Statistics

	// 负载均衡索引
	roundRobinIndex map[string]int
	failedAccounts  map[string]*failedAccountInfo
}

type failedAccountInfo struct {
	Count       int
	LastFailTime int64
}

// ProxyStatus 代理状态
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
	LastUpdated     int64                `json:"lastUpdated"`
}

// NewServer 创建代理服务器
func NewServer(sm *store.StoreManager, l *logger.Logger) *ProxyServer {
	ps := &ProxyServer{
		store:           sm,
		logger:          l,
		roundRobinIndex: make(map[string]int),
		failedAccounts:  make(map[string]*failedAccountInfo),
		status: types.ProxyStatus{
			IsRunning: false,
			Port:     8080,
			Host:     "127.0.0.1",
		},
		stats: types.Statistics{
			ModelUsage:    make(map[string]string),
			ProviderUsage: make(map[string]string),
			AccountUsage:  make(map[string]string),
		},
	}
	ps.setupRoutes()
	return ps
}

// setupRoutes 设置路由
func (ps *ProxyServer) setupRoutes() {
	ps.router = http.NewServeMux()

	// 健康检查
	ps.router.HandleFunc("/", ps.handleHealth)
	ps.router.HandleFunc("/health", ps.handleHealth)
	ps.router.HandleFunc("/stats", ps.handleStats)

	// OpenAI 兼容接口
	ps.router.HandleFunc("/v1/chat/completions", ps.handleChatCompletions)
	ps.router.HandleFunc("/v1/completions", ps.handleCompletions)
	ps.router.HandleFunc("/v1/models", ps.handleModels)
	ps.router.HandleFunc("/v1/models/", ps.handleGetModel)

	// Management API
	ps.router.HandleFunc("/v0/management/", ps.handleManagement)
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
func (ps *ProxyServer) GetStatus() *types.ProxyStatus {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	status := ps.status
	if status.IsRunning {
		status.Uptime = time.Now().Unix() - status.StartedAt
	}

	return &status
}

// GetStatistics 获取统计信息
func (ps *ProxyServer) GetStatistics() *types.Statistics {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	stats := ps.stats
	stats.LastUpdated = time.Now().Unix()
	return &stats
}

// startHTTPServer 启动HTTP服务器
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

// handleHealth 健康检查
func (ps *ProxyServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := ps.GetStatus()
	stats := ps.store.GetStatistics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     status,
		"statistics": stats,
	})
}

// handleStats 统计信息
func (ps *ProxyServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := ps.store.GetStatistics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleChatCompletions 处理聊天补全请求
func (ps *ProxyServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := fmt.Sprintf("chatcmpl-%d-%s", time.Now().Unix(), randomString(8))

	if r.Method != http.MethodPost {
		ps.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析请求
	var req types.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ps.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Model == "" {
		ps.writeError(w, http.StatusBadRequest, "Missing required field: model")
		return
	}

	if len(req.Messages) == 0 {
		ps.writeError(w, http.StatusBadRequest, "Missing required field: messages")
		return
	}

	// 获取配置
	config := ps.store.GetConfig()

	// API Key 验证
	if config.ProxyConfig.EnableApiKey {
		if !ps.validateAPIKey(r) {
			ps.writeError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}
	}

	// 负载均衡选择账户
	selection := ps.selectAccount(req.Model, config.ProxyConfig.LoadBalanceConfig.Strategy)
	if selection == nil {
		ps.writeError(w, http.StatusServiceUnavailable, "No available account for model: "+req.Model)
		return
	}

	account := selection.Account
	provider := selection.Provider

	// 记录请求开始
	ps.recordRequestStart(req.Model, provider.ID, account.ID)

	// 构建转发请求
	forwardReq, err := ps.buildForwardRequest(req, provider, account)
	if err != nil {
		ps.writeError(w, http.StatusInternalServerError, "Failed to build request: "+err.Error())
		return
	}

	// 发送请求
	resp, err := ps.forwardRequest(forwardReq, provider)
	if err != nil {
		ps.recordRequestFailure(time.Since(startTime).Milliseconds())
		ps.markAccountFailed(account.ID)
		ps.writeError(w, http.StatusBadGateway, "Forward request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 处理响应
	latency := time.Since(startTime).Milliseconds()

	if req.Stream {
		ps.handleStreamResponse(w, resp, req.Model, requestID, account.ID, provider.ID, latency)
	} else {
		ps.handleNonStreamResponse(w, resp, req.Model, account.ID, provider.ID, latency)
	}
}

// handleCompletions 处理补全请求（兼容旧API）
func (ps *ProxyServer) handleCompletions(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := fmt.Sprintf("cmpl-%d-%s", time.Now().Unix(), randomString(8))

	if r.Method != http.MethodPost {
		ps.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Model       string   `json:"model"`
		Prompt      string   `json:"prompt"`
		MaxTokens   *int     `json:"max_tokens,omitempty"`
		Temperature *float64 `json:"temperature,omitempty"`
		Stream      bool     `json:"stream,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ps.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Model == "" || req.Prompt == "" {
		ps.writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// 将 prompt 转换为 messages 格式
	messages := []types.ChatMessage{
		{Role: "user", Content: req.Prompt},
	}

	chatReq := types.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:     req.Stream,
	}

	// 复用聊天补全逻辑
	config := ps.store.GetConfig()

	if config.ProxyConfig.EnableApiKey {
		if !ps.validateAPIKey(r) {
			ps.writeError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}
	}

	selection := ps.selectAccount(req.Model, config.ProxyConfig.LoadBalanceConfig.Strategy)
	if selection == nil {
		ps.writeError(w, http.StatusServiceUnavailable, "No available account")
		return
	}

	forwardReq, err := ps.buildForwardRequest(chatReq, selection.Provider, selection.Account)
	if err != nil {
		ps.writeError(w, http.StatusInternalServerError, "Failed to build request")
		return
	}

	resp, err := ps.forwardRequest(forwardReq, selection.Provider)
	if err != nil {
		ps.writeError(w, http.StatusBadGateway, "Forward request failed")
		return
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Milliseconds()

	if req.Stream {
		ps.handleStreamResponse(w, resp, req.Model, requestID, selection.Account.ID, selection.Provider.ID, latency)
	} else {
		ps.handleNonStreamResponse(w, resp, req.Model, selection.Account.ID, selection.Provider.ID, latency)
	}
}

// handleModels 处理模型列表请求
func (ps *ProxyServer) handleModels(w http.ResponseWriter, r *http.Request) {
	models := ps.getAvailableModels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   models,
	})
}

// handleGetModel 处理单个模型请求
func (ps *ProxyServer) handleGetModel(w http.ResponseWriter, r *http.Request) {
	modelID := strings.TrimPrefix(r.URL.Path, "/v1/models/")

	models := ps.getAvailableModels()
	for _, model := range models {
		if model["id"] == modelID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"object": "model",
				"id":     model["id"],
				"name":   model["name"],
			})
			return
		}
	}

	ps.writeError(w, http.StatusNotFound, "Model not found")
}

// handleManagement 处理管理API请求
func (ps *ProxyServer) handleManagement(w http.ResponseWriter, r *http.Request) {
	config := ps.store.GetConfig()

	if config.ManagementConfig.AuthToken != "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+config.ManagementConfig.AuthToken {
			ps.writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}

	path := strings.TrimPrefix(r.URL.Path, "/v0/management")
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)

	if len(parts) == 0 {
		ps.writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	switch parts[0] {
	case "providers":
		ps.handleManagementProviders(w, r, parts)
	case "accounts":
		ps.handleManagementAccounts(w, r, parts)
	case "config":
		ps.handleManagementConfig(w, r)
	case "statistics":
		ps.handleManagementStatistics(w, r)
	case "sessions":
		ps.handleManagementSessions(w, r, parts)
	default:
		ps.writeError(w, http.StatusNotFound, "Not found")
	}
}

// ==================== 辅助方法 ====================

func (ps *ProxyServer) validateAPIKey(r *http.Request) bool {
	config := ps.store.GetConfig()

	authHeader := r.Header.Get("Authorization")
	var providedKey string

	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			providedKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	} else {
		providedKey = r.URL.Query().Get("api_key")
		if providedKey == "" {
			providedKey = r.Header.Get("X-API-Key")
		}
	}

	if providedKey == "" {
		return false
	}

	for _, key := range config.ProxyConfig.ApiKeys {
		if key.Key == providedKey && key.Enabled {
			return true
		}
	}

	return false
}

func (ps *ProxyServer) getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return r.RemoteAddr
}

type accountSelection struct {
	Account     *types.Account
	Provider    *types.Provider
	ActualModel string
}

func (ps *ProxyServer) selectAccount(model string, strategy types.LoadBalanceStrategy) *accountSelection {
	providers := ps.store.GetAllProviders()
	var candidates []*accountSelection

	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}

		accounts := ps.store.GetAccountsByProviderID(provider.ID, true)
		for i := range accounts {
			account := &accounts[i]
			if !ps.isAccountAvailable(account.ID) {
				continue
			}

			actualModel := ps.mapModel(model, provider)
			candidates = append(candidates, &accountSelection{
				Account:     account,
				Provider:    &provider,
				ActualModel: actualModel,
			})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// 根据策略选择
	switch strategy {
	case types.LoadBalanceFillFirst:
		return ps.selectFillFirst(candidates)
	case types.LoadBalanceFailover:
		return ps.selectFailover(candidates)
	default: // round-robin
		return ps.selectRoundRobin(candidates)
	}
}

func (ps *ProxyServer) isAccountAvailable(accountID string) bool {
	account, err := ps.store.GetAccountByID(accountID)
	if err != nil {
		return false
	}

	if account.Status != types.AccountStatusActive {
		return false
	}

	info, exists := ps.failedAccounts[accountID]
	if exists {
		if time.Now().Unix()-info.LastFailTime < 60 { // 1分钟恢复时间
			if info.Count >= 3 {
				return false
			}
		}
	}

	return true
}

func (ps *ProxyServer) selectRoundRobin(candidates []*accountSelection) *accountSelection {
	if len(candidates) == 0 {
		return nil
	}

	key := candidates[0].Provider.ID
	index := ps.roundRobinIndex[key]
	selection := candidates[index%len(candidates)]

	ps.roundRobinIndex[key] = index + 1
	return selection
}

func (ps *ProxyServer) selectFillFirst(candidates []*accountSelection) *accountSelection {
	if len(candidates) == 0 {
		return nil
	}

	// 选择使用最少的账户
	var selected *accountSelection
	minRequests := int64(0x7fffffffffffffff)

	for _, c := range candidates {
		if c.Account.RequestCount < minRequests {
			minRequests = c.Account.RequestCount
			selected = c
		}
	}

	return selected
}

func (ps *ProxyServer) selectFailover(candidates []*accountSelection) *accountSelection {
	if len(candidates) == 0 {
		return nil
	}

	// 按请求数排序，选择使用最多的（假设最可靠的）
	var selected *accountSelection
	maxRequests := int64(0)

	for _, c := range candidates {
		if c.Account.RequestCount > maxRequests {
			maxRequests = c.Account.RequestCount
			selected = c
		}
	}

	return selected
}

func (ps *ProxyServer) mapModel(model string, provider types.Provider) string {
	// 简单的模型映射逻辑
	switch provider.ID {
	case "deepseek":
		if strings.HasPrefix(model, "gpt-") {
			return "deepseek-chat"
		}
	case "glm":
		if strings.HasPrefix(model, "gpt-") {
			return "glm-4"
		}
	case "kimi":
		if strings.HasPrefix(model, "gpt-") {
			return "moonshot-v1-8k"
		}
	case "qwen":
		if strings.HasPrefix(model, "gpt-") {
			return "qwen-turbo"
		}
	}
	return model
}

func (ps *ProxyServer) recordRequestStart(model, providerID, accountID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.stats.TotalRequests++
}

func (ps *ProxyServer) recordRequestFailure(latency int64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.stats.FailedRequests++
	ps.stats.TotalLatency += latency
}

func (ps *ProxyServer) markAccountFailed(accountID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	info, exists := ps.failedAccounts[accountID]
	if !exists {
		info = &failedAccountInfo{}
		ps.failedAccounts[accountID] = info
	}

	info.Count++
	info.LastFailTime = time.Now().Unix()
}

func (ps *ProxyServer) buildForwardRequest(req types.ChatCompletionRequest, provider *types.Provider, account *types.Account) (*http.Request, error) {
	// 获取解密的凭证
	creds, err := ps.store.GetDecryptedCredentials(account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// 构建请求URL
	chatPath := provider.ChatPath
	if chatPath == "" {
		chatPath = "/chat/completions"
	}
	url := provider.APIEndpoint + chatPath

	// 转换请求格式（根据提供商）
	body := ps.convertRequestForProvider(req, provider, creds)

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置头部
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	for k, v := range provider.Headers {
		httpReq.Header.Set(k, v)
	}

	// 根据认证类型设置授权头
	switch provider.AuthType {
	case types.AuthTypeToken, types.AuthTypeUserToken:
		if token, ok := creds["token"]; ok {
			httpReq.Header.Set("Authorization", "Bearer "+token)
		}
	case types.AuthTypeRefreshToken:
		if refreshToken, ok := creds["refresh_token"]; ok {
			httpReq.Header.Set("Authorization", "Bearer "+refreshToken)
		}
	}

	return httpReq, nil
}

func (ps *ProxyServer) convertRequestForProvider(req types.ChatCompletionRequest, provider *types.Provider, creds map[string]string) interface{} {
	// 转换为各提供商特定的格式
	switch provider.ID {
	case "deepseek":
		return ps.convertForDeepSeek(req, creds)
	case "glm":
		return ps.convertForGLM(req, creds)
	case "kimi":
		return ps.convertForKimi(req, creds)
	case "qwen":
		return ps.convertForQwen(req, creds)
	default:
		return req
	}
}

func (ps *ProxyServer) convertForDeepSeek(req types.ChatCompletionRequest, creds map[string]string) map[string]interface{} {
	// DeepSeek 特定转换
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		body["max_tokens"] = *req.MaxTokens
	}
	if req.WebSearch {
		body["web_search"] = true
	}
	if req.ReasoningEffort != "" {
		body["reasoning_effort"] = req.ReasoningEffort
	}
	if len(req.Tools) > 0 {
		body["tools"] = req.Tools
	}

	return body
}

func (ps *ProxyServer) convertForGLM(req types.ChatCompletionRequest, creds map[string]string) map[string]interface{} {
	// GLM 特定转换
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		body["max_tokens"] = *req.MaxTokens
	}
	if req.DeepResearch {
		body["deep_research"] = true
	}
	if len(req.Tools) > 0 {
		body["tools"] = req.Tools
	}

	return body
}

func (ps *ProxyServer) convertForKimi(req types.ChatCompletionRequest, creds map[string]string) map[string]interface{} {
	// Kimi 特定转换
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		body["max_tokens"] = *req.MaxTokens
	}

	return body
}

func (ps *ProxyServer) convertForQwen(req types.ChatCompletionRequest, creds map[string]string) map[string]interface{} {
	// Qwen 特定转换
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		body["max_tokens"] = *req.MaxTokens
	}

	return body
}

func (ps *ProxyServer) forwardRequest(req *http.Request, provider *types.Provider) (*http.Response, error) {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (ps *ProxyServer) handleStreamResponse(w http.ResponseWriter, resp *http.Response, model, requestID, accountID, providerID string, latency int64) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		ps.writeError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	reader := bufio.NewReader(resp.Body)
	created := time.Now().Unix()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				ps.logger.Error("Stream read error", logger.Field{Key: "error", Value: err.Error()})
			}
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析 SSE 格式
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				break
			}

			// 转换响应格式
			transformed := ps.transformStreamData(data, model, requestID, created)
			fmt.Fprintf(w, "data: %s\n\n", transformed)
			flusher.Flush()
		}
	}

	// 记录统计
	ps.store.RecordRequest(true, latency, model, providerID, accountID)
	ps.store.MarkAccountUsed(accountID)
}

func (ps *ProxyServer) transformStreamData(data, model, requestID string, created int64) string {
	// 尝试解析并转换
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return data
	}

	// 确保响应格式符合 OpenAI 标准
	if id, ok := chunk["id"].(string); !ok || id == "" {
		chunk["id"] = requestID
	}
	if _, ok := chunk["object"].(string); !ok {
		chunk["object"] = "chat.completion.chunk"
	}
	chunk["created"] = created
	if _, ok := chunk["model"].(string); !ok {
		chunk["model"] = model
	}

	result, _ := json.Marshal(chunk)
	return string(result)
}

func (ps *ProxyServer) handleNonStreamResponse(w http.ResponseWriter, resp *http.Response, model, accountID, providerID string, latency int64) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ps.writeError(w, http.StatusBadGateway, "Failed to read response")
		return
	}

	// 转换响应格式
	transformed := ps.transformResponse(body, model)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(transformed)

	// 记录统计
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	ps.store.RecordRequest(success, latency, model, providerID, accountID)

	if success {
		ps.store.MarkAccountUsed(accountID)
	} else {
		ps.markAccountFailed(accountID)
	}
}

func (ps *ProxyServer) transformResponse(body []byte, model string) []byte {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return body
	}

	// 确保响应格式符合 OpenAI 标准
	if id, ok := resp["id"].(string); !ok || id == "" {
		resp["id"] = fmt.Sprintf("chatcmpl-%d", time.Now().Unix())
	}
	if _, ok := resp["object"].(string); !ok {
		resp["object"] = "chat.completion"
	}
	resp["created"] = time.Now().Unix()
	if _, ok := resp["model"].(string); !ok {
		resp["model"] = model
	}

	result, _ := json.Marshal(resp)
	return result
}

func (ps *ProxyServer) getAvailableModels() []map[string]string {
	models := []map[string]string{
		{"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo"},
		{"id": "gpt-4", "name": "GPT-4"},
		{"id": "gpt-4-turbo", "name": "GPT-4 Turbo"},
		{"id": "deepseek-chat", "name": "DeepSeek Chat"},
		{"id": "deepseek-coder", "name": "DeepSeek Coder"},
		{"id": "glm-4", "name": "GLM-4"},
		{"id": "glm-4-flash", "name": "GLM-4 Flash"},
		{"id": "moonshot-v1-8k", "name": "Moonshot V1 8K"},
		{"id": "moonshot-v1-32k", "name": "Moonshot V1 32K"},
		{"id": "qwen-turbo", "name": "Qwen Turbo"},
		{"id": "qwen-plus", "name": "Qwen Plus"},
		{"id": "qwen-max", "name": "Qwen Max"},
		{"id": "minimax-chat", "name": "MiniMax Chat"},
		{"id": "mimo-chat", "name": "Mimo Chat"},
		{"id": "perplexity", "name": "Perplexity"},
	}

	return models
}

func (ps *ProxyServer) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "invalid_request_error",
			"code":    status,
		},
	})
}

// ==================== Management API 处理器 ====================

func (ps *ProxyServer) handleManagementProviders(w http.ResponseWriter, r *http.Request, parts []string) {
	switch r.Method {
	case http.MethodGet:
		if len(parts) > 1 && parts[1] != "" {
			ps.handleGetProvider(w, r, parts[1])
		} else {
			ps.handleListProviders(w, r)
		}
	case http.MethodPost:
		ps.handleCreateProvider(w, r)
	default:
		ps.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (ps *ProxyServer) handleListProviders(w http.ResponseWriter, r *http.Request) {
	providers := ps.store.GetAllProviders()
	ps.writeJSON(w, map[string]interface{}{"success": true, "data": providers})
}

func (ps *ProxyServer) handleGetProvider(w http.ResponseWriter, r *http.Request, id string) {
	provider, err := ps.store.GetProviderByID(id)
	if err != nil {
		ps.writeError(w, http.StatusNotFound, "Provider not found")
		return
	}
	ps.writeJSON(w, map[string]interface{}{"success": true, "data": provider})
}

func (ps *ProxyServer) handleCreateProvider(w http.ResponseWriter, r *http.Request) {
	var req types.Provider
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ps.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := ps.store.CreateProvider(req); err != nil {
		ps.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ps.writeJSON(w, map[string]interface{}{"success": true})
}

func (ps *ProxyServer) handleManagementAccounts(w http.ResponseWriter, r *http.Request, parts []string) {
	switch r.Method {
	case http.MethodGet:
		if len(parts) > 1 && parts[1] != "" {
			ps.handleGetAccount(w, r, parts[1])
		} else {
			ps.handleListAccounts(w, r)
		}
	case http.MethodPost:
		ps.handleCreateAccount(w, r)
	default:
		ps.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (ps *ProxyServer) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := ps.store.GetAllAccounts()
	ps.writeJSON(w, map[string]interface{}{"success": true, "data": accounts})
}

func (ps *ProxyServer) handleGetAccount(w http.ResponseWriter, r *http.Request, id string) {
	account, err := ps.store.GetAccountByID(id)
	if err != nil {
		ps.writeError(w, http.StatusNotFound, "Account not found")
		return
	}
	ps.writeJSON(w, map[string]interface{}{"success": true, "data": account})
}

func (ps *ProxyServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProviderID  string            `json:"providerId"`
		Name        string            `json:"name"`
		Credentials map[string]string `json:"credentials"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ps.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	account, err := ps.store.CreateAccount(req.ProviderID, req.Name, req.Credentials)
	if err != nil {
		ps.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ps.writeJSON(w, map[string]interface{}{"success": true, "data": account})
}

func (ps *ProxyServer) handleManagementConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		config := ps.store.GetConfig()
		ps.writeJSON(w, map[string]interface{}{"success": true, "data": config})
	} else if r.Method == http.MethodPut {
		var config types.AppConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			ps.writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if err := ps.store.UpdateConfig(config); err != nil {
			ps.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		ps.writeJSON(w, map[string]interface{}{"success": true})
	} else {
		ps.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (ps *ProxyServer) handleManagementStatistics(w http.ResponseWriter, r *http.Request) {
	stats := ps.store.GetStatistics()
	ps.writeJSON(w, map[string]interface{}{"success": true, "data": stats})
}

func (ps *ProxyServer) handleManagementSessions(w http.ResponseWriter, r *http.Request, parts []string) {
	ps.writeError(w, http.StatusNotImplemented, "Sessions API not implemented")
}

func (ps *ProxyServer) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ==================== 工具函数 ====================

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// Port 返回当前端口
func (ps *ProxyServer) Port() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.status.Port
}

// IsRunning 返回是否运行中
func (ps *ProxyServer) IsRunning() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.status.IsRunning
}

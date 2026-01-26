package gost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gost-panel/internal/errors"
	"io"
	"net/http"
	"time"

	"gost-panel/pkg/logger"
)

// Client Gost API 客户端
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// Config 客户端配置
type Config struct {
	APIURL   string // API 地址，如 http://ip:port
	Username string
	Password string
	Timeout  time.Duration
}

// NewClient 创建 Gost 客户端
func NewClient(cfg *Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	// 移除尾部斜杠
	apiURL := cfg.APIURL
	if len(apiURL) > 0 && apiURL[len(apiURL)-1] == '/' {
		apiURL = apiURL[:len(apiURL)-1]
	}

	return &Client{
		baseURL:  apiURL,
		username: cfg.Username,
		password: cfg.Password,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name      string           `json:"name"`
	Addr      string           `json:"addr"`
	Handler   *HandlerConfig   `json:"handler,omitempty"`
	Listener  *ListenerConfig  `json:"listener,omitempty"`
	Forwarder *ForwarderConfig `json:"forwarder,omitempty"`
	Limiter   string           `json:"limiter,omitempty"`  // 流量速率限制器名称
	CLimiter  string           `json:"climiter,omitempty"` // 并发连接数限制器名称
	RLimiter  string           `json:"rlimiter,omitempty"` // 请求速率限制器名称
	Observer  string           `json:"observer,omitempty"` // 观察器名称
	Metadata  map[string]any   `json:"metadata,omitempty"` // 元数据配置
	Status    *ServiceStatus   `json:"status,omitempty"`   // 服务运行状态
}

// ServiceStatus 服务运行时状态信息
type ServiceStatus struct {
	State string `json:"state"` // 状态: ready, running, failed, closed.
}

// GostResponse 通用 API 响应包裹
type GostResponse struct {
	Data json.RawMessage `json:"data"`
}

// HandlerConfig 处理器配置
type HandlerConfig struct {
	Type  string      `json:"type"`
	Chain string      `json:"chain,omitempty"` // 链名称
	Auth  *AuthConfig `json:"auth,omitempty"`
}

// ListenerConfig 监听器配置
type ListenerConfig struct {
	Type string `json:"type"`
}

// ForwarderConfig 转发器配置
type ForwarderConfig struct {
	Nodes    []*ForwarderNode `json:"nodes"`
	Selector *SelectorConfig  `json:"selector,omitempty"`
}

// SelectorConfig 选择器配置
type SelectorConfig struct {
	Strategy    string        `json:"strategy,omitempty"`
	MaxFails    int           `json:"maxFails,omitempty"`
	FailTimeout time.Duration `json:"failTimeout,omitempty"`
}

// ForwarderNode 转发目标节点
type ForwarderNode struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ChainConfig 链配置
type ChainConfig struct {
	Name string       `json:"name"`
	Hops []*HopConfig `json:"hops,omitempty"`
}

// HopConfig 跳配置
type HopConfig struct {
	Name  string        `json:"name"`
	Nodes []*NodeConfig `json:"nodes,omitempty"`
}

// NodeConfig 链节点配置
type NodeConfig struct {
	Name      string           `json:"name"`
	Addr      string           `json:"addr"`
	Connector *ConnectorConfig `json:"connector,omitempty"`
	Dialer    *DialerConfig    `json:"dialer,omitempty"`
}

// ConnectorConfig 连接器配置
type ConnectorConfig struct {
	Type string      `json:"type"`
	Auth *AuthConfig `json:"auth,omitempty"`
}

// DialerConfig 拨号器配置
type DialerConfig struct {
	Type string `json:"type"`
}

// LimiterConfig 流量速率限制器配置
// limits 格式: "$ 100MB 100MB" (服务级别 入站 出站)
// limits 格式: "$$ 10MB" (连接级别)
// limits 格式: "192.168.1.1 1MB 5MB" (IP级别)
type LimiterConfig struct {
	Name   string        `json:"name"`
	Limits []string      `json:"limits,omitempty"` // 限制规则列表
	Plugin *PluginConfig `json:"plugin,omitempty"` // 插件配置
}

// CLimiterConfig 并发连接数限制器配置
// limits 格式: "$ 1000" (服务级别最大连接数)
// limits 格式: "$$ 100" (IP级别默认最大连接数)
type CLimiterConfig struct {
	Name   string        `json:"name"`
	Limits []string      `json:"limits,omitempty"` // 限制规则列表
	Plugin *PluginConfig `json:"plugin,omitempty"` // 插件配置
}

// RLimiterConfig 请求速率限制器配置
// limits 格式: "$ 100" (服务级别每秒请求数)
// limits 格式: "$$ 10" (IP级别默认每秒请求数)
type RLimiterConfig struct {
	Name   string        `json:"name"`
	Limits []string      `json:"limits,omitempty"` // 限制规则列表
	Plugin *PluginConfig `json:"plugin,omitempty"` // 插件配置
}

// ObserverConfig 观察器配置
type ObserverConfig struct {
	Name   string        `json:"name"`
	Plugin *PluginConfig `json:"plugin,omitempty"` // 插件配置
}

// PluginConfig 插件配置
type PluginConfig struct {
	Type    string `json:"type,omitempty"`    // 插件类型: grpc, http
	Addr    string `json:"addr,omitempty"`    // 插件地址
	Timeout string `json:"timeout,omitempty"` // 超时时间
}

// HealthCheck 健康检查
func (c *Client) HealthCheck() error {
	resp, err := c.doRequest("GET", "/config", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// GostConfig Gost 完整配置
type GostConfig struct {
	Services  []ServiceConfig  `json:"services"`
	Chains    []ChainConfig    `json:"chains"`
	Limiters  []LimiterConfig  `json:"limiters"`
	CLimiters []CLimiterConfig `json:"climiters"`
	RLimiters []RLimiterConfig `json:"rlimiters"`
	Observers []ObserverConfig `json:"observers"`
}

// GetConfig 获取节点配置
func (c *Client) GetConfig() (*GostConfig, error) {
	resp, err := c.doRequest("GET", "/config", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取配置失败: %s", string(body))
	}

	var config GostConfig
	if err = json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}

	return &config, nil
}

// SaveConfig 保存节点配置
func (c *Client) SaveConfig() error {
	resp, err := c.doRequest("POST", "/config?format=yaml", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("保存配置失败: %s", string(body))
	}

	return nil
}

// exists 检查指定路径的资源是否存在
func (c *Client) exists(path string) bool {
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var gResp GostResponse
	if err = json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return false
	}

	// 根据用户提供的 curl 结果，不存在时 data 为 null (RawMessage 为 "null")
	return string(gResp.Data) != "null" && len(gResp.Data) > 0
}

// CreateService 创建服务 (幂等)
func (c *Client) CreateService(svc *ServiceConfig) error {
	path := fmt.Sprintf("/config/services/%s", svc.Name)
	if c.exists(path) {
		logger.Debugf("服务 %s 已存在，跳过创建", svc.Name)
		return nil
	}

	resp, err := c.doRequest("POST", "/config/services", svc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建服务失败: %s", string(body))
	}

	return nil
}

// DeleteService 删除服务 (幂等)
func (c *Client) DeleteService(name string) error {
	path := fmt.Sprintf("/config/services/%s", name)
	if !c.exists(path) {
		logger.Debugf("服务 %s 不存在，跳过删除", name)
		return nil
	}

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除服务失败: %s", string(body))
	}

	return nil
}

// CreateChain 创建链 (幂等)
func (c *Client) CreateChain(chain *ChainConfig) error {
	path := fmt.Sprintf("/config/chains/%s", chain.Name)
	if c.exists(path) {
		logger.Debugf("链 %s 已存在，跳过创建", chain.Name)
		return nil
	}

	resp, err := c.doRequest("POST", "/config/chains", chain)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建链失败: %s", string(body))
	}

	return nil
}

// DeleteChain 删除链 (幂等)
func (c *Client) DeleteChain(name string) error {
	path := fmt.Sprintf("/config/chains/%s", name)
	if !c.exists(path) {
		logger.Debugf("链 %s 不存在，跳过删除", name)
		return nil
	}

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除链失败: %s", string(body))
	}

	return nil
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加基础认证
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	return c.httpClient.Do(req)
}

// BuildTCPForwardService 构建 TCP 转发服务配置
func BuildTCPForwardService(name string, listenPort int, targets []string, strategy string) *ServiceConfig {
	nodes := make([]*ForwarderNode, 0)
	for i, target := range targets {
		nodes = append(nodes, &ForwarderNode{
			Name: fmt.Sprintf("target-%d", i),
			Addr: target,
		})
	}

	// 默认策略
	if strategy == "" {
		strategy = "round"
	}

	return &ServiceConfig{
		Name: name,
		Addr: fmt.Sprintf(":%d", listenPort),
		Handler: &HandlerConfig{
			Type: "rtcp",
		},
		Listener: &ListenerConfig{
			Type: "rtcp",
		},
		Forwarder: &ForwarderConfig{
			Nodes: nodes,
			Selector: &SelectorConfig{
				Strategy:    strategy,
				MaxFails:    1,
				FailTimeout: 30 * time.Second,
			},
		},
	}
}

// BuildUDPForwardService 构建 UDP 转发服务配置
func BuildUDPForwardService(name string, listenPort int, targets []string, strategy string) *ServiceConfig {
	nodes := make([]*ForwarderNode, 0)
	for i, target := range targets {
		nodes = append(nodes, &ForwarderNode{
			Name: fmt.Sprintf("target-%d", i),
			Addr: target,
		})
	}

	// 默认策略
	if strategy == "" {
		strategy = "round"
	}

	return &ServiceConfig{
		Name: name,
		Addr: fmt.Sprintf(":%d", listenPort),
		Handler: &HandlerConfig{
			Type: "rudp",
		},
		Listener: &ListenerConfig{
			Type: "rudp",
		},
		Forwarder: &ForwarderConfig{
			Nodes: nodes,
			Selector: &SelectorConfig{
				Strategy:    strategy,
				MaxFails:    1,
				FailTimeout: 30 * time.Second,
			},
		},
	}
}

// CreateLimiter 创建限流器 (幂等)
func (c *Client) CreateLimiter(limiter *LimiterConfig) error {
	path := fmt.Sprintf("/config/limiters/%s", limiter.Name)
	if c.exists(path) {
		logger.Debugf("限流器 %s 已存在，跳过创建", limiter.Name)
		return nil
	}

	resp, err := c.doRequest("POST", "/config/limiters", limiter)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建限流器失败: %s", string(body))
	}

	return nil
}

// DeleteLimiter 删除限流器 (幂等)
func (c *Client) DeleteLimiter(name string) error {
	path := fmt.Sprintf("/config/limiters/%s", name)
	if !c.exists(path) {
		logger.Debugf("限流器 %s 不存在，跳过删除", name)
		return nil
	}

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除限流器失败: %s", string(body))
	}

	return nil
}

// CreateObserver 创建观察器 (幂等)
func (c *Client) CreateObserver(observer *ObserverConfig) error {
	path := fmt.Sprintf("/config/observers/%s", observer.Name)
	if c.exists(path) {
		logger.Debugf("观察器 %s 已存在，跳过创建", observer.Name)
		return nil
	}

	resp, err := c.doRequest("POST", "/config/observers", observer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.ErrTunnelObserverCreateFailed
	}

	return nil
}

// DeleteObserver 删除观察器 (幂等)
func (c *Client) DeleteObserver(name string) error {
	path := fmt.Sprintf("/config/observers/%s", name)
	if !c.exists(path) {
		logger.Debugf("观察器 %s 不存在，跳过删除", name)
		return nil
	}

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除观察器失败: %s", string(body))
	}

	return nil
}

// CreateCLimiter 创建并发连接数限制器 (幂等)
func (c *Client) CreateCLimiter(climiter *CLimiterConfig) error {
	path := fmt.Sprintf("/config/climiters/%s", climiter.Name)
	if c.exists(path) {
		logger.Debugf("并发连接限制器 %s 已存在，跳过创建", climiter.Name)
		return nil
	}

	resp, err := c.doRequest("POST", "/config/climiters", climiter)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建并发连接限制器失败: %s", string(body))
	}

	return nil
}

// DeleteCLimiter 删除并发连接数限制器 (幂等)
func (c *Client) DeleteCLimiter(name string) error {
	path := fmt.Sprintf("/config/climiters/%s", name)
	if !c.exists(path) {
		logger.Debugf("并发连接限制器 %s 不存在，跳过删除", name)
		return nil
	}

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除并发连接限制器失败: %s", string(body))
	}

	return nil
}

// CreateRLimiter 创建请求速率限制器 (幂等)
func (c *Client) CreateRLimiter(rlimiter *RLimiterConfig) error {
	path := fmt.Sprintf("/config/rlimiters/%s", rlimiter.Name)
	if c.exists(path) {
		logger.Debugf("请求速率限制器 %s 已存在，跳过创建", rlimiter.Name)
		return nil
	}

	resp, err := c.doRequest("POST", "/config/rlimiters", rlimiter)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建请求速率限制器失败: %s", string(body))
	}

	return nil
}

// DeleteRLimiter 删除请求速率限制器 (幂等)
func (c *Client) DeleteRLimiter(name string) error {
	path := fmt.Sprintf("/config/rlimiters/%s", name)
	if !c.exists(path) {
		logger.Debugf("请求速率限制器 %s 不存在，跳过删除", name)
		return nil
	}

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除请求速率限制器失败: %s", string(body))
	}

	return nil
}

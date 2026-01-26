package dto

// ==================== 观察器相关 ====================

// ObserverEvent GOST 观察器上报的事件
type ObserverEvent struct {
	Kind    string          `json:"kind"`    // 事件类型: service, handler
	Service string          `json:"service"` // 服务名称
	Type    string          `json:"type"`    // 事件类型: status, stats
	Status  *ObserverStatus `json:"status"`  // 状态信息
	Stats   *ObserverStats  `json:"stats"`   // 统计信息
	Client  string          `json:"client"`  // 客户端标识（handler级别）
}

// ObserverStatus 服务状态
type ObserverStatus struct {
	State string `json:"state"` // 状态: running, ready, failed, closed
	Msg   string `json:"msg"`   // 状态消息
}

// ObserverStats 流量统计
type ObserverStats struct {
	TotalConns   int64 `json:"totalConns"`   // 总连接数
	CurrentConns int64 `json:"currentConns"` // 当前连接数
	InputBytes   int64 `json:"inputBytes"`   // 输入字节数
	OutputBytes  int64 `json:"outputBytes"`  // 输出字节数
	TotalErrs    int64 `json:"totalErrs"`    // 总错误数
}

// ObserverReportReq GOST 观察器上报请求
type ObserverReportReq struct {
	Events []ObserverEvent `json:"events"` // 事件列表
}

// ObserverReportResp 观察器上报响应
type ObserverReportResp struct {
	OK bool `json:"ok"` // 是否成功
}

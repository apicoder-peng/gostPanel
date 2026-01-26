package utils

import "gost-panel/internal/model"

// GostStateToForwardStatus 将 Gost 服务状态转换为转发规则状态
func GostStateToForwardStatus(state string) model.ForwardStatus {
	switch state {
	case "ready", "running":
		return model.ForwardStatusRunning
	case "failed":
		return model.ForwardStatusError
	default:
		return model.ForwardStatusStopped
	}
}

// GostStateToTunnelStatus 将 Gost 服务状态转换为隧道规则状态
func GostStateToTunnelStatus(state string, chainExists bool) model.TunnelStatus {
	switch state {
	case "ready", "running":
		if chainExists {
			return model.TunnelStatusRunning
		}
		return model.TunnelStatusError
	case "failed":
		return model.TunnelStatusError
	default:
		return model.TunnelStatusStopped
	}
}

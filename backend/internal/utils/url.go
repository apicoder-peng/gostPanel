package utils

import (
	"strings"
)

// ExtractHostPort 从 URL 中提取主机和端口
func ExtractHostPort(apiURL string) (string, string) {
	url := apiURL
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}

	// 移除路径
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	// 分离主机和端口
	if idx := strings.LastIndex(url, ":"); idx != -1 {
		return url[:idx], url[idx+1:]
	}

	return url, ""
}

// ExtractHost 从 URL 中提取主机（IP 或域名）
func ExtractHost(apiURL string) string {
	host, _ := ExtractHostPort(apiURL)
	return host
}

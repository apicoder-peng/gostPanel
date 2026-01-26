package router

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"gost-panel/pkg/logger"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var staticFiles embed.FS

func (r *Router) setupStatic(engine *gin.Engine) {
	// 获取 dist 子目录的内容
	// 使用 all:dist 确保包含以 _ 或 . 开头的文件 (如 _plugin-vue_export-helper.js)
	staticFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic(err)
	}

	// 使用 NoRoute 接管所有未匹配的请求
	engine.NoRoute(func(c *gin.Context) {
		pathStr := c.Request.URL.Path

		// 跳过 API 路由
		if strings.HasPrefix(pathStr, "/api") {
			return
		}

		// 规范化路径
		fileItem := strings.TrimPrefix(pathStr, "/")
		if fileItem == "" {
			fileItem = "index.html"
		}

		// 尝试从嵌入文件系统中读取文件
		data, err := fs.ReadFile(staticFS, fileItem)
		if err != nil {
			// 如果没找到文件，判断是否为前端页面路径（无后缀）
			// 如果是页面路径，返回 index.html 实现 SPA 路由
			if !strings.Contains(path.Base(fileItem), ".") {
				indexData, err := fs.ReadFile(staticFS, "index.html")
				if err == nil {
					c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
					return
				}
			}

			// 对于带后缀的资源文件，如果找不到，显式返回 404
			logger.Warnf("静态资源未找到: %s", fileItem)
			// 这里不写 c.Status(404)，让它走默认逻辑或返回
			return
		}

		// 获取扩展名并设置正确的 MIME 类型
		ext := path.Ext(fileItem)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			switch ext {
			case ".js":
				contentType = "application/javascript"
			case ".css":
				contentType = "text/css"
			case ".svg":
				contentType = "image/svg+xml"
			default:
				contentType = "application/octet-stream"
			}
		}

		c.Data(http.StatusOK, contentType, data)
	})
}

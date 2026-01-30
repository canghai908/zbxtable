package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:web
var distFS embed.FS

// GetDistFS 返回嵌入的前端文件系统
func GetDistFS() http.FileSystem {
	// 获取 dist 子目录
	sub, err := fs.Sub(distFS, "web")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// GetDistFSRoot 返回嵌入的前端文件系统（包含 dist 目录）
func GetDistFSRoot() embed.FS {
	return distFS
}

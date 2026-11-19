package middleware

import (
	"net/http"
	"os"
)

// NoListFileSystem 包装 http.FileSystem，禁止目录列举。
// 已知文件路径仍可直接访问（Logo / 背景图 / 报表文件名直链）。
type NoListFileSystem struct {
	FS http.FileSystem
}

// Open 打开文件；若目标是目录则返回不存在，避免默认 FileServer 列出目录内容。
func (n NoListFileSystem) Open(name string) (http.File, error) {
	f, err := n.FS.Open(name)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if stat.IsDir() {
		_ = f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

// NoListDir 返回禁用目录列举的本地目录文件系统。
func NoListDir(root string) http.FileSystem {
	return NoListFileSystem{FS: http.Dir(root)}
}

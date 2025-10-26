package storage

import (
	"fmt"
	"io"

	"github.com/vaaandark/PixelHub/internal/config"
)

// Provider 定义存储提供商接口
type Provider interface {
	// Upload 上传文件，返回存储 key 和访问 URL
	Upload(filename string, content io.Reader, contentType string) (storageKey string, url string, err error)

	// Delete 删除文件
	Delete(storageKey string) error

	// GetURL 获取文件的访问 URL
	GetURL(storageKey string) string
}

// NewProvider 根据配置创建存储提供商
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Storage.Provider {
	case "tencent-cos":
		return NewTencentCOSProvider(&cfg.Storage.TencentCOS)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Storage.Provider)
	}
}

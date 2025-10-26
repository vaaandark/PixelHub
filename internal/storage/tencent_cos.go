package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/vaaandark/PixelHub/internal/config"
)

type TencentCOSProvider struct {
	client *cos.Client
	config *config.TencentCOSConfig
}

func NewTencentCOSProvider(cfg *config.TencentCOSConfig) (*TencentCOSProvider, error) {
	u, err := url.Parse(cfg.BucketURL)
	if err != nil {
		return nil, fmt.Errorf("invalid bucket URL: %w", err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	return &TencentCOSProvider{
		client: client,
		config: cfg,
	}, nil
}

func (p *TencentCOSProvider) Upload(filename string, content io.Reader, contentType string) (string, string, error) {
	// 使用原始文件名作为 storage key
	storageKey := filename

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	_, err := p.client.Object.Put(context.Background(), storageKey, content, opt)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to COS: %w", err)
	}

	url := p.GetURL(storageKey)
	return storageKey, url, nil
}

func (p *TencentCOSProvider) Delete(storageKey string) error {
	_, err := p.client.Object.Delete(context.Background(), storageKey)
	if err != nil {
		return fmt.Errorf("failed to delete from COS: %w", err)
	}
	return nil
}

func (p *TencentCOSProvider) GetURL(storageKey string) string {
	// 如果配置了 CDN URL，使用 CDN
	if p.config.CDNURL != "" {
		return p.config.CDNURL + "/" + storageKey
	}
	// 否则使用 COS 直接访问 URL
	return p.config.BucketURL + "/" + storageKey
}

// GenerateStorageKey 生成存储 key（可以加上时间戳、UUID 等）
func GenerateStorageKey(originalFilename string) string {
	// 简单实现：保留文件扩展名
	_ = path.Ext(originalFilename)
	// 这里可以添加更复杂的逻辑，如加上时间戳、UUID 等
	return originalFilename
}

package llm

import "context"

// ImageAnalysisResult 图片分析结果
type ImageAnalysisResult struct {
	Description string   `json:"description"` // 图片描述
	Tags        []string `json:"tags"`        // 标签列表
}

// TagGenerator 标签生成器接口
type TagGenerator interface {
	// GenerateImageInfo 为图片生成描述和标签
	// imageURL: 图片的可访问 URL
	// prompt: 提示词（为空时使用默认提示词）
	// 返回: 分析结果（包含 description 和 tags）和错误
	GenerateImageInfo(ctx context.Context, imageURL string, prompt string) (*ImageAnalysisResult, error)
}

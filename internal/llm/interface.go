package llm

import "context"

// TagGenerator 标签生成器接口
type TagGenerator interface {
	// GenerateTags 为图片生成标签
	// imageURL: 图片的可访问 URL
	// prompt: 提示词（为空时使用默认提示词）
	// delimiter: 标签分隔符（用于解析 LLM 返回的文本）
	// 返回: 标签列表和错误
	GenerateTags(ctx context.Context, imageURL string, prompt string, delimiter string) ([]string, error)
}

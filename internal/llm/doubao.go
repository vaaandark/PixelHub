package llm

import (
	"context"
	"fmt"
	"time"

	ark "github.com/sashabaranov/go-openai"
	"github.com/vaaandark/PixelHub/internal/config"
)

// DoubaoGenerator Doubao 标签生成器
type DoubaoGenerator struct {
	client           *ark.Client
	model            string
	timeout          time.Duration
	defaultPrompt    string
	defaultDelimiter string
}

// NewDoubaoGenerator 创建 Doubao 标签生成器
func NewDoubaoGenerator(cfg *config.LLMConfig) (*DoubaoGenerator, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("LLM API key is required")
	}

	arkConfig := ark.DefaultConfig(cfg.APIKey)
	arkConfig.BaseURL = cfg.BaseURL
	client := ark.NewClientWithConfig(arkConfig)

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	defaultPrompt := cfg.DefaultPrompt
	if defaultPrompt == "" {
		defaultPrompt = DefaultPrompt
	}

	defaultDelimiter := cfg.DefaultDelimiter
	if defaultDelimiter == "" {
		defaultDelimiter = DefaultDelimiter
	}

	return &DoubaoGenerator{
		client:           client,
		model:            cfg.Model,
		timeout:          timeout,
		defaultPrompt:    defaultPrompt,
		defaultDelimiter: defaultDelimiter,
	}, nil
}

// GenerateTags 为图片生成标签
func (g *DoubaoGenerator) GenerateTags(ctx context.Context, imageURL string, prompt string, delimiter string) ([]string, error) {
	// 使用默认值
	if prompt == "" {
		prompt = g.defaultPrompt
	}
	if delimiter == "" {
		delimiter = g.defaultDelimiter
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// 构建请求
	req := ark.ChatCompletionRequest{
		Model: g.model,
		Messages: []ark.ChatCompletionMessage{
			{
				Role: ark.ChatMessageRoleUser,
				MultiContent: []ark.ChatMessagePart{
					{
						Type: ark.ChatMessagePartTypeImageURL,
						ImageURL: &ark.ChatMessageImageURL{
							URL: imageURL,
						},
					},
					{
						Type: ark.ChatMessagePartTypeText,
						Text: prompt,
					},
				},
			},
		},
	}

	// 调用 LLM
	resp, err := g.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	// 提取文本
	text := resp.Choices[0].Message.Content

	// 处理标签
	tags := ProcessTags(text, delimiter)

	return tags, nil
}

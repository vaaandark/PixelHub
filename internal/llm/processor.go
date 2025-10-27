package llm

import (
	"strings"
)

// ProcessTags 处理 LLM 返回的文本，提取标签列表
func ProcessTags(text string, delimiter string) []string {
	if text == "" {
		return []string{}
	}

	// 按分隔符分割
	parts := strings.Split(text, delimiter)

	// 去重map
	seen := make(map[string]bool)
	var tags []string

	for _, part := range parts {
		// 去除首尾空格
		tag := strings.TrimSpace(part)

		// 过滤空白标签
		if tag == "" {
			continue
		}

		// 去重
		if seen[tag] {
			continue
		}
		seen[tag] = true

		tags = append(tags, tag)
	}

	return tags
}

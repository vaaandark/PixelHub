package llm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParseImageAnalysisResult 解析 LLM 返回的 JSON 文本
func ParseImageAnalysisResult(text string) (*ImageAnalysisResult, error) {
	if text == "" {
		return nil, fmt.Errorf("empty response")
	}

	// 尝试提取 JSON 部分（LLM 可能在前后添加额外文本）
	jsonText := extractJSON(text)
	if jsonText == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	// 解析 JSON
	var result ImageAnalysisResult
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 清理和验证数据
	result.Description = strings.TrimSpace(result.Description)
	result.Tags = cleanTags(result.Tags)

	return &result, nil
}

// extractJSON 从文本中提取 JSON 部分
func extractJSON(text string) string {
	// 尝试匹配 JSON 对象（花括号包裹）
	re := regexp.MustCompile(`\{[^}]*"description"[^}]*"tags"[^}]*\}`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}

	// 如果找不到，尝试直接查找完整的 JSON
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 && start < end {
		return text[start : end+1]
	}

	return ""
}

// cleanTags 清理标签列表
func cleanTags(tags []string) []string {
	seen := make(map[string]bool)
	var cleaned []string

	for _, tag := range tags {
		// 去除首尾空格
		tag = strings.TrimSpace(tag)

		// 过滤空白标签
		if tag == "" {
			continue
		}

		// 去重
		if seen[tag] {
			continue
		}
		seen[tag] = true

		cleaned = append(cleaned, tag)
	}

	return cleaned
}

package llm

const (
	// DefaultPrompt 默认提示词
	DefaultPrompt = "请分析这张图片，为它生成一个简洁的描述（20-50字）和 5-10 个描述性标签。标签应准确、简洁，涵盖：主题、场景、风格、色彩、情感等方面。"

	// JSONFormatPrompt JSON 格式约束提示词
	JSONFormatPrompt = `

请严格按照以下 JSON 格式输出，不要包含任何其他内容：
{
  "description": "图片的简洁描述",
  "tags": ["标签1", "标签2", "标签3"]
}`
)

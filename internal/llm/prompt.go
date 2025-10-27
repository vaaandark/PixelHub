package llm

// DefaultPrompt 默认的中文提示词
const DefaultPrompt = `请分析这张图片，生成 5-10 个中文描述性标签，用半角逗号分隔。
标签应准确、简洁，涵盖：主题、场景、风格、色彩、情感等方面。
请只返回标签列表，不要包含其他内容。`

// DefaultDelimiter 默认分隔符
const DefaultDelimiter = ","

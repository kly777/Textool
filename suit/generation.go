package suit

import (
	"context"
	"fmt"
	"log"

	"regexp"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// processChunk 处理单个文本块
func (tp *TextProcessor) processChunk(index int, chunk, context string) (string, string) {
	prompt := fmt.Sprintf(`请严格按照以下规范将文本转换为结构化Markdown：
【核心原则】
1. 原始内容完整保留：不得翻译、删减或改写任何内容
2. 逻辑结构映射：
   - 自动推断章节层级（#主标题/##二级标题/###三级标题）
   - 使用列表/代码块等元素保持原有排版逻辑
3. 关键信息强化：
   - 专业术语和核心概念用**加粗**
4. 上下文衔接：
   - 严格依据提供的上下文线索（见下文）保持文档连贯性,但不用在结果中输出上下文
5. 输出限制:
	 - 只应该输出当前文本格式化后的内容，而没有上下文线索


【输入内容】
上下文线索：%s
---
当前文本：%s`, context, chunk)

	startTime := time.Now()
	log.Println("开始处理文本块", "index", index, "chunk_length", len(chunk))

	var mdContent string
	for attempt := range 3 {
		log.Println("API调用尝试", "attempt", attempt+1)

		md, err := tp.generateMD(prompt)
		if err == nil {
			mdContent = md
			log.Println("处理成功",
				"attempts", attempt+1,
				"duration", time.Since(startTime))
			break
		}

		log.Println("处理失败",
			"attempt", attempt+1,
			"error", err,
			"retry_in", attempt*2)
		time.Sleep(time.Duration(attempt*2) * time.Second)
	}

	// 本地生成摘要
	summary := extractSummary(mdContent)
	return mdContent, summary
}

// generateMD 调用API生成Markdown
func (tp *TextProcessor) generateMD(prompt string) (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(tp.apiKey),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
	)

	chatCompletion, err := client.Chat.Completions.New(
		context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},

			Model: "qwen-long",
		},
	)
	if err != nil {
		panic(err.Error())
	}

	response := chatCompletion.Choices[0].Message.Content
	return response, nil
}

// joinResults 将多个result对象的md字段合并为一个字符串。
// 参数:
//
//	results ([]result) - 包含多个result对象的切片，每个result对象应包含md字段
//
// 返回值:
//
//	string - 所有result对象的md字段按顺序拼接，每段之间用两个换行符分隔
func joinResults(results []result) string {
	// 收集所有result对象的md字段到output切片
	var output []string
	for _, res := range results {
		output = append(output, res.md)
	}

	// 使用双换行符将各md字段拼接为单个字符串返回
	return strings.Join(output, "\n\n")
}

// TextProcessor的postProcess方法用于规范化Markdown文本中的换行符。
// 将连续三个及以上换行符替换为标准的两个换行符，确保输出格式一致性。
//
// 参数:
//
//	mdText string - 需要处理的原始Markdown文本
//
// 返回值:
//
//	string - 处理后的规范化文本，其中连续换行被压缩为双换行
func postProcess(mdText string) string {
	re := regexp.MustCompile(`\n{3,}`)
	return re.ReplaceAllString(mdText, "\n\n")
}

// extractSummary 从Markdown中提取摘要
func extractSummary(md string) string {
	re := regexp.MustCompile(`(?m)^#+\s*(.+)$`)
	matches := re.FindAllStringSubmatch(md, 3)
	if len(matches) == 0 {
		return "无关键标题"
	}

	var titles []string
	for _, m := range matches {
		if len(m) > 1 {
			titles = append(titles, m[1])
		}
	}
	return strings.Join(titles, "; ")
}

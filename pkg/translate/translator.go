package translate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	systemPrompt = `您作为本体论领域的专业翻译,请将以下本体论相关内容翻译为中文，要求：
0. 主要是翻译，其次是可视化的解释，你只用处理给你的部分，不要延申其他的东西,仅翻译待翻译行，不要增加其他内容
1. 保持原始Markdown结构：
    • 标题层级（#→######）完全保留
    • 列表项保持缩进和项目符号

2. 术语处理：
    • 核心术语首现使用【中文(英文)】格式（例：本体论(Ontology)）
    • 后续统一使用中文表述

3. 人名不翻译（如：John Smith）

4. 技术术语保持准确性

5. 错误修正：
    • 自动修复常见错误：
        - "corners tone" → "cornerstone"
        - "inter-relation$" → "inter-relations"
    • 合并错误换行的段落
    • 删除扫描残留字符（如■、▢）
6. 输入可能很短，不要自我发挥，只翻译，不增加内容
7. 仅输出待翻译行翻译后内容，不要输出原文和你的思考
`
)

type Translator struct {
	client *openai.Client
}

func NewTranslator(apiKey string) *Translator {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)
	return &Translator{client: &client}
}

func (t *Translator) ProcessFile(inputPath, outputPath string) error {
	lines, err := readFileLines(inputPath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	results := make([]string, len(lines))
	var wg sync.WaitGroup
	var processed int64                            // 进度计数器
	sem := make(chan struct{}, runtime.NumCPU()*2) // 并发控制

	// 启动进度监控goroutine
	var num int = 0
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				progress := float64(atomic.LoadInt64(&processed)) / float64(num) * 100
				fmt.Printf("\r处理进度: %.1f%% (%d/%d)", progress, atomic.LoadInt64(&processed), num)
			case <-sem: // 所有任务完成
				if processed == int64(num) {
					fmt.Printf("\r处理进度: 100%% (全部完成)  \n")
					return
				}
			}
		}
	}()

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			results[i] = ""
			continue
		}
		num++

		wg.Add(1)
		sem <- struct{}{} // 资源不足时阻塞

		go func(idx int, line string) {
			defer func() {
				atomic.AddInt64(&processed, 1) // 更新进度
				wg.Done()
				<-sem
			}()

			result, err := t.translateLine(line, idx, lines)
			if err != nil {
				log.Printf("行 %d 翻译失败: %v", idx, err)
				results[idx] = fmt.Sprintf("【失败】%s", line)
			} else {
				results[idx] = result
			}
		}(i, line)
	}

	wg.Wait()
	close(sem)
	return writeFile(outputPath, results)
}

func (t *Translator) translateLine(line string, idx int, allLines []string) (string, error) {
	const contextLines = 2
	start := max(0, idx-contextLines)
	end := min(len(allLines), idx+contextLines+1)

	contextWindow := ""
	for i := start; i < end; i++ {
		if i != idx {
			contextWindow += allLines[i] + "\n"
		}
	}

	const maxRetries = 5
	retryDelays := []time.Duration{2, 4, 8, 16, 32} // 指数退避时间（秒）

	for attempt := range maxRetries {
		processor := &StreamProcessor{}
		response, err := t.client.Chat.Completions.New(
			context.TODO(),
			openai.ChatCompletionNewParams{
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(fmt.Sprintf("%s\n\n上下文片段：\n%s\n\n待翻译行：%s", systemPrompt, contextWindow, line)),
				},
				Model:     "deepseek-reasoner",
				MaxTokens: openai.Int(16000),
			},
		)
		if err != nil {
			return "", fmt.Errorf("API请求失败: %w", err)
		}

		for _, choice := range response.Choices {
			if choice.Message.Content != "" {
				processor.feed(choice.Message.Content)
			}
		}

		fullContent := processor.getContent()
		if fullContent != "" {
			return line + "\n" + fullContent, nil
		}

		time.Sleep(retryDelays[attempt] * time.Second)
	}

	return "", fmt.Errorf("超过最大重试次数")
}

func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeFile(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	fmt.Println("翻译结果已保存到文件：", path)
	return nil
}

type StreamProcessor struct {
	buffer bytes.Buffer
}

func (p *StreamProcessor) feed(content string) {
	p.buffer.WriteString(content)
}

func (p *StreamProcessor) getContent() string {
	return p.buffer.String()
}

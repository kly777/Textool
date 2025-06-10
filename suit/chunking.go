// chunking.go
package suit

import (
	"strings"
)

func (tp *TextProcessor) splitText(text string) []string {
	lines := strings.Split(text, "\n")
	chunks := makeChunks(lines, tp.chunkSize)
	return chunks
}

func makeChunks(lines []string, chunkSize int) []string {
	var chunks []string
	currentChunk := ""
	currentLen := 0

	for _, line := range lines {
		lineLen := len(line)

		// 空块直接接受当前行
		if currentLen == 0 {
			currentChunk = line
			currentLen = lineLen
			continue
		}

		// 计算添加换行符后的总长度
		totalLenWithNewLine := currentLen + 1 + lineLen

		// 超过阈值则保存当前块
		if totalLenWithNewLine > chunkSize {
			chunks = append(chunks, currentChunk)
			currentChunk = line
			currentLen = lineLen
		} else {
			// 否则继续累加
			currentChunk += "\n" + line
			currentLen = totalLenWithNewLine
		}
	}

	// 添加最后的剩余内容
	if currentLen > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

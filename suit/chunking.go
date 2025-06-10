// chunking.go
package suit

import (
	"strings"
)

// splitText 将输入文本按换行符分割成行，并将行切片分割为指定大小的块。
// 参数:
//
//	text: 需要被分割的原始文本字符串
//
// 返回值:
//
//	分割后的字符串切片，每个元素为一个块的内容
func splitText(text string, chunkSize int) (chunks []string) {
	// 按换行符将文本分割为行切片
	lines := strings.Split(text, "\n")
	// 将行切片分割为指定大小的块
	chunks = makeChunks(lines, chunkSize)
	return chunks
}

// makeChunks 将给定的文本行按最大块长度分割成多个块。
// 参数:
//
//	lines: 需要被分割的文本行切片。
//	chunkSize: 每个块的最大长度（必须大于0）。
//
// 返回值:
//
//	分割后的字符串切片，每个元素代表一个块。块内行通过换行符连接，
//	且每个块的总长度不超过chunkSize。当单行长度超过chunkSize时，
//	该行将单独作为一块。
func makeChunks(lines []string, chunkSize int) []string {
	var chunks []string
	currentChunk := ""
	currentLen := 0

	// 遍历所有文本行进行分块处理
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

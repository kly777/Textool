package div

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type headingInfo struct {
	level   int
	title   string
	dirPath string
	content []string
}

func sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[\\/*?:"<>|]`)
	return reg.ReplaceAllString(name, "_")
}

func processMDFile(inputFile, outputDir string) error {
	data, err := readFile(inputFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	lines := strings.Split(string(data), "\n")

	var stack []headingInfo
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\n")
		if err := processLine(line, &stack, outputDir); err != nil {
			return err
		}
	}

	// 处理栈中剩余元素
	for len(stack) > 0 {
		popped := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if err := writeHeadingFile(popped); err != nil {
			return err
		}
	}

	return nil
}

func readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func processLine(line string, stack *[]headingInfo, outputDir string) error {
	headingRe := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	matches := headingRe.FindStringSubmatch(line)
	if len(matches) == 3 {
		level := len(matches[1])
		rawTitle := strings.TrimSpace(matches[2])
		title := sanitizeFilename(rawTitle)

		// 弹出层级小于当前层级的元素
		for len(*stack) > 0 && (*stack)[len(*stack)-1].level >= level {
			popped := (*stack)[len(*stack)-1]
			*stack = (*stack)[:len(*stack)-1]

			if err := writeHeadingFile(popped); err != nil {
				return err
			}
		}

		// 确定父目录
		var parentDir string
		if len(*stack) == 0 {
			parentDir = outputDir
		} else {
			parentDir = (*stack)[len(*stack)-1].dirPath
		}

		// 创建当前目录
		currentDir := filepath.Join(parentDir, title)
		if err := createDirectory(currentDir); err != nil {
			return fmt.Errorf("创建目录失败: %v", err)
		}

		// 压入堆栈
		*stack = append(*stack, headingInfo{
			level:   level,
			title:   title,
			dirPath: currentDir,
			content: []string{line},
		})
	} else {
		if len(*stack) > 0 {
			(*stack)[len(*stack)-1].content = append((*stack)[len(*stack)-1].content, line)
		}
	}

	return nil
}

func createDirectory(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

func writeHeadingFile(h headingInfo) error {
	mdPath := filepath.Join(h.dirPath, h.title+".md")
	content := strings.Join(h.content, "\n")

	if err := os.WriteFile(mdPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	return nil
}

func main() {
	if err := processMDFile("formalOntology.md", "dived"); err != nil {
		fmt.Printf("处理失败: %v\n", err)
		os.Exit(1)
	}
}

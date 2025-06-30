package combine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Combine(rootPath, outputPath string, excludePrefixes, excludeSuffixes []string) {
	fmt.Println("Combining...")
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil && !os.IsExist(err) {
		panic(err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()

	results := []string{}
	num := 0

	err = filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				fileName := path
				for i := range excludePrefixes {
					if strings.HasPrefix(fileName, excludePrefixes[i]) {
						return nil
					}
				}
				for i := range excludeSuffixes {
					if strings.HasSuffix(fileName, excludeSuffixes[i]) {
						return nil
					}
				}

				ext := filepath.Ext(fileName)
				switch ext {
				case ".txt", ".md", ".go", ".rs", ".mod", ".sum",
					".yaml", ".yml", ".json", ".xml", ".html",
					".css", ".js", ".ts", ".java", ".py", ".rb", ".php":
					// 允许的文本文件扩展名
				default:
					return nil // 跳过非文本文件
				}

				fileContentByte, err := os.ReadFile(path)
				if err != nil {
					panic(err)
				}
				fileContent := string(fileContentByte)
				result := "// " + fileName + "\n" + fileContent
				results = append(results, result)
				num += 1
				fmt.Println("已处理文件：", fileName)
			}
			return nil
		})
	if err != nil {
		panic(err)
	}
	result := strings.Join(results, "\n------\n")
	if _, err := file.Write([]byte(result)); err != nil {
		panic(fmt.Errorf("写入合并文件失败: %w", err))
	}
	fmt.Println("处理了", num, "个文件")
	fmt.Println("Done")
}

package combine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Combine(rootPath, outputPath string, blockPrefix, blockSuffix []string) {
	fmt.Println("Combining...")
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil && !os.IsExist(err) {
		panic(err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	results := []string{}
	num := 0

	err = filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				fileName := path
				for i := range blockPrefix {
					if strings.HasPrefix(fileName, blockPrefix[i]) {
						return nil
					}
				}
				for i := range blockSuffix {
					if strings.HasSuffix(fileName, blockSuffix[i]) {
						return nil
					}
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
	file.Write([]byte(result))
	fmt.Println("处理了", num, "个文件")
	fmt.Println("Done")
}

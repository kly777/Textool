package suit

import (
	"os"
)

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func writeFile(path, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type config struct {
	Ds_api_key string `json:"ds_api_key"`
	Bl_api_key string `json:"bl_api_key"`
}

func GetConfig() (*config, error) {
	configData, err := os.ReadFile("./config.json")
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析JSON
	var cfg config
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	return &cfg, nil
}

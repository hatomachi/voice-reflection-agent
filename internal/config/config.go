package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	SaveDir       string `json:"saveDir"`       // 会議保存先ディレクトリ（デフォルト: ./data/meetings）
	ClaudeCommand string `json:"claudeCommand"` // Claude CLI コマンド名（デフォルト: claude）
	ClaudeArgs    string `json:"claudeArgs"`    // 追加引数（デフォルト: -p）
}

var (
	defaultConfig = Config{
		SaveDir:       filepath.Join(".", "data", "meetings"),
		ClaudeCommand: "claude",
		ClaudeArgs:    "-p",
	}
	currentConfig Config
	configLock    sync.RWMutex
	configPath    = filepath.Join(".", "data", "config.json")
)

func InitConfig() Config {
	configLock.Lock()
	defer configLock.Unlock()

	currentConfig = defaultConfig

	data, err := os.ReadFile(configPath)
	if err == nil {
		_ = json.Unmarshal(data, &currentConfig)
	}

	if currentConfig.SaveDir == "" {
		currentConfig.SaveDir = defaultConfig.SaveDir
	}
	if currentConfig.ClaudeCommand == "" {
		currentConfig.ClaudeCommand = defaultConfig.ClaudeCommand
	}
	if currentConfig.ClaudeArgs == "" {
		currentConfig.ClaudeArgs = defaultConfig.ClaudeArgs
	}

	return currentConfig
}

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return currentConfig
}

func UpdateConfig(cfg Config) error {
	configLock.Lock()
	defer configLock.Unlock()

	if cfg.SaveDir != "" {
		currentConfig.SaveDir = cfg.SaveDir
	}
	if cfg.ClaudeCommand != "" {
		currentConfig.ClaudeCommand = cfg.ClaudeCommand
	}
	if cfg.ClaudeArgs != "" {
		currentConfig.ClaudeArgs = cfg.ClaudeArgs
	}

	_ = os.MkdirAll(filepath.Dir(configPath), 0755)
	data, err := json.MarshalIndent(currentConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

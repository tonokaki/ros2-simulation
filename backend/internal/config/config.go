package config

import (
	"fmt"
	"os"
)

// Config はアプリケーション全体の設定を保持する
type Config struct {
	Port        string
	DatabaseURL string
	OpenAIKey   string
	RosBridgeURL string
	UseMockLLM  bool
	UseMockROS  bool
}

// Load は環境変数から設定を読み込む
func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://robotasker:robotasker@localhost:5432/robotasker?sslmode=disable"),
		OpenAIKey:    getEnv("OPENAI_API_KEY", ""),
		RosBridgeURL: getEnv("ROSBRIDGE_URL", "ws://localhost:9090"),
		UseMockLLM:   getEnv("USE_MOCK_LLM", "true") == "true",
		UseMockROS:   getEnv("USE_MOCK_ROS", "true") == "true",
	}
}

// DSN はPostgreSQL接続文字列を返す
func (c *Config) DSN() string {
	return c.DatabaseURL
}

func (c *Config) ListenAddr() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

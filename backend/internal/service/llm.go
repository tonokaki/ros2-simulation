package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/takaki0/robotasker-backend/internal/model"
)

// LLMService は自然言語をタスクに変換するインターフェース
type LLMService interface {
	ParseCommand(ctx context.Context, text string) (*model.ParsedTask, error)
}

// MockLLMService は正規表現ベースのモック実装
type MockLLMService struct{}

func NewMockLLMService() *MockLLMService {
	return &MockLLMService{}
}

// ParseCommand は自然言語テキストを構造化タスクに変換する（モック実装）
func (s *MockLLMService) ParseCommand(ctx context.Context, text string) (*model.ParsedTask, error) {
	parsed := &model.ParsedTask{
		Action:     model.TaskActionGoto,
		Priority:   "normal",
		Confidence: 0.8,
	}

	// アクション判定
	lower := strings.ToLower(text)
	switch {
	case containsAny(lower, "届け", "配達", "deliver", "持って", "運んで"):
		parsed.Action = model.TaskActionDeliver
	case containsAny(lower, "巡回", "patrol", "パトロール", "見回り"):
		parsed.Action = model.TaskActionPatrol
	case containsAny(lower, "戻", "帰", "return", "充電"):
		parsed.Action = model.TaskActionReturn
		parsed.TargetLocation = "充電ステーション"
		parsed.Confidence = 0.9
	default:
		parsed.Action = model.TaskActionGoto
	}

	// 場所抽出
	locationPatterns := []struct {
		pattern *regexp.Regexp
		name    string
	}{
		{regexp.MustCompile(`会議室\s*[AaBb]`), ""},
		{regexp.MustCompile(`受付`), "受付"},
		{regexp.MustCompile(`休憩室`), "休憩室"},
		{regexp.MustCompile(`倉庫`), "倉庫"},
		{regexp.MustCompile(`エントランス`), "エントランス"},
		{regexp.MustCompile(`充電ステーション`), "充電ステーション"},
	}

	for _, lp := range locationPatterns {
		match := lp.pattern.FindString(text)
		if match != "" {
			if lp.name != "" {
				parsed.TargetLocation = lp.name
			} else {
				// 会議室A/B の正規化
				normalized := strings.ReplaceAll(match, " ", "")
				normalized = strings.ReplaceAll(normalized, "a", "A")
				normalized = strings.ReplaceAll(normalized, "b", "B")
				parsed.TargetLocation = normalized
			}
			parsed.Confidence = 0.9
			break
		}
	}

	// アイテム抽出（配達系のとき）
	if parsed.Action == model.TaskActionDeliver {
		itemPatterns := regexp.MustCompile(`(資料|荷物|書類|コーヒー|お茶|弁当|パッケージ)`)
		if match := itemPatterns.FindString(text); match != "" {
			parsed.Item = match
		}
	}

	// 優先度判定
	if containsAny(lower, "急ぎ", "至急", "urgent", "緊急", "すぐ") {
		parsed.Priority = "high"
	}

	// 場所が取れなかった場合は信頼度を下げる
	if parsed.TargetLocation == "" && parsed.Action != model.TaskActionPatrol {
		parsed.Confidence = 0.4
	}

	return parsed, nil
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

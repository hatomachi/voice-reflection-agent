package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"voice-reflection-agent/internal/config"
)

// Claude CLI を実行する
func ExecuteClaude(ctx context.Context, prompt string) (string, error) {
	cfg := config.GetConfig()
	cmdName := cfg.ClaudeCommand

	claudePath, err := exec.LookPath(cmdName)
	if err != nil {
		return "", fmt.Errorf("コマンド '%s' が見つかりません。Claude CLI（Claude Code 等）がインストールされ、PATHが通っているか確認してください", cmdName)
	}

	// 引数の組み立て
	var args []string
	if cfg.ClaudeArgs != "" {
		args = strings.Fields(cfg.ClaudeArgs)
	}

	cmdArgs := append([]string{}, args...)
	cmdArgs = append(cmdArgs, prompt)

	cmd := exec.CommandContext(ctx, claudePath, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(prompt)

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr != "" {
			return "", fmt.Errorf("%w: %s", err, errStr)
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

// セルフリフレクションを生成する
func GenerateReflection(ctx context.Context, rawText string, vaultPath string) (string, error) {
	now := time.Now()
	dateWithDay := now.Format("2006-01-02")
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	dateWithDay = fmt.Sprintf("%s (%s)", dateWithDay, weekdays[now.Weekday()])

	// Vaultコンテキストの取得
	vPath := resolveVaultPath(vaultPath)
	roleDefs := loadVaultFile(vPath, "00_役割定義.md", "（00_役割定義.md は未検出です。一般的な7つの習慣基準で評価してください）")
	dashboard := loadVaultFile(vPath, "00_Dashboard.md", "（00_Dashboard.md は未検出です）")

	prompt := ReflectionPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{ROLE_DEFINITIONS}}", roleDefs)
	prompt = strings.ReplaceAll(prompt, "{{DASHBOARD_CONTEXT}}", dashboard)
	prompt = strings.ReplaceAll(prompt, "{{RECENT_REFLECTIONS}}", "（最新のセルフリフレクション）")
	prompt = strings.ReplaceAll(prompt, "{{RAW_TRANSCRIPTION}}", rawText)
	prompt = strings.ReplaceAll(prompt, "{{DATE_WITH_DAY}}", dateWithDay)

	return ExecuteClaude(ctx, prompt)
}

// 会議サマリーを生成する
func GenerateMeetingSummary(ctx context.Context, meetingID, duration, content string, screenshotCount int) (string, error) {
	prompt := MeetingSummaryPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{MEETING_ID}}", meetingID)
	prompt = strings.ReplaceAll(prompt, "{{DURATION}}", duration)
	prompt = strings.ReplaceAll(prompt, "{{SCREENSHOT_COUNT}}", fmt.Sprintf("%d", screenshotCount))
	prompt = strings.ReplaceAll(prompt, "{{CONTENT}}", content)

	return ExecuteClaude(ctx, prompt)
}

// personal-vault の自動探索
func resolveVaultPath(specified string) string {
	if specified != "" {
		if fi, err := os.Stat(specified); err == nil && fi.IsDir() {
			return specified
		}
	}

	candidates := []string{
		"../personal-vault",
		"../../personal-vault",
		filepath.Join(os.Getenv("HOME"), "work", "personal-vault"),
		filepath.Join(os.Getenv("USERPROFILE"), "work", "personal-vault"),
		filepath.Join(os.Getenv("USERPROFILE"), "Documents", "personal-vault"),
	}

	for _, c := range candidates {
		if c == "" {
			continue
		}
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}

	return ""
}

func loadVaultFile(vaultPath, fileName, fallback string) string {
	if vaultPath == "" {
		return fallback
	}
	target := filepath.Join(vaultPath, fileName)
	data, err := os.ReadFile(target)
	if err != nil {
		return fallback
	}
	return string(data)
}

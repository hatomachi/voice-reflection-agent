package ai

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

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

// 会議サマリーを生成する
func GenerateMeetingSummary(ctx context.Context, meetingID, duration, content string, screenshotCount int) (string, error) {
	prompt := MeetingSummaryPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{MEETING_ID}}", meetingID)
	prompt = strings.ReplaceAll(prompt, "{{DURATION}}", duration)
	prompt = strings.ReplaceAll(prompt, "{{SCREENSHOT_COUNT}}", fmt.Sprintf("%d", screenshotCount))
	prompt = strings.ReplaceAll(prompt, "{{CONTENT}}", content)

	return ExecuteClaude(ctx, prompt)
}

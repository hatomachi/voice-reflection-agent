package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"voice-reflection-agent/internal/ai"
	"voice-reflection-agent/internal/config"
)

type APIHandler struct{}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", h.handleHealth)
	mux.HandleFunc("/api/ai/claude", h.handleClaude)
	mux.HandleFunc("/api/ai/meeting-summary", h.handleMeetingSummary)
	mux.HandleFunc("/api/save/meeting", h.handleSaveMeeting)
	mux.HandleFunc("/api/open-folder", h.handleOpenFolder)
	mux.HandleFunc("/api/config", h.handleConfig)
}

// Health レスポンス
type HealthResponse struct {
	Status         string `json:"status"`
	Native         bool   `json:"native"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	ClaudeFound    bool   `json:"claudeFound"`
	ClaudePath     string `json:"claudePath,omitempty"`
	VaultFound     bool   `json:"vaultFound"`
	VaultPath      string `json:"vaultPath,omitempty"`
	DefaultSaveDir string `json:"defaultSaveDir"`
}

func (h *APIHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := config.GetConfig()
	claudePath, err := exec.LookPath(cfg.ClaudeCommand)

	vaultCandidates := []string{
		"../personal-vault",
		"../../personal-vault",
		filepath.Join(os.Getenv("HOME"), "work", "personal-vault"),
		filepath.Join(os.Getenv("USERPROFILE"), "work", "personal-vault"),
	}
	var detectedVault string
	for _, v := range vaultCandidates {
		if v != "" {
			if fi, err := os.Stat(v); err == nil && fi.IsDir() {
				detectedVault, _ = filepath.Abs(v)
				break
			}
		}
	}

	resp := HealthResponse{
		Status:         "ok",
		Native:         true,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		ClaudeFound:    err == nil,
		ClaudePath:     claudePath,
		VaultFound:     detectedVault != "",
		VaultPath:      detectedVault,
		DefaultSaveDir: cfg.SaveDir,
	}

	writeJSON(w, http.StatusOK, resp)
}

// Claude 実行リクエスト/レスポンス
type ClaudeRequest struct {
	Prompt  string `json:"prompt"`
	System  string `json:"system,omitempty"`
	Command string `json:"command,omitempty"`
	Args    string `json:"args,omitempty"`
}

type ClaudeResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
	Command string `json:"command"`
}

func (h *APIHandler) handleClaude(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClaudeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ClaudeResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, ClaudeResponse{
			Success: false,
			Error:   "Prompt is required",
		})
		return
	}

	cfg := config.GetConfig()
	cmdName := cfg.ClaudeCommand
	if req.Command != "" {
		cmdName = req.Command
	}

	// コマンドの存在確認
	claudePath, err := exec.LookPath(cmdName)
	if err != nil {
		writeJSON(w, http.StatusOK, ClaudeResponse{
			Success: false,
			Command: cmdName,
			Error:   fmt.Sprintf("コマンド '%s' が見つかりませんでした。Claude CLI（例: claude）がインストールされ、PATHが通っているか確認してください。", cmdName),
		})
		return
	}

	// タイムアウトを5分に設定
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 引数の組み立て
	// デフォルトでは claude -p "<prompt>" または stdin 受付
	argsStr := cfg.ClaudeArgs
	if req.Args != "" {
		argsStr = req.Args
	}

	var args []string
	if argsStr != "" {
		args = strings.Fields(argsStr)
	}

	// 全文プロンプトの合成（system がある場合は先頭に配置）
	fullPrompt := req.Prompt
	if req.System != "" {
		fullPrompt = fmt.Sprintf("System Instructions:\n%s\n\nUser Request:\n%s", req.System, req.Prompt)
	}

	// claude -p <prompt> を第一優先で組み立てる
	// もし args に -p があればその後ろに渡す
	cmdArgs := append([]string{}, args...)
	cmdArgs = append(cmdArgs, fullPrompt)

	cmd := exec.CommandContext(ctx, claudePath, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 標準入力にも渡しておく（stdin対応CLI向け）
	cmd.Stdin = strings.NewReader(fullPrompt)

	runErr := cmd.Run()
	outputStr := strings.TrimSpace(stdout.String())
	errStr := strings.TrimSpace(stderr.String())

	if runErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			writeJSON(w, http.StatusOK, ClaudeResponse{
				Success: false,
				Command: cmdName,
				Error:   "Claudeの実行がタイムアウト（5分）しました。",
			})
			return
		}

		errMsg := runErr.Error()
		if errStr != "" {
			errMsg = fmt.Sprintf("%s\nStderr: %s", errMsg, errStr)
		}
		writeJSON(w, http.StatusOK, ClaudeResponse{
			Success: false,
			Command: cmdName,
			Error:   "Claude実行エラー: " + errMsg,
		})
		return
	}

	writeJSON(w, http.StatusOK, ClaudeResponse{
		Success: true,
		Result:  outputStr,
		Command: cmdName,
	})
}

// 会議保存リクエスト/レスポンス
type SaveImageItem struct {
	Name   string `json:"name"`
	Base64 string `json:"base64"`
}

type SaveMeetingRequest struct {
	DirName       string          `json:"dirName"`
	FolderPath    string          `json:"folderPath,omitempty"`
	ReadmeContent string          `json:"readmeContent"`
	AudioBase64   string          `json:"audioBase64,omitempty"`
	AudioFileName string          `json:"audioFileName,omitempty"`
	Images        []SaveImageItem `json:"images,omitempty"`
}

type SaveMeetingResponse struct {
	Success   bool   `json:"success"`
	SavedPath string `json:"savedPath"`
	FileCount int    `json:"fileCount"`
	Error     string `json:"error,omitempty"`
}

func (h *APIHandler) handleSaveMeeting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SaveMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, SaveMeetingResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		})
		return
	}

	cfg := config.GetConfig()
	baseDir := cfg.SaveDir
	if req.FolderPath != "" {
		baseDir = req.FolderPath
	}

	dirName := req.DirName
	if dirName == "" {
		dirName = time.Now().Format("2006-01-02_150405")
	}

	targetDir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, SaveMeetingResponse{
			Success: false,
			Error:   "フォルダ作成失敗: " + err.Error(),
		})
		return
	}

	fileCount := 0

	// 1. README.md の保存
	if req.ReadmeContent != "" {
		readmePath := filepath.Join(targetDir, "README.md")
		if err := os.WriteFile(readmePath, []byte(req.ReadmeContent), 0644); err != nil {
			writeJSON(w, http.StatusInternalServerError, SaveMeetingResponse{
				Success: false,
				Error:   "README.md 保存失敗: " + err.Error(),
			})
			return
		}
		fileCount++
	}

	// 2. 音声ファイル（MP3）の保存
	if req.AudioBase64 != "" {
		audioData, err := decodeBase64Data(req.AudioBase64)
		if err == nil && len(audioData) > 0 {
			audioName := req.AudioFileName
			if audioName == "" {
				audioName = "meeting_audio.mp3"
			}
			audioPath := filepath.Join(targetDir, audioName)
			if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
				writeJSON(w, http.StatusInternalServerError, SaveMeetingResponse{
					Success: false,
					Error:   "音声ファイル保存失敗: " + err.Error(),
				})
				return
			}
			fileCount++
		}
	}

	// 3. スクリーンショット画像群の保存
	if len(req.Images) > 0 {
		imagesDir := filepath.Join(targetDir, "images")
		_ = os.MkdirAll(imagesDir, 0755)

		for _, img := range req.Images {
			if img.Base64 == "" || img.Name == "" {
				continue
			}
			imgData, err := decodeBase64Data(img.Base64)
			if err != nil || len(imgData) == 0 {
				continue
			}
			imgPath := filepath.Join(imagesDir, filepath.Base(img.Name))
			if err := os.WriteFile(imgPath, imgData, 0644); err == nil {
				fileCount++
			}
		}
	}

	absPath, _ := filepath.Abs(targetDir)
	writeJSON(w, http.StatusOK, SaveMeetingResponse{
		Success:   true,
		SavedPath: absPath,
		FileCount: fileCount,
	})
}

// フォルダを開くリクエスト
type OpenFolderRequest struct {
	Path string `json:"path"`
}

func (h *APIHandler) handleOpenFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OpenFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	targetPath := strings.TrimSpace(req.Path)
	if targetPath == "" {
		cfg := config.GetConfig()
		targetPath = cfg.SaveDir
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = targetPath
	}

	// OSごとのフォルダオープン
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", absPath)
	case "darwin":
		cmd = exec.Command("open", absPath)
	default:
		cmd = exec.Command("xdg-open", absPath)
	}

	if err := cmd.Start(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"path":    absPath,
	})
}

// 設定の取得/更新
func (h *APIHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, config.GetConfig())
	case http.MethodPost:
		var newCfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := config.UpdateConfig(newCfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"config":  config.GetConfig(),
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ユーティリティ: Base64デコード（Data URLヘッダー対応）
func decodeBase64Data(raw string) ([]byte, error) {
	idx := strings.Index(raw, ";base64,")
	cleanBase64 := raw
	if idx != -1 {
		cleanBase64 = raw[idx+8:]
	}
	return base64.StdEncoding.DecodeString(cleanBase64)
}


// 会議サマリー生成リクエスト/レスポンス
type MeetingSummaryRequest struct {
	MeetingID       string `json:"meetingId"`
	Duration        string `json:"duration"`
	Content         string `json:"content"`
	ScreenshotCount int    `json:"screenshotCount"`
}

type MeetingSummaryResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *APIHandler) handleMeetingSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MeetingSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, MeetingSummaryResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := ai.GenerateMeetingSummary(ctx, req.MeetingID, req.Duration, req.Content, req.ScreenshotCount)
	if err != nil {
		writeJSON(w, http.StatusOK, MeetingSummaryResponse{
			Success: false,
			Error:   "会議サマリー生成失敗: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, MeetingSummaryResponse{
		Success: true,
		Result:  result,
	})
}


func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}


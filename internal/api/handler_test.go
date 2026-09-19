package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"voice-reflection-agent/internal/config"
)

func TestMeetingAPI(t *testing.T) {
	// テスト用の一時ディレクトリを設定
	tempDir, err := os.MkdirTemp("", "meeting_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.GetConfig()
	cfg.SaveDir = tempDir
	_ = config.UpdateConfig(cfg)

	// テスト会議ディレクトリ作成
	testFolder := "2026-09-19_120000"
	meetingDir := filepath.Join(tempDir, testFolder)
	imgDir := filepath.Join(meetingDir, "images")
	_ = os.MkdirAll(imgDir, 0755)

	_ = os.WriteFile(filepath.Join(meetingDir, "README.md"), []byte("# Test Meeting\n- **05:30** あってるよ"), 0644)
	_ = os.WriteFile(filepath.Join(meetingDir, "transcript.json"), []byte(`[{"time":"05:30","seconds":330,"text":"あってるよ"}]`), 0644)
	_ = os.WriteFile(filepath.Join(meetingDir, "meeting_audio.mp3"), []byte("dummy-audio-content"), 0644)
	_ = os.WriteFile(filepath.Join(imgDir, "screen_120530.jpg"), []byte("dummy-image-content"), 0644)

	handler := NewAPIHandler()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. GET /api/meetings
	reqList := httptest.NewRequest("GET", "/api/meetings", nil)
	recList := httptest.NewRecorder()
	mux.ServeHTTP(recList, reqList)

	if recList.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /api/meetings, got %d", recList.Code)
	}

	var listResp ListMeetingsResponse
	if err := json.NewDecoder(recList.Body).Decode(&listResp); err != nil {
		t.Fatalf("Failed to decode /api/meetings response: %v", err)
	}

	if len(listResp.Meetings) == 0 {
		t.Fatalf("Expected at least 1 meeting, got 0")
	}

	if listResp.Meetings[0].Folder != testFolder {
		t.Errorf("Expected folder %s, got %s", testFolder, listResp.Meetings[0].Folder)
	}

	if !listResp.Meetings[0].HasAudio || !listResp.Meetings[0].HasTranscript || !listResp.Meetings[0].HasReadme {
		t.Errorf("Expected hasAudio, hasTranscript, hasReadme to be true")
	}

	// 2. GET /api/meetings/detail?folder=...
	reqDetail := httptest.NewRequest("GET", "/api/meetings/detail?folder="+testFolder, nil)
	recDetail := httptest.NewRecorder()
	mux.ServeHTTP(recDetail, reqDetail)

	if recDetail.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /api/meetings/detail, got %d", recDetail.Code)
	}

	var detailResp MeetingDetailResponse
	if err := json.NewDecoder(recDetail.Body).Decode(&detailResp); err != nil {
		t.Fatalf("Failed to decode /api/meetings/detail response: %v", err)
	}

	if detailResp.Folder != testFolder {
		t.Errorf("Expected detail folder %s, got %s", testFolder, detailResp.Folder)
	}

	if len(detailResp.Images) != 1 {
		t.Errorf("Expected 1 image, got %d", len(detailResp.Images))
	}

	// 3. GET /api/meetings/file?folder=...&file=meeting_audio.mp3
	reqFile := httptest.NewRequest("GET", "/api/meetings/file?folder="+testFolder+"&file=meeting_audio.mp3", nil)
	recFile := httptest.NewRecorder()
	mux.ServeHTTP(recFile, reqFile)

	if recFile.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /api/meetings/file, got %d", recFile.Code)
	}

	if recFile.Body.String() != "dummy-audio-content" {
		t.Errorf("Expected file body 'dummy-audio-content', got '%s'", recFile.Body.String())
	}
}

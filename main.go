package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"voice-reflection-agent/internal/api"
	"voice-reflection-agent/internal/config"
	"voice-reflection-agent/internal/web"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	noBrowser := flag.Bool("no-browser", false, "Do not open browser automatically")
	saveDir := flag.String("save-dir", "", "Base directory to save meetings (overrides config)")
	flag.Parse()

	// 設定の初期化
	cfg := config.InitConfig()
	if *saveDir != "" {
		cfg.SaveDir = *saveDir
		_ = config.UpdateConfig(cfg)
	}

	// 空きポートの探索・バインド
	listener, actualPort, err := listenOnPort(*port)
	if err != nil {
		log.Fatalf("Failed to bind port: %v", err)
	}

	staticFS, err := getStaticFS()
	if err != nil {
		log.Fatalf("Failed to load embedded static filesystem: %v", err)
	}
	spaHandler := web.NewSPAHandler(staticFS)

	apiHandler := api.NewAPIHandler()

	mux := http.NewServeMux()

	// APIルート登録
	apiHandler.RegisterRoutes(mux)

	// 静的SPAハンドラ
	mux.Handle("/", spaHandler)

	handler := corsMiddleware(mux)

	server := &http.Server{
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // AIやファイル保存の長時間処理に対応
		IdleTimeout:  60 * time.Second,
	}

	appURL := fmt.Sprintf("http://localhost:%d", actualPort)
	fmt.Printf("\n🎙️  Voice Reflection Agent (Native Engine) is running!\n")
	fmt.Printf("👉 Access URL : %s\n", appURL)
	fmt.Printf("💾 Save Dir   : %s\n", cfg.SaveDir)
	fmt.Printf("🧠 Claude CMD : %s %s\n\n", cfg.ClaudeCommand, cfg.ClaudeArgs)

	if !*noBrowser {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = openBrowser(appURL)
		}()
	}

	// シャットダウン制御
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	fmt.Println("\nShutting down Voice Reflection Agent...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}
	fmt.Println("Server gracefully stopped.")
}

func listenOnPort(startPort int) (net.Listener, int, error) {
	for port := startPort; port < startPort+100; port++ {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return l, port, nil
		}
	}
	return nil, 0, fmt.Errorf("could not find an open port starting from %d", startPort)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

.PHONY: all run build build-windows clean

APP_NAME := voice-reflection-agent

# デフォルト: Mac/ローカル環境向けビルド
build:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/$(APP_NAME) .

# Windows向けポータブル単一exeのクロスコンパイル（Mac上で実行可能・CGO不要）
build-windows:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(APP_NAME).exe .

# ローカル実行（即時起動・ブラウザ自動オープン）
run:
	go run main.go

# クリーンアップ
clean:
	rm -rf bin/

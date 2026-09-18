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

# マルチプラットフォーム一括ビルド（GitHub Actions同等）
dist:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(APP_NAME).exe .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(APP_NAME)-darwin-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(APP_NAME)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(APP_NAME)-linux-amd64 .

# ローカル実行（即時起動・ブラウザ自動オープン）
run:
	go run main.go

# クリーンアップ
clean:
	rm -rf dist/ bin/$(APP_NAME) bin/$(APP_NAME).exe

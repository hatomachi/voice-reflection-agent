# 🎙️ Voice Reflection Agent

iPhoneのボイスメモから「今日もやもやしたこと」「AIに相談したいこと」「明日必ずやるたった1つの石」を喋り、MacBook常駐の最上位AI（`agy` / Gemini 3.5+）が『7つの習慣』の役割定義に照らし合わせて深層リフレクションを自律生成、`personal-vault`（Obsidian）に自動蓄積・Git同期するパイプラインです。

---

## 🌟 特徴

- **Inboundポート開放ゼロ**: 自宅ルーターのポート開放不要。MacBook ➡ GitHub への Outbound HTTPS 通信のみで安全にキューを処理。
- **最上位知能の自律活用**: クラウド単発APIの表面的な相槌ではなく、MacBook上の `agy`（Gemini 3.5+）が `personal-vault/00_役割定義.md` と `00_Dashboard.md` を自律参照して「鋭い指摘（逃避の看破）」と「温かい承認」を生成。
- **夜のフリクションゼロ**: iPhoneボイスメモで喋って「共有」からショートカット1タップ（またはWebAppからワンタップ）。
- **翌朝0.1秒でTODO消化**: 生成されたノートは `webapp-obsidian`（PWA）から瞬時に閲覧でき、「たった1つの石」をタップしてタスク消化可能。

---

## 🏗️ システム構成

```text
voice-reflection-agent/
├── AGENTS.md                                   # AIエージェント向け総合指針
├── README.md                                   # 本ドキュメント
├── bin/
│   └── reflection-daemon.sh                   # 環境変数を整備してデーモンを起動するシェル
├── src/
│   ├── __init__.py
│   ├── config.py                              # パス・ポーリング設定
│   ├── context_loader.py                      # 00_役割定義.md / 00_Dashboard.md 読み込み
│   ├── agy_runner.py                          # agy CLI ヘッドレス呼び出し制御
│   ├── git_sync.py                            # Outbound git fetch / commit / push
│   ├── processor.py                           # キューパース・リフレクション生成・保存
│   └── daemon.py                              # 常駐監視・ワンショット実行 CLI
├── web/                                       # iPhone録音用 Cloudflare Pages PWA
│   ├── index.html                             # 単一SPA（Wake Lock、MediaRecorder、Gemini文字起こし、GitHub送信）
│   ├── manifest.webmanifest                   # PWAマニフェスト設定
│   ├── sw.js                                  # Service Worker (オフラインキャッシュ)
│   └── icon.svg                               # アプリアイコン
├── config/
│   └── com.user.voice-reflection.plist        # macOS launchd 設定（ログイン時自動常駐）
├── templates/
│   └── reflection_prompt.md                   # Gemini 3.5+ 用プロンプトテンプレート
└── wrangler.json                              # Cloudflare Pages 静的配信設定
```

---

## 🚀 使い方

### 1. 手動テスト・ワンショット実行 (`--once`)

キューフォルダ（`personal-vault/00_Inbox/queue/`）にあるキューを1回だけ処理して終了します。

```bash
# voice-reflection-agent ディレクトリで実行
./bin/reflection-daemon.sh --once
```

### 2. 常駐監視デーモンとして起動 (`--daemon`)

バックグラウンドで1分間隔で GitHub を監視し、新着キューを自動検知してリフレクションを生成・push します。

```bash
./bin/reflection-daemon.sh
# 終了時は Ctrl+C (SIGINT / SIGTERM でグレースフル停止)
```

### 3. macOS launchd への常駐登録（Macログイン時に自動起動）

```bash
# plist を LaunchAgents にシンボリックリンクまたはコピー
mkdir -p ~/Library/LaunchAgents
cp config/com.user.voice-reflection.plist ~/Library/LaunchAgents/

# サービスを登録・起動
launchctl load ~/Library/LaunchAgents/com.user.voice-reflection.plist

# 状態確認
launchctl list | grep voice-reflection

# 停止・解除する場合
launchctl unload ~/Library/LaunchAgents/com.user.voice-reflection.plist
```

---

## 📱 iPhone / PC クライアント（Cloudflare Pages PWA）

iPhoneでの夜間セルフリフレクションおよびPCでの会議・画面記録を快適に行うため、**完全サーバーレスのPWA（`web/index.html`）** を同梱しています。

### 🌟 PWAの主な機能
1. **GitHub ＆ 社内 GitLab バックエンド対応**:
   - `webapp-obsidian` や `obsidian-todo-calendar` と同様、GitHub のみならず社内 GitLab（ALB/リバースプロキシ対応、Project ID、PAT）およびローカル保存モードに対応。
2. **Gemini API のオプショナル化（会社環境対応）**:
   - 会社環境など Gemini API が利用できない場合でも、文字起こしをスキップして手動メモやローカル音声保存で快適に利用可能（将来の社内ホストWhisper連携にも道を開く設計）。
3. **PC向け会議・画面記録 ＆ 完全ローカル保存（ZIP一括）モード**:
   - 画面キャプチャ（複数モニタ対応）、スライド変化自動検知JPEGスクショ、ブラウザ内LAME MP3エンコード（NotebookLM直行可能）。
   - Git連携なしでも利用可能で、会議終了後に **「📦 会議一式をまとめて保存 (ZIP)」（MP3 + スクショJPEG群 + 議事録Markdown）** や MP3、README.md を手元PCにワンクリックで完全保存可能。
4. **Screen Wake Lock（画面スリープ防止）**:
   - 録音開始と同時に自動で画面スリープを防止。夜間に数分〜十数分語り続けても画面が暗転・中断しません。
5. **リアルタイム音量ビジュアライザー**:
   - Web Audio API により、ベッドサイドの小声でもマイクが拾えているかを波形・インジケーターで視覚化。
6. **完全サーバーレス＆高セキュリティ**:
   - PAT や API Key はすべて端末の `localStorage` にのみ保存され、外部の中継サーバーは一切不要。

### ☁️ Cloudflare Pages デプロイ手順（3分で完了）
1. Cloudflare Dashboard ➡ **Workers & Pages** ➡ **Create application** ➡ **Pages** ➡ **Connect to Git**
2. 本リポジトリ（`voice-reflection-agent`）を選択
3. ビルド設定：
   - **Framework preset**: None
   - **Build command**: （空欄）
   - **Build output directory**: `web`
4. **Save and Deploy** をクリック ➡ 数秒で公開完了！

### 📲 初期設定
1. ブラウザまたはホーム画面追加アプリから起動
2. 右上の ⚙️ 設定から以下を選択・入力：
   - **保存先バックエンド**: `GitHub`、`社内 GitLab`、または `ローカルのみ`
   - **GitHub**: Token、Owner、Repository、Branch
   - **社内 GitLab**: Base URL、Project ID、Token、Branch
   - **Gemini API Key (任意)**: 個人利用時は [Google AI Studio](https://aistudio.google.com/) のキーを設定。会社環境では空欄でOK。
3. 「接続テスト」を押して疎通を確認し、「設定を保存」をタップすれば準備完了！

---

## 📥 キューファイル形式（JSON）

`personal-vault` リポジトリの `00_Inbox/queue/YYYY-MM-DD-HHmmss.json` にコミットされます。

```json
{
  "date": "2026-09-13",
  "timestamp": "2026-09-13T23:30:00+09:00",
  "text": "今日の音声文字起こしテキスト...",
  "source": "web-pwa"
}
```

※プレーンテキスト形式（`00_Inbox/queue/YYYY-MM-DD-HHmmss.txt`）でも自動処理されます。

---

## 🔗 関連リンク

- [AI Agent 指針 (AGENTS.md)](AGENTS.md)
- [プロダクト作戦ノート (personal-vault)](file:///Users/s-ikari/work/personal-vault/10_%E8%81%B7%E4%BA%BA%E3%83%BB%E7%99%BA%E6%98%8E%E5%AE%B6/voice-reflection-agent.md)
- [人生の役割定義 (personal-vault)](file:///Users/s-ikari/work/personal-vault/00_%E5%BD%B9%E5%89%B2%E5%AE%9A%E7%BE%A9.md)

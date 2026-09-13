# 🎙️ Voice Reflection Agent

iPhoneのボイスメモから「今日もやもやしたこと」「AIに相談したいこと」「明日必ずやるたった1つの石」を喋り、MacBook常駐の最上位AI（`agy` / Gemini 3.5+）が『7つの習慣』の役割定義に照らし合わせて深層リフレクションを自律生成、`personal-vault`（Obsidian）に蓄積する自動化パイプラインです。

---

## 🌟 特徴

- **Inboundポート開放ゼロ**: 自宅ルーターのポート開放不要。MacBook ➡ GitHub への Outbound HTTPS 通信のみで安全にキューを処理。
- **最上位知能の活用**: クラウド単発APIの表面的な要約ではなく、MacBook上の `agy`（Gemini 3.5+）が `personal-vault` 全体を自律探索して「鋭い指摘」と「温かい承認」を生成。
- **夜のフリクションゼロ**: iPhoneボイスメモで喋って「共有」からショートカットを1タップするだけ。
- **翌朝0.1秒でTODO消化**: 生成されたノートは `webapp-obsidian`（PWA）から瞬時に閲覧でき、「たった1つの石」をタップしてタスク消化可能。

---

## 🔗 ドキュメント

- [AI Agent 指針 (AGENTS.md)](AGENTS.md)
- [プロダクト作戦ノート (personal-vault)](file:///Users/s-ikari/work/personal-vault/10_%E8%81%B7%E4%BA%BA%E3%83%BB%E7%99%BA%E6%98%8E%E5%AE%B6/voice-reflection-agent.md)
- [人生の役割定義 (personal-vault)](file:///Users/s-ikari/work/personal-vault/00_%E5%BD%B9%E5%89%B2%E5%AE%9A%E7%BE%A9.md)

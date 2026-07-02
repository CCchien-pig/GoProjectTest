# Project Rules

## 1. 自動記錄 SA 關鍵字 (Automated SA Prompt Logging)
當使用者詢問關於架構決策、資料庫選型、效能優化、資安防護等具備 Solution Architect (SA) 深度的問題，或者是關鍵的 Debug 過程時，你（AI）必須**主動且自動地**將該次問答的精華記錄到 `c:\Projects\CC\Go\GoProjectTest\.docs\SA_Strategy.md` 檔案的「補充三：AI 使用關鍵字紀錄」區塊中。

**執行守則：**
1. 不需要使用者特別下指令（例如「請幫我記錄...」），只要你判斷該對話有 SA 價值，就在回答完使用者的問題後，直接呼叫編輯檔案的工具（`replace_file_content` 或 `write_to_file`）把紀錄寫進去。
2. 紀錄請依照 `SA_Strategy.md` 既有的格式：
   ```markdown
   ### [日期] — [任務目標]
   **關鍵字 / Prompt**：
   > [提煉使用者剛才詢問的核心關鍵字或問題]

   **獲得的關鍵洞見**：
   - [條列式總結你給出的架構/技術洞見]

   **後續驗證**：
   - [建議使用者如何在專案中驗證或實作此觀念]
   ```
3. 寫入完成後，在回覆使用者的訊息中簡單提一句：「*（已自動將本次架構討論記錄至 SA_Strategy.md）*」即可。

## 2. 自動維護開發對話紀錄 (Automated Conversation Record)
每次完成一個重要階段的開發、Bug 修正、或是專案進度的推進後，必須主動將這次的進度與重要發現寫入 `c:\Projects\CC\Go\GoProjectTest\.docs\Conversation Record.md`。

**執行守則：**
1. 將本次對話解決的問題、修正的 Bug（如 ScyllaDB Keyspace 錯誤、Nil Panic 等）摘要寫入。
2. 確保 Conversation Record 中有明確的時間點、處理模組與結論，無需使用者手動要求。
3. 寫入完成後，在回覆使用者的訊息中簡單提一句：「*（已自動將本次進度記錄至 Conversation Record.md）*」。

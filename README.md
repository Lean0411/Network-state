# 網路監控工具

這個程式是一個網路監控工具，專為 Linux 系統設計。它提供以下功能：

1. 測量指定網路介面的**Total, UL, DL Bandwidth**。
2. 測量與目標的**jitter**。
3. 測量與目標的**delay**。
4. 測量與目標的**packet loss rate**。

## 使用方法

1. 確保您的系統是 Linux，並且有 `ping` 命令可用。
2. 修改程式碼中的常量以適應您的需求，例如 `NetworkStatsFilePath`, `interfaceName`, 和 `host`。
3. 編譯程式：

```bash
go build main.go
```

4. 執行二進位檔：

```bash
./main
```

執行程式後，將周期性地顯示網路速度、jitter、延遲和封包丟失率。

## 注意事項

- 程式讀取 `/proc/net/dev` 以獲取網路數據，所以需要在 Linux 上執行。
- 若網路數據似乎回繞或重置，則會顯示一個錯誤消息。

## 錯誤處理

如果在讀取 `/proc/net/dev` 或執行 `ping` 時出現問題，程式會輸出相關的錯誤消息。

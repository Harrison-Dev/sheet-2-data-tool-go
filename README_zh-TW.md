# Excel 結構產生器

一個用於從 Excel 檔案提取結構定義並生成 JSON 資料的工具。支援命令列介面（CLI）和圖形化介面（GUI）兩種操作模式，具備結構化日誌記錄功能。

## ✨ 功能特色

- **自動結構提取**：自動分析 Excel 檔案並生成 YAML 格式的結構定義
- **資料轉換**：根據結構定義將 Excel 資料轉換為 JSON 格式
- **雙模式操作**：
  - CLI 模式：適合自動化處理和腳本撰寫
  - GUI 模式：提供友善的視覺化操作介面
- **結構化日誌**：內建結構化日誌記錄，支援可設定的等級和格式
- **結構更新**：支援增量更新已存在的結構定義
- **跨平台支援**：支援 Windows、macOS（Intel 和 Apple Silicon）

## 🚀 安裝方式

### 下載預編譯版本
從 [Releases](https://github.com/yourusername/sheet-2-data-tool-go/releases) 頁面下載適合您作業系統的版本。

### 從原始碼編譯

需求：
- Go 1.19 或更高版本
- CGO 支援（GUI 模式需要）

```bash
# 克隆專案
git clone https://github.com/yourusername/sheet-2-data-tool-go.git
cd sheet-2-data-tool-go

# 編譯
go build .

# 或使用建置腳本
./scripts/build_macos.sh      # macOS（建立通用二進位檔）
scripts\build_windows.bat     # Windows
```

## 📖 使用方式

### GUI 模式

直接執行程式即可啟動圖形化介面：

```bash
./excel-schema-generator
```

在 GUI 中可以：
1. 選擇包含 Excel 檔案的資料夾
2. 指定結構定義檔案的儲存位置
3. 設定 JSON 輸出資料夾
4. 點擊按鈕執行相應操作

### CLI 模式

#### 1. 生成初始結構定義

從 Excel 資料夾生成基礎結構定義：

```bash
./excel-schema-generator generate -folder /path/to/excel/files [選項]
```

這會掃描指定資料夾中的所有 Excel 檔案，並在當前目錄生成 `schema.yml` 檔案。

#### 2. 更新結構定義

當 Excel 檔案有變更時，更新現有的結構定義：

```bash
./excel-schema-generator update -folder /path/to/excel/files [選項]
```

這會使用 Excel 檔案中發現的任何新欄位或工作表來更新現有的 `schema.yml` 檔案。

#### 3. 生成 JSON 資料

根據結構定義從 Excel 檔案提取資料：

```bash
./excel-schema-generator data -folder /path/to/excel/files [選項]
```

這會生成包含所有 Excel 檔案資料的 `output.json` 檔案，資料格式遵循結構定義。

**共通選項：**
- `-verbose`：啟用詳細日誌
- `-log-level`：設定日誌等級（debug, info, warn, error）（預設："info"）
- `-log-format`：設定日誌格式（text, json）（預設："text"）

**範例：**

```bash
# 使用除錯日誌生成結構
./excel-schema-generator generate -folder ./excel_files -log-level debug -verbose

# 使用 JSON 格式日誌更新結構
./excel-schema-generator update -folder ./excel_files -log-format json

# 僅使用錯誤等級日誌生成資料
./excel-schema-generator data -folder ./excel_files -log-level error
```

## 🔄 工作流程

1. **初始化**：使用 `generate` 命令從您的 Excel 檔案建立初始結構
2. **自訂**：編輯 `schema.yml` 調整資料類型和欄位名稱
3. **更新**：當 Excel 結構變更時使用 `update` 命令
4. **輸出**：使用 `data` 命令生成最終的 JSON 資料

## 📋 結構定義格式

`schema.yml` 檔案結構範例：

```yaml
files:
  example.xlsx:
    sheets:
      Sheet1:
        offset_header: 1        # 標題行位置（從 1 開始）
        class_name: "ExampleData"
        sheet_name: "Sheet1"
        data_class:
          - name: "Id"          # 必須有一個類型為 "int" 的 "Id" 欄位
            data_type: "int"
          - name: "name"
            data_type: "string"
          - name: "value"
            data_type: "float"
          - name: "active"
            data_type: "bool"
```

### 欄位說明

- `offset_header`：標題行的位置（從 1 開始索引）
- `class_name`：工作表的資料類別名稱
- `sheet_name`：Excel 工作表名稱
- `data_class`：欄位定義列表
  - `name`：欄位名稱（區分大小寫）
  - `data_type`：資料類型（string、int、float、bool）

**重要提示**：如果工作表的結構定義中沒有 "Id" 欄位，系統會自動生成一個從 0 開始的連續整數 ID 欄位。

## 🔧 進階功能

### 結構化日誌

應用程式具備完整的結構化日誌記錄：

- **日誌等級**：Debug、Info、Warn、Error
- **日誌格式**：Text（人類可讀）或 JSON（機器可解析）
- **上下文資訊**：所有日誌條目都包含相關上下文，如檔案名稱、工作表名稱等

日誌輸出範例（text 格式）：
```
time=2025-07-30T10:15:30.123+08:00 level=INFO msg="Schema generation completed" file=schema.yml
```

日誌輸出範例（JSON 格式）：
```json
{"time":"2025-07-30T10:15:30.123+08:00","level":"INFO","msg":"Schema generation completed","file":"schema.yml"}
```

## 🏗️ 建置專案

### macOS

```bash
# 建置通用二進位檔（支援 Intel 和 Apple Silicon）
./scripts/build_macos.sh
```

### Windows

```bash
# 建置 Windows 執行檔
scripts\build_windows.bat
```

### 開發建置

```bash
go build .
```

## 🧪 測試

執行所有測試：

```bash
go test ./...
```

執行詳細輸出的測試：

```bash
go test ./... -v
```

執行特定套件測試：

```bash
go test ./excelschema -v
go test ./pkg/logger -v
```

## 📁 專案結構

```
sheet-2-data-tool-go/
├── main.go                    # 主程式入口，支援 CLI
├── gui.go                     # GUI 實作
├── config.go                  # 設定檔管理
├── excelschema/               # 核心功能套件
│   ├── models.go              # 資料結構定義
│   ├── generate-schema.go     # 結構生成邏輯
│   ├── update-schema.go       # 結構更新邏輯
│   ├── generate-data.go       # 資料生成邏輯
│   └── *_test.go              # 完整測試套件
├── pkg/                       # 額外套件
│   └── logger/                # 結構化日誌系統
│       ├── logger.go
│       └── logger_test.go
└── scripts/                   # 建置腳本
    ├── build_macos.sh
    └── build_windows.bat
```

## 📦 依賴套件

- [fyne.io/fyne/v2](https://fyne.io/) - GUI 框架
- [github.com/xuri/excelize/v2](https://github.com/qax-os/excelize) - Excel 檔案處理
- [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2) - YAML 解析
- 內建 `log/slog` - 結構化日誌（Go 1.19+）

## 🔍 問題排解

### 常見問題

1. **自動生成 ID**：當結構定義中沒有 "Id" 欄位時，系統會自動生成從 0 開始的連續 ID
2. **權限錯誤**：確保應用程式對指定目錄具有讀寫權限
3. **空白輸出**：檢查 offset_header 值是否正確指向您的標題行

### 提示

1. **大型檔案**：處理大型 Excel 檔案時，使用適當的日誌等級以獲得更好的效能（在生產環境中避免使用 debug）
2. **結構驗證**：在生成資料之前，務必檢視生成的 schema.yml
3. **資料類型**：確保結構中的資料類型與 Excel 檔案中的實際資料相符

## 📄 授權

[請加入您的授權資訊]

## 🤝 貢獻

歡迎提交 Issue 和 Pull Request！

1. Fork 儲存庫
2. 建立功能分支
3. 為新功能添加測試
4. 確保所有測試通過
5. 提交 pull request

## 👨‍💻 作者

[請加入作者資訊]

## 📝 更新日誌

### v1.1.0（最新版）
- ✨ 新增具有可設定等級和格式的結構化日誌
- 🧪 為所有功能提供完整測試覆蓋
- 📚 改進文件說明
- 🔧 更好的錯誤處理和日誌記錄

### v1.0.0
- 🎉 首次發布
- ✨ Excel 轉 YAML 結構生成
- ✨ 結構更新功能
- ✨ 從 Excel 生成 JSON 資料
- ✨ GUI 和 CLI 介面
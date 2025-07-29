# Excel Schema Generator - Excel 結構產生器

一個用於從 Excel 檔案提取結構定義並生成 JSON 資料的工具。支援命令列介面（CLI）和圖形化介面（GUI）兩種操作模式。

## 功能特色

- **自動結構提取**：自動分析 Excel 檔案並生成 YAML 格式的結構定義
- **資料轉換**：根據結構定義將 Excel 資料轉換為 JSON 格式
- **雙模式操作**：
  - CLI 模式：適合自動化處理和批次作業
  - GUI 模式：提供友善的視覺化操作介面
- **結構更新**：支援增量更新已存在的結構定義
- **跨平台支援**：支援 Windows、macOS（Intel 和 Apple Silicon）

## 安裝方式

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
```

## 使用方式

### GUI 模式

直接執行程式即可啟動圖形化介面：

```bash
./data-generator
```

在 GUI 中可以：
1. 選擇包含 Excel 檔案的資料夾
2. 指定結構定義檔案的儲存位置
3. 設定 JSON 輸出資料夾
4. 點擊按鈕執行相應操作

### CLI 模式

#### 1. 生成初始結構定義

從 Excel 檔案夾生成基礎結構定義：

```bash
./data-generator generate -folder /path/to/excel/files
```

這會在當前目錄生成 `schema.yml` 檔案。

#### 2. 更新結構定義

當 Excel 檔案有變更時，更新現有的結構定義：

```bash
./data-generator update -folder /path/to/excel/files
```

#### 3. 生成 JSON 資料

根據結構定義從 Excel 檔案提取資料：

```bash
./data-generator data -folder /path/to/excel/files
```

這會在當前目錄生成 `output.json` 檔案。

## 工作流程

1. **初始化**：使用 `generate` 命令建立初始結構定義
2. **自訂**：編輯 `schema.yml` 調整資料類型和欄位名稱
3. **更新**：當 Excel 結構變更時使用 `update` 命令
4. **輸出**：使用 `data` 命令生成最終的 JSON 資料

## 結構定義格式

`schema.yml` 檔案結構範例：

```yaml
files:
  example.xlsx:
    sheets:
      Sheet1:
        offset_header: 0
        class_name: "ExampleData"
        sheet_name: "Sheet1"
        data_class:
          - name: "id"
            data_type: "string"
          - name: "name"
            data_type: "string"
          - name: "value"
            data_type: "number"
```

### 欄位說明

- `offset_header`：標題行的偏移量（0 表示第一行）
- `class_name`：資料類別名稱
- `sheet_name`：Excel 工作表名稱
- `data_class`：欄位定義列表
  - `name`：欄位名稱
  - `data_type`：資料類型（string、number、boolean 等）

## 建置專案

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

## 專案結構

```
sheet-2-data-tool-go/
├── main.go              # 主程式入口
├── gui.go               # GUI 介面實作
├── config.go            # 設定檔管理
├── excelschema/         # 核心功能套件
│   ├── models.go        # 資料結構定義
│   ├── generate-schema.go # 結構生成邏輯
│   ├── update-schema.go   # 結構更新邏輯
│   └── generate-data.go   # 資料生成邏輯
└── scripts/             # 建置腳本
    ├── build_macos.sh
    └── build_windows.bat
```

## 依賴套件

- [fyne.io/fyne/v2](https://fyne.io/) - GUI 框架
- [github.com/xuri/excelize/v2](https://github.com/qax-os/excelize) - Excel 檔案處理
- [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2) - YAML 解析

## 授權

[請加入您的授權資訊]

## 貢獻

歡迎提交 Issue 和 Pull Request！

## 作者

[請加入作者資訊]
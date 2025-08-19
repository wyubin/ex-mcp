# swagger-tool

一個用於 **合併多個 Swagger (OpenAPI v2) YAML 檔案** 的 Golang CLI 工具。  
支援自動去重 `tags`，並且對 `tags`、`paths`、`definitions` 自動排序，讓輸出的結果穩定、方便版本控管。

---

## 功能特色

- ✅ 合併多個 swagger yaml 檔案
- ✅ 自動去除重複的 `tags`（依 `name` 判斷唯一性）
- ✅ 自動排序
  - `tags` 按 `name` 排序
  - `paths` 按路徑排序
  - `definitions` 按名稱排序
- ✅ 輸出結果可選擇 **stdout** 或指定輸出檔案
- ✅ 支援 shell wildcard (`*.yaml`)

---

## 使用方法

```sh
# 合併多個檔案並輸出到終端機
go run main.go merge user.yaml item.yaml
# 合併並輸出到指定檔案
go run main.go merge user.yaml item.yaml -o merged.yaml
# 使用 wildcard 合併資料夾內所有 YAML
go run main.go merge *.yaml -o merged.yaml
```

輸出結果 merged.yaml 會包含：
- 合併後的 paths
- 合併後的 definitions
- 去重後排序好的 tags
- 其他 swagger 基本資訊

# TODO
還需要基於以下做 yaml 的 key 順序輸出
```plaintext
swagger:
info:
host:
basePath:
schemes:
consumes:
produces:
tags:
paths:
definitions:
```
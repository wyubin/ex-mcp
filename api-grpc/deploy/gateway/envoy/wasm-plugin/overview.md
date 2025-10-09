🧩 Proxy-Wasm Go SDK 概覽（中文重寫版）
一、什麼是 Proxy-Wasm？

Proxy-Wasm 是一種讓代理伺服器（例如 Envoy）可以動態擴充功能的機制。
它允許開發者以 Wasm（WebAssembly）模組 的方式撰寫外掛（Plugin），而不需要重新編譯或修改 Envoy 本體。

常見用途包括：

HTTP / TCP 請求的攔截與修改

動態授權（authorization）

日誌收集與流量監控

自訂 header、body 處理或流量分流邏輯

在 Envoy 啟動時，它會載入 Wasm 模組並在特定階段呼叫對應的 hook。

二、Proxy-Wasm 模型概念

在 Envoy 中，每個 Wasm Plugin 會在一個獨立的 VM（虛擬機環境） 內運行。
SDK 會提供多層級的 context 讓開發者能針對不同階段的事件進行邏輯處理：

VMContext
 └── PluginContext
      └── (多個) HttpContext / TcpContext


各層的職責如下：

Context 類型	對應生命週期	主要用途
VMContext	整個 Wasm VM 啟動到結束	設定初始環境，例如註冊 filter、初始化全域資源
PluginContext	每個 Plugin instance 生命週期	儲存 plugin 級別設定、啟動定時任務或初始化狀態
HttpContext	每個 HTTP 連線或請求	處理實際請求邏輯（修改 header、body、路由等）
TcpContext	每個 TCP 連線	處理原始 TCP 流量事件
三、Go SDK 的結構

proxy-wasm-go-sdk 是官方 Go 語言的 Proxy-Wasm SDK，
它提供了開發 Envoy Wasm Plugin 所需的介面與 callback。

核心套件架構
proxy-wasm-go-sdk/
├── proxywasm/        # Envoy 與 Plugin 溝通的 API
├── types/            # 事件 hook、callback 與 enum 定義
├── examples/         # 實際範例程式
└── internal/         # 底層 runtime glue code

核心函式
proxywasm.LogInfo("something happened")
proxywasm.GetHttpRequestHeader("authorization")
proxywasm.ReplaceHttpResponseBody([]byte("custom body"))


SDK 提供大量 helper 讓你能方便存取：

Header / Trailer / Body

Metadata

Cluster / Route 資訊

Metric 與 Log

Shared Data 與 Queue（不同 context 間溝通）

四、Context 生命週期流程
1️⃣ VM 啟動階段

Envoy 啟動時建立 Wasm VM。
→ SDK 呼叫 OnVMStart()

func (*vmContext) OnVMStart(vmConfigurationSize int) types.OnVMStartStatus {
    proxywasm.LogInfo("VM started")
    return types.OnVMStartStatusOK
}

2️⃣ Plugin 啟動階段

每個 plugin 被載入時建立一個 PluginContext。
→ 呼叫 OnPluginStart()

這時可載入設定檔、註冊 metric、初始化狀態。

3️⃣ 請求處理階段

Envoy 收到 HTTP/TCP 請求時：
→ 建立 HttpContext 或 TcpContext

這是主要邏輯執行區域，例如：

攔截 request header

修改 body

決定是否中止或繼續 downstream

常見 callback：

OnHttpRequestHeaders(numHeaders int, endOfStream bool)
OnHttpRequestBody(bodySize int, endOfStream bool)
OnHttpResponseHeaders(numHeaders int, endOfStream bool)
OnHttpResponseBody(bodySize int, endOfStream bool)

4️⃣ 銷毀階段

當請求結束、plugin 卸載或 VM 停止時，
Envoy 會呼叫相對應的 OnHttpStreamDone(), OnPluginDone(), OnVMShutdown()。

五、Shared Data 與 Queue

SDK 提供兩種跨 Context 溝通的方式：

Shared Data
儲存 key-value pair，適合同步狀態（如計數器、cache）。

proxywasm.SetSharedData("key", []byte("value"), 0)
val, cas, _ := proxywasm.GetSharedData("key")


Shared Queue
適合同步事件，例如從 HttpContext 發送訊息給 PluginContext。

id, _ := proxywasm.RegisterSharedQueue("myqueue")
proxywasm.EnqueueSharedQueue(id, []byte("event"))

六、Metrics 與 Logging

可以在 Wasm 內直接產生 Envoy metrics：

counter, _ := proxywasm.DefineCounterMetric("request_total")
proxywasm.IncrementCounter(counter, 1)


支援三種型別：

Counter

Gauge

Histogram

同時也可透過 Envoy 的 log level：

proxywasm.LogInfo("processing request...")
proxywasm.LogWarn("unexpected header")

七、錯誤處理與返回控制

在每個 hook 內，可控制是否繼續下游流程或中止：

return types.ActionPause       // 暫停請求等待外部操作
return types.ActionContinue    // 繼續執行後續 filter
return types.ActionResetStream // 立即中止


這些行為可用於：

request 驗證與拒絕

rate limit

延遲回應

動態修改路由目標

八、實際範例（簡化）
type httpContext struct {
    proxywasm.DefaultHttpContext
}

func (ctx *httpContext) OnHttpRequestHeaders(int, bool) types.Action {
    value, _ := proxywasm.GetHttpRequestHeader("x-user-id")
    if value == "" {
        proxywasm.SendHttpResponse(403, nil, []byte("Forbidden"), -1)
        return types.ActionPause
    }
    return types.ActionContinue
}


這個範例示範如何在收到 HTTP 請求時，
檢查自訂 header 並決定是否拒絕。

九、開發與部署流程

1️⃣ 撰寫 Go plugin
2️⃣ 透過 tinygo 或 wasm-opt 編譯成 .wasm
3️⃣ 放入 Envoy 設定中：

- name: envoy.filters.http.wasm
  typed_config:
    "@type": type.googleapis.com/udpa.type.v1.TypedStruct
    type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
    value:
      config:
        vm_config:
          runtime: "envoy.wasm.runtime.v8"
          code:
            local:
              filename: "/etc/envoy/plugin/main.wasm"


4️⃣ 啟動 Envoy，即可載入 plugin。

🔟 總結

Proxy-Wasm 提供 Envoy 可擴充性，讓 Go Plugin 介入資料流。

Go SDK 將 Proxy-Wasm API 包裝為簡單的事件式開發模型。

各層 context 負責不同階段邏輯，支援共享資料與事件佇列。

適合開發：

自訂驗證邏輯

JSON ⇄ gRPC 轉換

請求過濾、流量分析、分流控制
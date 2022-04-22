# API server

## 傳統限流方法的局限性

- 粒度粗
  - 無法為不同用戶, 不同場景設置不同的限流
- 單隊列
  - 共享限流窗口/桶, 一個壞用戶就可能會將整個系統堵塞, 其他正常用戶的請求無法被及時處理
- 不公平
  - 正常用戶的請求會被排到隊尾, 無法及時處理而餓死
- 無優先級
  - 重要的系統指令一併被限流, 系統故障難以修復

## API Priority and Fairness

- APF 的實現依賴兩個非常重要的資源 FlowSchema, PriorityLevelConfiguration
- APF 對請求進行更細粒度的分類, 每一個請求分類對應一個 FlowSchema(FS)
- FS 內的請求又會根據 distinguisher (區分器) 進一步劃分成不同的 Flow
- FS 會設置一個優先級 (Priority Level, PL), 不同優先級的並發資源是隔離的. 所以不同優先級的資源不會相互排擠. 特定優先級的請求可以被優先處理.
- 一個 PL 可以對應多個 FS, PL 中維護了一個 QueueSet, 用於緩存不能及時處理的請求, 請求不回應為超出 PL 的並發限制而被丟棄
- FS 中的每個 Flow 通過 shuffle sharding 算法從 QueueSet 選取特定的 queues 緩存請求
- 每次從 QueueSet 中取請求執行時, 會先用 fair queuing 算法從 QueueSet 中選取一個 queue, 然後從這個 queue 中取出 oldest 請求執行. 所以即便是同一個 PL 內的請求, 也不會出現一個 Flow 內的請求一直佔用資源的不公平現象.

FlowSchema:

```YAML
apiVersion: <xxx>
kind: FlowSchema
metadata:
    name: name # FlowSchema 名
spec:
    distinguisherMethod:
        type: ByNamespace # 區分器
    matchingPrecedence: 800 # 規則優先級
    priorityLevelConfiguration: # 對應規則優先級
        name: workload-high # k get prioritylevel configuration
    rules:
    - resourceRules:
        - resources: # ↓
            - '*'    # ↓
          verbs:     # 對應的資源和請求類型
            - '*'    # ↑
        subjects:
        - kind: User
          user:
            name: system:kube-schedule
```

PriorityLevelConfiguration:

```YAML
apiVersion: etc
kind: PriorityLevelConfiguration
metadata:
    name: global-default
spec:
    limited:
        assuredConcurrencyShares: 20 # 允許的並發請求
        limitResponse:
            queuing:
                handSize: 6 # shuffle sharding 的配置, 每個 FS+distinguisher 的請求會被 enqueue 到多少個隊列
                queueLengthLimit: 50 # 每個隊列中的對象數量
                queues: 128 # 當前 PriorityLevel 的隊列總數
            type: Queue
        type: Limited
```

## 排隊

- 即使在同一優先級內, 也可能存在大量不同的流量源
- 再過載情況下, 放置一個請求流餓死其他流是非常有價值的(尤其是在一個較為常見的場景中, 一個有故障的客戶端會瘋狂地向 kube-apiserver 發送請求, 理想情況下, 這個有故障的客戶端不應該對其他客戶端產生太大的影響)
- 公平排隊算法在處理具有相同優先級的請求時, 實現了上述場景
- 每個請求都被分配到某個 flow 中, 該 flow 由對應的 FlowSchema 的名字加上一個 Flow Disinguisher 來標識
- 這裡的`流區分項`可以是發出請求的用戶,namespace 或者什麼都不不是
- 系統嘗試為不同 flow 中具有相同優先級的請求賦予近似相等的權重
- 將請求劃分到流中之後, APF 功能將請求分配到隊列中
- 分配時使用一種稱為`混洗分片`(Shuffle-Sharding)的技術, 該技術可以相對有效利用隊列隔離`低高強度流`
- 排隊算法的細節可以針對每個優先級進行調整, 並允許管理員在內存佔用, 公平性 ( 當總流量超標時, 各個獨立的流將都會取得進展 ), 突發流量的容忍度以及配對引發的額外延遲之間的進行權衡

## 設置合適的緩存大小

- API Server 與 etcd 之間基於 gRPC 協議進行通信 ,gRPC 協議保證了兩者在大規模集群中的數據高速傳輸. gRPC 基於連接複用的 HTTP/2 協議, 即針對相同分組的對象, API Server 和 etcd 之間共享相同的 TCP 連接, 不同請求由不同 Stream 傳輸.
- 一個 HTTP/2 連接有其 stream 配額, 配額的大小限制了能支持的並發請求. APIServer 提供了集群對象的緩存機制, 當客戶端發起查詢請求時, APIServer 默認會將其緩存直接返回給客戶端, 緩存區大小可以通過參數, "--watch-cache-sizes" 設置. 針對訪問請求比較低多的對象, 適當設置緩存的大小, 極大降低對 etcd 的訪問頻率, 節省網絡調用, 降低了對 etcd 集群的讀寫壓力, 從而提高對象訪問的性能.
- 但是 API Server 也是允許客戶端忽略緩存的, 例如客戶端請求中 ListOption 中沒有設置 resourceVersion, 這時 API Server 直接從 etcd 拉去最新數據並返回給客戶端. 客戶端應盡量避免次操作, 應在 ListOption 中設置 resourceVersion 為 0, API Server 則將從緩存裡讀數據, 而不會直接訪問 etcd.

## 客戶端盡量使用長連接

- 當查詢請求的返回數據較大且此類請求並發量較大時, 容易引發 `TCP 鏈路的阻塞`, 導致其他查詢操作超時. 因此基於 K8s 開發組件時, 例如某些 DaemonSet 和 Controller, 如果要查詢某類對象, 應盡量通過長鏈接 ListWatch 監聽對象變更, 避免全量從 API Server 獲取資源. 如果在同一類應用程序中, 如果有多個 Informer 監聽 API Server 資源變化, 可以將這些 Informer 合併, 減少和 API Server 的長鏈接數, 從而降低對 API Server 的壓力

## 搭建多租戶的 k8s 集群

- 授信:
  - 認證:
    - 禁止匿名用戶訪問, 只允許可信用戶做操作.
  - 授權:
    - 基於授權的操作, 防止多用戶之間互相影響, 比如用戶刪除 k8s 核心服務, 或者 A 用戶刪除或修改 B 用戶的應用.
- 隔離:
  - 可見性隔離:
    - 用戶只關心自己的應用, 無需看到其他用戶的服務和部署
  - 資源隔離:
    - 有些關鍵項目對資源需求較高, 需要專有設備, 不與其他人共享
  - 應用訪問隔離:
    - 用戶創建的服務, 按既定規則允許其他用戶訪問.
- 資源管理:
  - Quota 管理:
    - 誰能用多少資源?

## 回顧 GKV

- Group: 如 core, 應用相關的: apps 組. 每個 Group 複用一個 TCP 連接
- Kind: 定義一個對象的基本類型, 如 node, pd, deployment 等
- Version:
  - Inernal version 和 External version
  - 版本轉換

## 定義對象類型 types.go

- List: 單一對象數據結構
  - TypeMeta
  - ObjectMeta
  - Spec
  - Status

## 代碼生成:

- deepcopy-gen:
  - 為對象生成 DeepCopy 方法, 用於創建對象副本
- client-gen:
  - 創建 Clientset, 用於操作對象的 CRUD
- informer-gen:
  - 為對象創建 Informer 框架, 用於監聽對象變化
- lister-gen:
  - 為對象構建 Lister 框架, 用於為 Get 和 List 操作, 構建客戶端緩存
- conversion-gen:
  - 為對象構建 Conversion 方法, 用於內外版本轉換以及不同版本號的轉換.

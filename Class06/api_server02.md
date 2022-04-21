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
- 分配時使用一種稱為`混洗分片`(Shuffle-Sharding)的技術, 該技術可以相對有效de
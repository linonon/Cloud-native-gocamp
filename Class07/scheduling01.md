# Scheduling

## Kube-scheduler

kube-scheduler 負責分配調度 Pod 到集群內的節點上, 它監聽 kube-apiserver, 查詢還未分配 Node 的 Pod, 然後根據調度策略為這些 Pod 分配節點 ( 更新 Pod 的 NodeName 字段).

- 調度器需要充分考慮諸多的因素:
  - 公平調度
  - 資源高效利用
  - QoS: 服務質量
  - affintity & anti-afiinity: 親和性: 如 A,B 兩個服務器是否在一個機器上
  - data locality: 數據本地化, 用 job 找 data
  - inter-workload interference
  - deadlines

## 調度器

kube-scheduler 調度分為兩個階段, predicate 和 priority

- predicate: 過濾不符合條件的節點
- priority: 優先級排序, 選擇優先級最高的節點

### Predicates 策略

- PodFitsHostPorts: 檢查是否有 Host Ports 衝突
- PodFitsPorts: 同 PodFitsHostPorts
- (最重要)PodFitsResources: 檢查 Node 的資源是否充足, 包括允許的 Pod 數量, CPU, 內存, GPU 個數以及其他的 OpaqueIntResources.
- HostName: 檢查 Pod.Spec.NodeName 是否與候選節點一致
- MatchNodeSelector: 檢查候選節點的 pod.Spec.NodeSelector 是否匹配
- NoVolumeZoneConflict: 檢查 volume zone 是否衝突
- ...

### Predicates plugin 工作原理

通過一個一個策略來過濾 NodeList, 最後得到符合條件的 NodeList

### Priorities 策略

- SelectorSpreadPriority: 優先減少及誒單上同屬一個 Service 或 Replication Controller 的 Pod 數量
- InterPodAffinityPriority: 優先將 Pod 調度到相同的拓撲上 ( 如同一個節點, Rack, Zone 等 )
- LeastRequestedPriority: 優先調度到請求資源少的節點上, 通常用在排隊作業
- BalancedResourceAllocation: 優先平衡各節點的資源使用, 通常用在長穩定的服務商
- NodePreferAvoidPodsPriority: alpha.Kubernetes.io/preferAvoidPods 字段判斷, 權重為 10000, 避免其他優先級策略的影響

## 資源需求

- CPU:
  - requests:
    - K8s 調度 Pod 時, 會判斷當前節點正在運行的 Pod 的 CPU Requeset 的總和, 再加上當前調度 Pod 的 CPU requeset, 計算其是否超過節點的 CPU 的可分配資源
  - limits:
    - 配置 cgroup 以限制資源上限, 當 Pod 競爭時, 會根據數值具體分配 CPU 時間片
- 內存:
  - requests:
    - 判斷節點的剩餘內存是否滿足 Pod 的內存請求量, 以確定是否可以將 Pod 調度到該節點
  - limits:
    - 配置 cgroup 以限制資源上限

## 磁盤資源需求

- 容器臨時存儲 ( ephemeral storage ) 包含日誌和可寫層數據, 可以通過定義 Pod Spec 中的 limits.ephemeral-storage 和 requests.ephemeral-storage 來申請.
- Pod 調度完成後, 計算節點對臨時存儲的限制不是基於 CGroup 的, 而是由 kubelet 定時獲取容器的日誌和容器可寫層的磁盤使用情況, 如果超過限制, 則會對 Pod 進行驅逐.

## Init Container 的資源需求

- 當 kube-scheduler 調度帶有多個 Init 容器的 Pod 時, 只計算 CPU.request 最多的 Init 容器, 而不是計算所有的 Init 容器總和
- 由於多個 init 容器按順序執行, 並且執行完成立即退出, 所以申請最多的資源 init 容器中的所需資源, 即可滿足 init 容器需求.
- kube-scheduler 在計算該節點被佔用的資源時, init 容器的資源依舊會被納入計算. 因為 init 容器在特定情況下可能會被再次執行, 比如由於更換鏡像而引起的 Sandbox 重建時.

## 把 Pod 調度到制定 Node 上

- 可以通過 nodeSelector, nodeAffinity, podAffinity 以及 Taints 和 tolerations 等來將 Pod 調度到需要的 Node 上
- 也可以通過設置 nodeName 參數, 將 Pod 調度到指定 node 節點上
- 比如, 使用 nodeSelector, 首先給 Node 加上標籤: `kubectl label nodes <your-node-name> disktype=ssd`
- 接著, 指定該 Pod 只想運行在帶有 disktype=ssd 標籤的 Node 上.

## Tains 和 Tolerations

Tains 和 Tolerations 用於保證 Pod 不被調度到不合適的 Node 上, 其中 Taint 應用於 Node 上, 而 Toleration 則用於 Pod 上.

- 目前支持的 Taint 類型:
  - NodSchedule: 新的 Pod 不調度到該 Node 上, 不影響正在運行的 Pod;
  - PreferNoSchedule: soft 版的 NoSchedule, 盡量不調度到該 Node 上;
  - NoExecute: 新的 Pod 不調度到該 Node 上, 並且刪除( evict )已在運行的 Pod. Pod 可以增加一個時間( tolerationSeconds ).

然而, 當 Pod 的 Tolerations 匹配 Node 的左右 Taints 的時候可以調度到該 Node 上, 當 Pod 是已經運行的時候, 也不會被刪除( evicted ). 另外對於 NoExecute, 如果 Pod 增加了一個 tolerationSeconds, 則會在該時間之後才刪除 Pod.

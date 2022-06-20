# 將應用遷移至 Kubernetes 平台

## 應用容器化

`穩定性, 可用性, 性能, 安全`

從多維度思考高可用的問題

- 單個實例視角
  - 資源需求
  - 配置管理
  - 數據保存
  - 日誌和指標收集
- 應用視角
  - 冗餘部署
  - 部署多少個實例
  - 負載均衡
  - 健康檢查
  - 服務發現
  - 監控
  - 故障轉移
  - 擴縮容
- 安全視角
  - 鏡像安全
  - 應用安全
  - 數據安全
  - 通訊安全

## 應用容器化的思考

- 應用本身
  - 啟動速度
  - 健康檢查
  - 啟動參數
- Dockerfile
  - 用什麼基礎鏡像
    - 基礎鏡像越小越好
  - 需要裝什麼 Utility
    - lib 越少越好
  - 多少個進程
    - 主次要分清楚, 哪個是決定狀態的主進程
    - Fork bomb 的危害
  - 代碼和配置分離
    - 配置如何管理
      - 環境變量  
      - 配置文件
  - 分層的控制
  - Entrypoint

## 容器額外開銷和風險

log stdout/stderr runtime 轉儲

- Log driver:
  - Blocking mode
  - Non blocking mode
- 共用 Kernel
  - 系統參數配置共享
  - 進程數共享
  - fd 數共享
  - 主機磁盤共享

## 容器化應用的資源監控

- 容器中看到的資源是主機資源
  - Top
  - Java runtime.GetAvailableProcesses()
  - cat /proc/cpuinfo (GOMAXPROXY會在這裡找, 這是不對的)
  - cat /proc/meminfo
  - df -k
- (獲得容器資源的)解決方案
  - 查詢 /proc/1/cgroup 是否包含`kubepods`關鍵字
  - 包含此關鍵詞, 則表明是運行在 Kubernetes 之上.

### 內存開銷

- 配額
  - cat /sys/fs/cgroup/memory/memory.limit_in_bytes
  - 36854771712
- 用量
  - cat /sys/fs/cgroup/memory/memory.usage_in_bytes

### CPU 開銷

- 配額
  - cat /sys/fs/cgroup/cpu/cpu.cfs_quota_us
    - -1 (best effort)
  - cat /sys/fs/cgroup/cpu/cpu.cfs_period_us
    - 100,000
- 用量
  - cat /sys/fs/cgroup/cpuacct/cpuacct.usage_percpu ( 按 CPU 區分 )
    - 140xxxxx 148xxxxx ... ...
  - cat /sys/fs/cgroup/cpuacct/cpuacct.usage
    - 總用量

## Pod spec

- 初始化需求 ( init container )
- 需要幾個主 container
- 權限? Privilege 和 SecurityContext (PSP)
- 共享那些 Namespace (PID, IPC, NET, UTS, MNT)
- 配置管理
- 優雅終止
- 健康檢查
  - Liveness Probe
  - Readiness Prob
- DNS 策略以及對 resolv.conf 的影響
- imagePullPolicy Image 拉去策略

## 在 Kubernetes 上部署應用的挑戰

- 資源規劃
  - 每個實例需要多少計算資源
    - CPU/GPU?
    - Memory
  - 超售需求 (研發傾向要更多的資源)
  - 每個實例需要多少存儲資源
    - 大小
    - 本地還是網盤
    - 讀寫性能
    - Disk IO
  - 網絡需求
    - 整個應用總體 QPS 和 帶寬

## Pod 的數據管理

- local-ssd: 獨佔的本地磁盤, 獨佔的 IO, 固定大小, 讀寫性能高
- local-dynamic: 基於 LVM, 動態分配空間, 效率低

- Node:
  - Root Disk
    - Emptydir (推薦, 生命週期跟隨 pod)
    - hostPath (管理員用)
    - log
    - Configmap (配置文件, Read Only, Mount 進容器就好)
    - Secret (配置文件, Read Only, Mount 進容器就好)

## Pod Disruption Budget

PDB是為了自主中斷時保障應用的高可用

在使用 PDB 時, 你需要弄清楚你的應用類型以及你想要的應對措施:

- 無狀態應用:
  - 目標: 至少有 60% 的副本 Availbale
  - 方案: 創建 PDB Object, 指定 minAvailable 為 60%, 或者 maxUnavailable 為 40%
- 單實例的有狀態應用:
  - 目標: 終止這個實例之前必須提前通知客戶並取得同意.
  - 方案: 創建 PDB Object, 並設定 maxUnavailable 為 0
- 多實例的有狀態應用:
  - 目標最少可用的實例數不能少於某個數 N, 例如 etcd(不能少於 3 個)
  - 方案: 設置 maxUnavailable = 1 或者 minavailable = N, 分別允許每次只刪除一個實例和每次刪除(expeted_replicas - minAvailable)個實例

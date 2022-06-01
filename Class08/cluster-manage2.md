# 生產化集群的管理

## 驅逐管理

- kubelet 會在系統資源不夠時終止一些容器進程, 以空出系統資源, 保證節點的穩定性.
- 但由 kubelet 發起的驅逐只停止 Pod 的所有容器進程, 並不會直接刪除 Pod.
  - Pod 的 status.phase 會被標記為 Failed.
  - status.reason 會被設置為 Evicted.
  - status.message 則會記錄被取出的原因.

## 驅逐策略

kubelet 獲得節點的可用額信息後, 會結合節點的容量信息來判斷當前節點運行的 Pod 是否滿足驅逐條件.

驅逐條件可以是絕對值或百分比, 當監控資源的可使用額少於設定的數值或百分比時, kubelet 就會發起驅逐操作.

kubelet 參數 evictionMinimumReclaim 可以設置每次回收的資源的最小值, 以防止小資源的多次回收.

- kubelet參數: evictionSoft
  - 分類: 軟驅逐
  - 驅逐方式: 當檢測到當前資源達到軟驅逐的閾值時, 並不會立即啟動驅逐操作, 而是需要等待一個寬限期
- kubelet 參數: evictionHard
  - 分類: 硬驅逐
  - 驅逐方式: 沒有寬限期, 一旦檢測到滿足硬驅逐的條件, 就直接終止容器來釋放緊張資源.

### 基於內存壓力的驅逐

memory.available 表示當前系統的可用內存情況.

kubelet 默認設置了 memory.available < 100Mi 的硬驅逐條件.

當 kubelet 檢測到當前節點可用內存資源緊張並滿足驅逐條件時, 會將節點的 MemoryPressure 狀態設置為 True, 調度器會組織 BestEffort Pod 調度到內存承壓的節點.

kubelet 啟動對內存不足的驅逐操作時, 會依照如下的順序選取目標 Pod:

1. 判斷 Pod 所有容器的內存使用量總和是否超出了請求的內存量, 超出請求資源的 Pod 會成為備選目標.
2. 查詢 Pod 的調度優先級, 低優先級的 Pod 被優先驅逐.
3. 計算 Pod 所有容器的內存使用量和 Pod 請求的內存量的差值, 差值越小, 越不容易被驅逐.

### 基於磁盤壓力的驅逐

以下任何一項滿足驅逐條件時, 它將會將節點的 DiskPressure 狀態設置為 True, 調度器不會再調度任何 Pod 到該節點上.

- nodefs.available
- nodefs.inodesFree
- imagefs.available
- imagefs.inodesFree

- 驅逐行為:
  - 有容器運行時分區:
    - nodefs 達到驅逐閾值, 那麼 kubelet 刪除已經退出的容器.
    - imagesfs 達到驅逐閾值, 那麼 kubelet 刪除所有未使用的鏡像.
  - 無容器運行時分區:
    - kubelet 同時刪除未運行的容器和未使用的鏡像

## 日誌管理

節點上需要通過運行 logrotate 的定時任務對系統服務日誌進行 rotate 清理, 以防止系統服務日誌佔用大量的磁盤空間.

- logrotate 的執行週期不能過長, 以防止日誌短時間內大量增長.
- 同時配置日誌的 rotate 條件, 在日誌不佔太多空間的情況下, 保證有足夠的日誌可供查看
- Dokcer:
  - 除了基於系統 logrotate 管理日誌, 還可以依賴Docker 自帶的日誌管理功能來設置容器日誌的數量和每個日誌文件的大小
  - Docker 寫入數據之前會對日職大校進行檢查和 rotate 操作, 確保日誌文件不會超過配置的數量和大小.
- Containerd:
  - 日誌的管理是通過 kubelet 定期 (默認為 10s)執行 du 命令, 來檢查容器日誌的數量和文件的大小的.
  - 每個容器日誌的大小和可以保留的文件個數, 可以通過 kubelet 的配置參數 container-log-max-size 和 container-log-max-files 進行調整.

## Kubernetes 集群可能存在的問題

- kubernetes 服務組件並不會感知以下問題, 就會導致 pod 仍會調度至問題節點.
  - 基礎架構守護程序問題: NTP 服務關閉
  - 硬件問題: CPU, 內存或磁盤損壞
  - 內核問題: 內核死鎖, 文件系統損壞
  - 容器 runtime 問題: runtime 守護程序無響應

### (社區解決方案) node-problem-detector

- Kubernetes 節點診斷工具, 可以將節點的異常, 例如:
  - Runtime 無響應
  - Linux Kernel 無響應
  - 網絡異常
  - 文件描述符異常
  - 硬件問題如 CPU, 內存或者磁盤故障

- node-problem-detector

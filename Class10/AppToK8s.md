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
- 解決方案
  - 查詢 /proc/1/cgroup 是否包含`kubepods`關鍵字
  - 包含此關鍵詞, 則表明是運行在 Kubernetes 之上.

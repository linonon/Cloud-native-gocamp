# 生產化集群的管理

## 生產化集群的考量

- 計算節點:
  - 如何批量安裝和升級計算節點的操作系統
  - 如何管理配置計算節點的網絡信息
  - 如何管理不同 SKU (Stock Keeping Unit)的計算節點?
  - 如何快速下架故障的計算節點?
  - 如何快速擴縮集群的規模?
- 控制平面:

## 操作系統的選擇

- 通用的: Ubuntu, CentOS, etc.
- 專為容器化優化的操作系統:
  - CoreOS, RedHat Atomic, Snappy Ubuntu Core, etc.
- 操作系統評估和選型的標準
  - 是否有生態系統
  - 成熟度
  - 內核版本
  - 對 Runtime 的支持
  - Init System
  - 包管理和系統升級
  - 安全

## 雲原生的原則

- 可變基礎設施的風險
  - 在災難發生的時候, 難以重新構建服務, 持續過多的手工操作, 缺乏記錄, 會導致很難由標準初始化後的服務來重新構建起等效的服務
  - 在服務運行過程中, 持續的修改服務器, 就猶如在程序中的可變變量的值發生變化而引入的狀態不一致的並發風險. 這些對於服務器的修改, 同樣會引入中間狀態, 從而導致不可預知的問題.
- 不可變基礎設施 (immutable infrastructure)
  - 不可變的容器鏡像
  - 不可變的主機操作系統

### Atomic

- 由 Red Hat 支持的軟件包安裝系統
- 多種 Distro
  - Fedora
  - CentOS
  - RHEL
- 優勢
  - 不可變操作系統, 面向容器優化的基礎設施
    - 靈活和安全性好
    - 只有 /etc 和 /var 可以修改, 其他目錄均為只讀
  - 基於 rpm-ostree 管理系統包
    - rpm-ostree 是一個開源項目, 是的生產系統中構建鏡像非常簡單
    - 支持操作系統升級和回滾的原子操作

### 構建 ostree

- rpm-ostree
  - 基於 treefile 將 rpm 包構建成 ostree
  - 管理 ostree 以及 bootloader 配置
- treefile
  - refer: 分支名 ( 版本, CPU 架構 )
  - repo: rpm package repositories
  - packages: 待安裝組件
- 將 rpm 構建成 ostree
  - rpm-ostree compose tree --unified-core --cachedir=cache --repo=./build-repo/path/to/treefile.json

### 加載 ostree

- 初始化項目
  ostree admin os-init centos-atomic-host
- 導入 ostree repo
  ostree remote add atomic <http://ostree.svr/ostree>
- 拉去 ostree
  ostree pull atomic centos-atomic-host/8/x86_64/standard
- 部署 os
  ostree admin deploy --os=centos-atomic-host-centos-atomic-host/8/x86_64/standard -- karg='root=/dev/atomicos/root'

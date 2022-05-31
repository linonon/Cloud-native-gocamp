# Kubelet scheduling & controller

## CNI (Container Network Interface)

- Kubernet 網絡模型設計的基礎原則是:
  - 所有的 Pod 能夠不通過 NAT 就能互相訪問
  - 所有的節點能夠不通過 NAT 就能互相訪問
  - 容器內看見的 IP 地址和外部組件看到的容器 IP 是一樣的.
- Kubernetes 的及群裡, IP 地址是以 Pod 為單位進行分配的, 每個 Pod 都擁有一個獨立的 IP 地址. 一個 Pod 內部的所有容器共享一個網絡棧, 即宿主機上的網絡命名空間, 包括他們的 IP地址, 網絡設備, 配置等都是共享的. 也就是說, Pod 裡面的所有容器能通過 localhost:port 來連接對方. 在 Kubernetes 中, 提供了一個輕量的通用容器網絡接口 CNI(Container Network Interface), 專門用於設置和刪除容器的網絡連通性. 容器運行時通過 CNI 調用網絡插件來完成容器的網絡設置.

## CNI 插件分類和常見插件

- IPAM: IP 地址分配
- 主插件: 網卡設置
  - bridge: 創建一個網橋, 並把主機端口和容器端口插入網橋
  - ipvlan: 為容器添加 ipvlan 網口
  - loopback: 設置 loopback 網口
- Meta: 附加功能
  - portmap: 設置主機端口和容器端口映射
  - bandwidth: 利用 Linux Traffic Control 限流
  - firewall: 通過 iptables 或者 firewalld 為容器設置防火墻規則.

### cidr 計算

`cidr: 192.168.166.128/26` 表示`前 26`都是 0, 後面可 0 可 1, 所以有 2^(32-26) = 2^6 = 64 個可分配的 IP 地址.

## CSI (Container Server Interface)

## 存儲對象關

用戶通過 PVC 來申請存儲. 控制器通過 PVC 的 StorageClass 和請求大小聲明來存儲後端創建新卷, 進而創建 PV, Pod 通過指定 PVC 來引用存儲.

## 生產實踐經驗分享

不同介質類型的磁盤, 需要設置不同的 StorageClass, 以便讓用戶做區分. StorageClass 需要設置磁盤介質的類型, 以便用戶了解該類存儲的屬性.

本地存儲的 PV 靜態部署模式下, 每個物理磁盤都盡量只創建一個 PV, 而不是劃分多個分區來提供多個本地存儲 PV, 避免在使用時分區之間的 IO 干擾.

本次存儲需要配合磁盤檢測來使用. 當集群部署規模化後, 每個集群的本次存儲 PV 可能會超過幾萬個, 其中如磁盤損壞將會是頻發實踐. 此時, 需要在檢測到磁盤損壞, 丟盤等問題後, 對節點的磁盤和相應的本地存儲 PV 進行特定的處理, 例如觸發警告, 自動 cordon 節點, 自動通知用戶等.

對於提供本地存儲節點的磁盤管理, 需要做到靈活管理和自動化. 節點磁盤的信息可以歸一, 集中化管理. 在 loacal-volume-provisioner 中增加部署邏輯, 當容器運行起來時, 拉取該節點需要提供本地存儲的磁盤信息, 例如磁盤的設備路徑, 以 FileSystem 或 Block 的模式提供本地存儲, 或者是否需要假如某個 LVM 的虛擬組 (VG) 等. local-volume-provisioner 根據獲取的磁盤信息對磁盤進行格式化, 或者加入到某個 VG, 從而形成對本地存儲支持的自動化閉環.

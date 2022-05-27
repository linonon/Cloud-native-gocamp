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

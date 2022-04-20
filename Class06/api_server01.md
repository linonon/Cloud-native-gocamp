# K8s APIserver

## kube-apiserver

k8s 的核心組件之一:

- 提供集群管理的 REST API 接口, 包括`認證授權`, `數據校驗`以及`集群狀態`變更等
- 提供其他模塊之間的數據交互和通信的樞紐 (其他模塊通過 API Server 查詢或修改數據, 只有 API server 才直接操作 etcd)

## 認證

開啟 TLS 時, 所有的請求都需要先認證, k8s 支持多重認證機制, 並支持同時開啟多個認證插件

### 認證插件

- X509 證書:
  - 使用 X509 客戶端證書只需要 API server 啟動時配置 `--client-ca-file=SOMEFILE`. 在證書認證時, 其 CN 域用作用戶名, 而組織機構域則用作 group 名
- 靜態 Token 文件:
  - 使用靜態 Token 文件認證只需要 API server 啟動時配置 `--token-auth-file=SOMEFILE`
  - 該文件為 csv 格式, 每行至少包括三列 `token`, `username`, `user id`
- 靜態密碼文件:
  - 與 Token 類似:
- ServiceAccount:
  - 是由 k8s 自動生成的, 並會自動掛載到容器的`/run/secrets/kubernetes.io/serviceaccount`目錄中.
- OpenID
  - OAuth 2.0 的認證機制
- Webhook 令牌身份確認:
  - `authentication-token-webhook-config-file` 指向一個配置文件, 其中描述如何遠程訪問的 Webhook 服務
  - `--authentication-token-webhook-cache-ttl` 用來設定身份認證決定的緩存時間, 默認時長兩分鐘.

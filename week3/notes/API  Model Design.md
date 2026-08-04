
Start Podman VM
```
podman machine init
podman machine start
podman info
podman run --rm docker.io/library/alpine:latest echo "Podman works"

podman stop cloud3-api
```

VM access repo
```
podman run --rm \
  -v "$PWD:/app" \
  -w /app \
  docker.io/library/golang:1.25 \
  go mod init cloud3-api
```

go init and exec
```
./go-container.sh version
./go-container.sh mod init cloud3-api
./go-container.sh run .
```

## 1. 创建实例

```
POST /instances
```

## 2. 列出实例

```
GET /instances
GET /instances/{id}
```

## 3. 删除实例

```
DELETE /instances/{id}
```

## 4. 获取连接信息

```
GET /instances/{id}/connection
```



## Useful Tools

API Framework → 使用 Go Echo

定义 Routes → 在 Echo 中定义 GET、POST、DELETE

实现 CRUD
→ 使用 Go 代码和 Memory Storage

Postman / curl 验证
→ 可以再加 EchoAPI 验证和管理 API 文档
##  Create Flow
用户发送：

```
POST /v1/instances
Content-Type: application/json
Authorization: Bearer ...

{
  "name": "demo-db",
  "instances": 3,
  "storageGi": 20
}
```

## 第一步：接收请求

API Router 找到：

```
POST /v1/instances
```

对应的 Controller。

## 第二步：认证

检查 Token，确认调用者身份。

## 第三步：授权

检查该用户是否有创建实例的权限。

## 第四步：解析 JSON

把 JSON 转换为：

```
CreateInstanceRequest
```

## 第五步：验证

例如：

```
name 合法
instances 在 1–5 之间
storageGi 至少为 5
```

## 第六步：检查冲突

查询 Kubernetes 是否已经存在：

```
Cluster/demo-db
```

如果存在，返回：

```
409 Conflict
```

## 第七步：构建 Kubernetes CR

API 在内存中构建：

```
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: demo-db
  namespace: postgres-demo
  labels:
    platform.example.com/managed-by: paas-api
spec:
  instances: 3
  storage:
    size: 20Gi
```

## 第八步：调用 Kubernetes API

API Pod 使用自己的 Service Account Token 调用 API Server。

## 第九步：Kubernetes 认证与 RBAC

API Server 确认：

```
这个 Service Account 是否可以在 postgres-demo Namespace 中创建 clusters.postgresql.cnpg.io？
```

如果没有权限，Kubernetes 会拒绝请求。

## 第十步：保存 CR

API Server 持久化 Cluster 资源。

## 第十一步：Operator 调和

CloudNativePG Operator Watch 到新 CR，开始创建数据库相关资源。

## 第十二步：API 返回响应

API 不需要阻塞到数据库完全 Ready。

它可以返回：

```
201 Created
Location: /v1/instances/demo-db
```

```
{
  "id": "demo-db",
  "status": "CREATING",
  "instances": 3,
  "storageGi": 20
}
```

## 第十三步：客户端查询状态

稍后客户端执行：

```
GET /v1/instances/demo-db
```

API 读取 Cluster `.status`，转换成产品状态：

```
{
  "id": "demo-db",
  "status": "READY",
  "connection": {
    "host": "demo-db-rw.postgres-demo.svc",
    "port": 5432,
    "database": "app",
    "username": "app"
  }
}
```

密码是否直接返回需要谨慎设计。更安全的做法可能是使用专门的凭证接口、一次性读取、Secret 引用或密码轮换流程，而不是每次 GET 都返回密码。
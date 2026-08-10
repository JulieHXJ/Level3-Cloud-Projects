
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

## API Methods
### 1. 获取server状态

```
GET /health
```
### 2. 列出实例

```
GET /instances
GET /instances/{id}

GET /instances/{id}/connection
```
description: 
Reads the CloudNativePG read-write Service and application Secret. The returned host is intended for clients running inside the Kubernetes cluster.
### 3. 创建实例

```
POST /instances
```

POST: Creates a CloudNativePG Cluster custom resource. PostgreSQL Pods, Services, Secrets, and persistent storage are created asynchronously by the CloudNativePG operator.
### 4. 删除实例

```
DELETE /instances/{id}
```

### 5. Update
```
PUT /instances/{id}
PATCH /instances/{id}
```

PUT： Updates the display name annotation and the desired number of CloudNativePG instances.整体替换
PATCH：修改指定字段



## Useful Tools

OpenAPI doc -> API 的唯一合同：路径、参数、JSON、状态码、JWT
API Framework → Go Echo，运行 HTTP Server、Middleware 和生成的路由
Swagger UI -> 展示并交互测试 API，不是最终用户 UI

定义 Routes → 在 Echo 中定义 GET、POST、DELET

check
```
curl -i http://localhost:8080/instances

curl -i -X POST http://localhost:8080/instances \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "memory-test",
    "instances": 1,
    "storage": "1Gi"
  }'

curl -i http://localhost:8080/instances/1

curl -i -X DELETE http://localhost:8080/instances/1
```
##  Create Flow

```
用户在 Mac 上发送：POST /v1/instances
↓ 
Load Balancer (not implemented)
↓ 
Ingress Controller  (not implemented)
↓ 
Service/cloud3-api 
│ 
│ selector: app=cloud3-api 
▼
PaaS API Pod
	│ 内部：
   Router
	   → Controller
	   → Validation：解析和验证 Request JSON
	   → Business Logic
	│
API Pod 内的 Kubernetes Client：向 API Server 创建 Cluster CR
│ 
│ 使用 ServiceAccount cloud3-api 的身份凭据
▼
ServiceAccount cloud3-api 
│ 
│ RoleBinding 
▼ 
Role cloud3-api 
│ 
│ 允许 create clusters 
▼ 
请求被允许
│ 
│ 
▼ 
Kubernetes API Server：
	验证 Cluster CR 格式 
	↓ 
	调用 CloudNativePG Admission Webhook 
	↓ 
	保存 Cluster CR
	↓ 
	etcd
│ 
│ 
▼ 
5. CloudNativePG Operator：
   Watch 到 Cluster CR
   → 再调用 Kubernetes API Server
   → 请求创建 PVC、Pod、Service、Secret

6. Kubernetes 内置 Controller 和 Scheduler 创建并调度 PostgreSQL Pod
		Scheduler
		   ↓
		决定每个 PostgreSQL Pod 放在哪个 Worker Node
		
		Controller
		   ↓
		确保请求的 Pod、PVC、Service 实际存在
		
		kubelet
		   ↓
		在对应 Worker Node 上启动 PostgreSQL Container

7. 三个 PostgreSQL Pod 分布在三个 Worker Node：
   Worker Node 1
	└── PostgreSQL Pod 1
	    └── Primary
	
	Worker Node 2
	└── PostgreSQL Pod 2
	    └── Replica
	
	Worker Node 3
	└── PostgreSQL Pod 3
	    └── Replica

8. Operator 持续观察并更新 Cluster Status

9. 用户通过 GET /v1/instances/{id}查询：
	Mac
	 ↓
	port-forward
	 ↓
	Service/cloud3-api
	 ↓
	API Pod
	 ↓
	Handler
	 ↓
	k8sStorage.Get()
	 ↓
	Kubernetes API Server
```

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

### 1. 接收请求

API Router 找到：
```
POST /v1/instances
```
对应的 Controller。

### 2. 认证
JWT b检查 Token，确认调用者身份。

### 3. 授权
检查该用户是否有创建实例的权限。

### 4. 解析 JSON
把 JSON 转换为 struct：
```
CreateInstanceRequest
```

### 5. 验证
例如：
```
name 合法
instances 在 1–5 之间
storageGi 至少为 5
```

### 6. 检查冲突

查询 Kubernetes 是否已经存在：
```
Cluster/demo-db
```

如果存在，返回：
```
409 Conflict
```

### 7. 构建 Kubernetes CR

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

### 8. 调用 Kubernetes API
API Pod 使用自己的 Service Account Token 调用 API Server。
### 9. Kubernetes 认证与 RBAC

API Server 确认：
```
这个 Service Account 是否可以在 postgres-demo Namespace 中创建 clusters.postgresql.cnpg.io？
```
如果没有权限，Kubernetes 会拒绝请求。

### 10. 保存 CR

API Server 持久化 Cluster 资源。

### 11. Operator 调和

CloudNativePG Operator Watch 到新 CR，开始创建数据库相关资源。

### 12. API 返回响应

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

### 13. 客户端查询状态

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

密码是否直接返回需要谨慎设计。更安全的做法可能是使用专门的凭证接口、一次性读取、Secret 引用或密码轮换流程，而不是每次 GET 都返回密码。‘




## Local memory workflow：
```
main()

memoryStorage := NewMemoryStorage()
创建 MemoryStorage
        │
        ▼

handler := NewHandler(mStorage)
放进 Handler
        │
        ▼

http.HandleFunc("/instances", handler.instanceHandler)

        │
        ▼

func (h *Handler) instanceHandler(...)

        │
        ▼

h.getInstances(...)

        │
        ▼

h.store.List() 
所有 CRUD 都通过 h.store 调用

        │
        ▼

MemoryStorage.List()
```



## context.Context

http.Request自带：
```
r.Context()
```
Context 主要携带：
- 请求是否被取消；
- 请求是否超时；
- 当前操作是否还应该继续。


## KubeStorage
Go 程序中的一个 struct
```
type KubeStorage struct { 
	client dynamic.Interface 
	coreClient kubernetes.Interface 
	namespace string 
}
```
作用是：
把你的业务模型 `DBInstance` 翻译成 Kubernetes Resource，再把 Kubernetes Resource 翻译回 `DBInstance`。
```
type DBInstance struct {
    ID        string
    Name      string
    Instances int
    Status    string
    CreatedAt string
}
```
## client-go
让 Go 程序能够调用 Kubernetes API


## Dynamic Client 
用来操作CloudNativePG Cluster CR。 

Cluster不是 Kubernetes 内置资源，而是 CloudNativePG 安装 CRD 后新增的自定义资源。所以你使用 Dynamic Client，并通过 GVR 告诉它操作什么资源。

GVR 描述 Kubernetes API 路径：i
```
Group:    postgresql.cnpg.io
Version:  v1
Resource: clusters
```



## Core Client
用于操作 Kubernetes 内置资源：
Service
Secret
Pod
ConfigMap

你的 connection endpoint 需要读取：

```
db-k9zhw-rw Service
db-k9zhw-app Secret
```


## Function calls
### 1. POST
```
1. curl 发送 HTTP 请求
        ↓
2. net/http 找到 /instances
        ↓
3. instancesHandler 判断是 POST
        ↓
4. createInstances 解析并验证 JSON
        ↓
5. h.store.Create(r.Context(), request)
        ↓
6. 实际调用 KubeStorage.Create()
        ↓
7. 构造 CloudNativePG Cluster CR
        ↓
8. Dynamic Client 调用 Kubernetes API Server
        ↓
9. Kubernetes API Server 保存 Cluster CR
        ↓
10. 返回创建后的 CR
        ↓
11. clusterToDBInstance 转成 DBInstance
        ↓
12. API 返回 201 Created
```

### 2. Connection
读取 Operator 创建出来的两个真实资源：
```
{id}-rw
{id}-app
```
RW service 始终把流量转发到当前 PostgreSQL Primary。
App Secret 包含：
```
username
password
dbname
host
port
uri
```

### 3. port-forward
```
Mac localhost:8080
        ↓ kubectl port-forward
Service cloud3-api:80
        ↓ targetPort
API Pod:8080
```


## Test

### 第一层：Unit Test

```
Handler + FakeStorage
```

Varify：
```
Router
JSON decoding
param check
HTTP status code
错误转换
Response body
```

run test
```
./go-container.sh test -v ./...
```

or
```
podman run --rm \
  -v "$PWD:/app" \
  -w /app \
  golang:1.25 \
  go test -v ./...
```

### 第二层：KubeStorage 单元测试

可以使用 Kubernetes 官方的 fake client：
```
dynamicfake.NewSimpleDynamicClient(...)
```
或者：
```
kubernetesfake.NewSimpleClientset(...)
```

验证：
```
KubeStorage.Create 是否生成正确 CR
KubeStorage.List 是否正确转换对象
GetConnection 是否读取正确 Secret
```

它仍然不连接真实 SKE。
这部分对你当前任务可能不是必须，但会让测试更完整。

### 第三层：Integration Test

```
API + real Kubernetes API
```

验证：
```
POST 是否真的创建 Cluster CR
PUT 是否真的修改 spec.instances
DELETE 是否真的删除 CR
Connection endpoint 是否真的读取 Secret
```
Manually done with `kubectl proxy` and Podman

### 第四层：End-to-End Test / port forward test

```
curl
  ↓
Kubernetes Service
  ↓
cloud3-api Pod
  ↓
Kubernetes API
  ↓
Operator
  ↓
PostgreSQL
```

你刚才通过：

```
kubectl port-forward service/cloud3-api 8080:80
```

再调用 API，就是端到端验证的一部分。


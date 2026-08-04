
## 本周架构
```
            用户侧
curl / Web UI / CLI / Terraform
              │
              │ HTTPS + JSON
              ▼
        PaaS REST API
              │
              ├── Authentication
              ├── Validation
              ├── Product Rules
              ├── Instance Model
              └── Error Mapping
              │
              │ Kubernetes Client
              ▼
      Kubernetes API Server
              │
              ▼
       Cluster Custom Resource
              │
              ▼
   CloudNativePG Operator
              │
              ├── Pod
              ├── PVC
              ├── Service
              ├── Secret
              └── Status
              │
              ▼
         PostgreSQL
```



## Client & Server

Client，客户端，是主动发起请求的一方。
Server，服务器，是等待请求、处理请求并返回响应的一方，常表示一个正在运行并监听端口的程序。

Request 和 Response
```
Client
  │
  │ Request
  ▼
Server
  │
  │ Response
  ▼
Client
```


## a HTTP Full Request

```
POST https://api.example.com/v1/instances
```

从客户端执行请求，到服务器返回响应，经历以下过程:
###  1. 解析 URL
协议：    https
主机名（目标服务器）：  api.example.com
端口：    默认 443
路径（请求资源）：    /v1/instances

### 2. DNS 查询 / DNS resolution
网络通信最终通常需要 IP 地址，而不是域名。
客户端 → 检查本地 DNS 缓存 → 检查操作系统缓存 → 询问配置的 DNS Resolver → 得到目标 IP

### 3. 建立 TCP 连接
得到 IP 以后，客户端需要与服务器建立 TCP 连接。HTTP/1.1 和 HTTP/2 通常运行在 TCP 上。
TCP 三次握手：
```
Client                    Server

SYN -------------------->

     <---------------- SYN-ACK

ACK -------------------->
```


### 4. 如果是 HTTPS，建立 TLS 加密连接
HTTP 本身默认不加密。中间设备可以看到：
- Authorization Header
- 用户名
- 数据库配置
- 请求 Body

而HTTPS 即 HTTP over TLS 为
```
HTTP
↓
TLS 加密层
↓
TCP
↓
IP
```

TLS 主要解决三个问题：
1. 加密：第三方不能轻易读取请求内容。
2. 身份验证：客户端需要确认自己连接的服务器确实是：
```
api.example.com
```
而不是攻击者伪装的服务器。
3. 完整性：数据在传输过程中如果被篡改，通信双方能够检测出来。

TLS证书：服务器向客户端提供的数字证书
客户端会检查：
- 证书是否过期；
- 域名是否匹配；
- 是否由受信任的 CA 签发；
- 签名是否有效。

TLS 握手完成后，客户端和服务器协商出用于本次连接的会话密钥。之后 HTTP 请求和响应都会被加密传输。


### 5. 发送 HTTP Request
一个 HTTP 请求通常包括：
- Request Line
- Headers：请求的元数据，也就是“关于这次请求的信息”。数据库名称、存储大小等业务数据通常不应随意塞入 Header。
- 空行
- Body（可选）：承载请求的主要内容，代表客户端希望创建的资源。

### 6. 服务器接收并处理
在 Kubernetes 中能路径：
```
Internet / Internal Client
        ↓
Load Balancer 或 Ingress
        ↓
Kubernetes Service
        ↓
API Pod
        ↓
Web Framework Router
        ↓
Controller
        ↓
Service / Business Logic
        ↓
Kubernetes Client
        ↓
Kubernetes API Server
```

### 7.  发送 HTTP Response
返回：
- Status Line：最重要的是状态码。
- Headers
- 空行
- Body（可选）：响应体是资源的 Representation，也就是资源的一种表现形式。这不一定是服务器内部对象的完整复制。服务器内部可能保存或读取很多信息，但 API 只暴露适合客户端使用的字段

### 8. 客户端读取结果
### 9. 连接复用或关闭

### 10. URL、URI、Path、Query、Header & Body
#### URL
完整资源位置，不只说明资源叫什么，还说明如何找到它。

#### URI
是资源标识符这一更广泛的概念。

```
URI 强调“标识资源”
URL 强调“资源的位置和访问方式”
```

例如：
```
/v1/instances/demo-db
```
在 API 设计讨论中通常称为 resource URI。

#### Path 
用来定位具体资源或资源集合。
典型关系：
```
/instances
└── /instances/{instanceId}
```


#### Query Parameter
```
GET /v1/instances?status=ready&limit=20
```
通常用于：
- 过滤；
- 排序；
- 分页；
- 搜索；
- 控制响应展示方式。

区分：
```
Path：你要访问哪个资源
Query：你想如何筛选或查看资源
```

例如：
```
GET /instances/demo-db
```
是获取一个指定实例。

```
GET /instances?status=ready
```
是获取实例集合中所有 Ready 的实例。


## HTTP Methods
常见方法：
```
GET：读取资源，不改变业务资源状态。
POST：向集合提交并创建新资源，一般返回state code和location
PUT：完整替换，或在确定 URI 上创建资源
PATCH：部分更新
DELETE：删除指定资源
```



## REST 和 HTTP 设计原则
### 1. Safe
该请求只读取信息，不应修改服务器上的业务状态。典型方法：
```
GET
HEAD
OPTIONS
```

### 2. Idempotent 幂等
对同一个资源执行一次或重复执行多次，预期最终状态相同。
- GET 是幂等的，重复读取不会改变资源。
- PUT 通常是幂等的，提交相同完整内容，最终状态仍然相同。
- DELETE 通常是幂等的。删除相同资源两次可能返回 `404`，第一次可能返回 `204`，响应不完全一样，但服务器最终资源状态相同。
- POST 通常**不幂等**，重复发送两次POST，可能创建两个订单。

可通过以下方式避免重复创建：
- 唯一资源名称；
- Idempotency-Key；
- 请求 ID；
- 数据库唯一约束；
- Kubernetes 资源名称唯一性。
### 3.Cacheable 可缓存
表示客户端或中间代理可以保存某个响应，在条件允许时重复使用，不必每次都访问服务器。
GET 最常与缓存结合。
是否可以缓存，通常由响应 Header 控制。



## REST (Representational State Transfer)
REST 是一种架构风格。

### 1. Resource 资源
资源可能是：
```
Instance
Backup
Credential
Plan
Operation
```
REST 更倾向于让 URL 表达资源，让 HTTP Method 表达动作

### 2. Representation 表现形式
资源存在于服务器中，但客户端接收到的是资源的一种表现形式。例如JSON

### 3. URI
### 4. Stateless 无状态通信
核心意思是：
服务器处理每次请求时，所需的请求上下文应当由这次请求本身提供，不能依赖“上一次请求刚好做了什么”。

每次请求应明确包含：
- 要删除哪个资源；
- 谁在发出请求；
- 操作是什么。

### 5. Uniform Interface 统一接口
统一接口表示 API 中的操作方式应保持一致。



## 数据表示格式：YAML / JSON
Kubernetes API 在网络上传输对象时，主要使用 JSON。Kubernetes 用户经常用 YAML 编写清单，但 YAML 是客户端输入格式；对象提交给 API Server 时通常以 JSON 形式进行 HTTP 传输。

`kubectl` 会：
1. 读取 YAML；
2. 解析成内部对象；
3. 将数据序列化成 API Server 能处理的格式，通常是 JSON；
4. 通过 HTTP 请求发送给 Kubernetes API Server。


## HTTP 状态码

主要分类：
```
1xx：信息性响应
2xx：Success
3xx：重定向
4xx：Request Error
5xx：Server Error
```

```
200 OK 请求成功，并通常返回响应 Body。
201 Created 新资源已经成功创建。
202 Accepted 请求已经接受，但异步处理尚未完成。
204 No Content 请求成功，但响应没有 Body。

400 Bad Request 请求格式错误，或基本输入无效。
401 Unauthorized 客户端尚未通过身份认证。
403 Forbidden 表示服务器知道你是谁，但你没有执行该操作的权限。
409 Conflict 请求在语法上可能完全正确，但与当前资源状态冲突。
422 Unprocessable Content 表示服务器能解析请求格式，但业务字段无法处理

500 Internal Server Error 服务器发生未预期错误。
```

##  OpenAPI
OpenAPI 是一种机器可读的 API 描述规范。

1. 它是 API Contract
2. 自动生成文档
3. 自动生成客户端SDK
4. 生成服务器接口骨架
5. 自动验证
6. mock server



## Kubernetes API





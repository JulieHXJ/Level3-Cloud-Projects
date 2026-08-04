## Operator

是一个运行在 Kubernetes 里的自动化运维程序，用来实现应用生命周期的管理和自动化。

CloudNativePG 是一个专为 Kubernetes 设计的开源PostgreSQL Operator，专门管理 PostgreSQL这种又复杂生命周期的应用。 当k8s收到创建请求时，会把这段数据保存进 etcd，但它本身不知道怎样安装 PostgreSQL， operator知道：
- 需要什么容器镜像
- 需要几个数据库实例
- 需要什么 Service
- 需要什么 Secret
- 需要什么 PVC
- 主节点失败后怎么办
- 配置变化后怎样滚动更新

它把数据库管理员的手动操作写成程序，例如：
- 创建 PostgreSQL
- 配置 PostgreSQL
- 初始化数据库
- 创建用户和密码
- 连接持久化磁盘
- 检测数据库状态
- 处理 Pod 故障
- 维护 Primary/Replica
- 执行滚动更新
- 更新 Service 指向

## Components
### 1. CRD
CustomResourceDefinition 自定义资源定义。

CRD 定义一个新的kubernetes资源，写明允许出现哪些字段、字段是什么类型，以及哪些字段是必需的，可以把 CRD 类比成 Java 的类定义。

规定用户能填写字段：
- instances
- storage.size
- imageName
- postgresql.parameters
- bootstrap
- Backup
### 2. CR
Custom Resource 自定义资源（期望状态）。
根据CRD创建的具体对象，CRD 是类型，CR 是具体实例。 类比java class和declaration。

### 3. Controller
Operator内部持续运行的程序或组件，持续向kubernetes API server查询：
- 观察 PostgreSQL CR
- 读取当前 Pod、PVC、Service、Secret 状态
- 决定是否需要创建或修改资源
- 更新 PostgreSQL CR 的 status

### 4. Reconciler Pattern
Operator 内部最核心的逻辑，不断把实际状态修正为期望状态。

持续运行以下步骤：
1. 读取 CR 中的 Desired State
2. 查询集群中的 Current State
3. 比较两者
4. 创建、修改或删除资源
5. 更新 CR 的 status
6. 等待下一次事件
7. 再次执行
### 总结
CR = 期望状态Desired state，Operator观察-比较-修正
Controller = 执行工作的程序
Reconcile = Controller 对某个资源进行一次 “读取 → 比较 → 修正”
Reconciliation loop = Controller 不断重复调用 reconcile

  


## CloudNativePG Service

背景：当一个 Pod 被删除并重建，新 Pod 的IP就会改变，如果应用直接连接旧 IP，数据库一旦重建，应用就找不到它。

Service 提供一个稳定的虚拟 IP 和 DNS 名称，然后把请求转发给后面的健康 Pod。Pod 可以被替换，Service 的名字不需要改变。

Kubernetes 的 `ClusterIP` Service 默认提供集群内部虚拟 IP，并将请求转发给匹配的后端 Pod

CloudNativePG 默认创建三类 Service：
```
managed-postgres-rw -> Primary (read & write)
managed-postgres-ro -> Replica (read only)
managed-postgres-rw -> both
```

-rw → 指向 Primary（可以读，也可以写）
-ro → 指向 Replica（用于只读请求）-> 目前没有Endpoint，表示 Service 只能从 Kubernetes 集群内部访问
-r → 指向任意可读 PostgreSQL 实例

原因：更安全的默认配置
```
Kubernetes 内部 Pod
    ↓
managed-postgres-rw:5432
    ↓
PostgreSQL Primary
```

当有三个instance时
```
managed-postgres-rw
└── 只指向 Primary

managed-postgres-ro
├── Replica 1
└── Replica 2

managed-postgres-r
├── Primary
├── Replica 1
└── Replica 2
```

### Primary 
主要数据库实例，负责数据的接受和修改：
```
INSERT
UPDATE
DELETE
CREATE TABLE
```


### Replica 
Primary 的副本。

当 Primary 中的数据变化，PostgreSQL 会将相关变化复制给 Replica。Replica 的两个主要用途是：
- 分担读取压力
- 提供高可用基础
- 在 Primary 故障时作为 Failover 候选

## Secret
Kubernetes 中专门保存少量敏感信息的对象，Pod 可以把 Secret 作为环境变量或文件使用，从而避免把密码直接写进应用代码或镜像。

Secret 本身不是数据库数据，managed-postgres-app是应用连接数据库用的凭证，只保存连接信息，不保存你创建的表，也不保存业务记录

Service→ 解决“数据库在哪里”
Secret→ 解决“我用什么身份登录”

## PostgreSQL Pod 调度

为了提高数据库可用性，多个 PostgreSQL 实例不应全部运行在同一 Worker Node。
```
PostgreSQL Pods
→ 使用 required anti-affinity
→ 强制分布在三个 Worker Nodes

Operator Pod
→ 保持默认自动调度

Client Pod
→ 保持默认自动调度
```
## Client pod

Client Pod 用于从 Kubernetes 集群内部测试 PostgreSQL。
```
┌──────────────────────────────┐
│ psql-client Pod              │
│                              │
│ PGHOST ───────┐              │
│ PGUSER ───┐   │ 来自 Secret  │
│ PGPASSWORD│   │              │
└───────────│───│──────────────┘
            │   │
            │   ▼
            │ managed-postgres-rw:5432
            │          Service
            │             │
            │             ▼
            │  managed-postgres-1
            │     Primary Pod
            │             │
            └──认证───────┤
                          │ SQL 写入
                          ▼
                 PVC managed-postgres-1
                          │
                          ▼
                          PV
                          │
                          ▼
                 STACKIT Cinder Volume
```

用 Deployment 创建 Client 应用

Kubernetes 官方支持通过 node affinity 等规则限制 Pod 可以调度到哪些节点


## SQL Functions

#### 1. 连接primary进入数据库：
```
kubectl exec -it \
  -n postgres-demo \
  deployment/psql-client \
  -- psql
```

查看连接信息：
```
\conninfo
```
#### 2. SELECT：

查看数据库：
```
SELECT
    version(),
    current_database(),   #数据库
    current_user,   #用户
    inet_server_addr(),
    pg_is_in_recovery();  # f -> primary, t -> replica
```

读取表数据：
```
SELECT * FROM platform_test; # 查看表的所有列

SELECT id, message, created_at
FROM platform_test;

SELECT id, message, created_at
FROM platform_test
ORDER BY id;  # 排序

SELECT *
FROM platform_test
WHERE id = 2;  # 条件查询
```
#### 3. TABLE：

查看表：
```
\dt
```

创建表：
```
CREATE TABLE IF NOT EXISTS platform_test (
    id BIGSERIAL PRIMARY KEY,  # 自动生成递增整数
    message TEXT NOT NULL, # 不能空
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

查看表结构
```
\d platform_test
```

#### 4. INSERT
```
INSERT INTO platform_test (message)
VALUES ('Managed PostgreSQL on STACKIT SKE works')
RETURNING *; # return the line after inserting
```


#### 5. UPDATE 修改数据
```
UPDATE platform_test
SET message = 'Client successfully connected through the rw Service'
WHERE id = 3
RETURNING *;
```

#### 6.  DELETE
删除数据：
```
DELETE FROM platform_test
WHERE id = 2
RETURNING *;
```

删除表：
```
DROP TABLE platform_test;
```

#### 7. 退出SQL
```
\q
```
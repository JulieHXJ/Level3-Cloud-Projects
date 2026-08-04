## 背景知识

PaaS：Platform as a Service
是一种云计算服务模式，它为开发人员提供开发、运行和管理应用程序的完整云端平台

IaaS、Kubernetes 和 PaaS 是三层不同的东西：

第一层：Infrastructure
VM、网络、磁盘、Load Balancer

第二层：Kubernetes Platform
Cluster、Node、Pod、Service、PVC

第三层：PaaS Product
Managed PostgreSQL、Managed Redis、Message Queue
 

## SKE (STACKIT Kubernetes Engine)
### 1. SKE管理责任

STACKIT 管理
```
Managed Control Plane 
├── Kubernetes API Server 
├── etcd 
├── Scheduler 
├── Controller Manager 
├── Control Plane 可用性 
└── Control Plane 备份和恢复
```
  
你管理
```
├── Node Pool 配置
├── Kubernetes Namespace
├── Operator
├── CRD 和 CR
├── PostgreSQL 数据库
├── Service 和访问方式
├── Persistent Volume 中的数据
└── 数据库备份
```

### 2. K8s核心资源

- Namespace：用于把资源逻辑隔离。cnpg-system 存放 Operator，database-platform 存放 PostgreSQL 产品。不同 Namespace 中可以存在同名资源。

- Pod：container实际运行的位置。本项目中的 PostgreSQL 最终会运行在一个或多个 Pod 中。

- Deployment：通常用于运行无状态应用。

- Service：提供稳定网络地址。Pod 名称和 IP 可以变化，但用户不能每次都去寻找新 Pod IP，因此使用 Service

- Secret：存储用户名、密码、证书等敏感信息。密码不能直接写进 Git 仓库。

- ConfigMap：存储非敏感配置。

- PersistentVolumeClaim (PVC)：数据库不能只把数据写在容器内部。Pod 删除后，容器文件系统可能消失，因此数据库需要 PVC 请求持久化存储。

- StorageClass：定义 Kubernetes 应该怎样动态创建 Persistent Volume。因此创建数据库前必须检查：

```
kubectl get storageclass
```

### 3. 声明式模型

Kubernetes 最重要的思想不是“执行一次命令”，而是声明你要的状态，然后由 Controller 持续把现实状态调整成目标状态。Kubernetes Controller 就是持续运行的控制循环。它观察集群状态，并不断让当前状态靠近 spec 中声明的目标状态。

通常 Kubernetes 对象中：
```
spec = 用户要求的目标状态
status = Controller 观察到的当前状态
```

所以检查一个 CR 时，不仅要看 spec，还要看：
```
kubectl get cluster demo-db -o yaml
```


## 数据存储关系

```
PostgreSQL 数据文件 
↓ 
Pod 中挂载的数据目录 
↓ 
PVC managed-postgres-1 
↓ 
PV pv-... 
↓ 
Cinder CSI Driver 
↓ 
STACKIT Block Storage Volume
```

## 架构
  ```
  Mac / CI Pipeline
        │
        │ terraform apply
        ▼
STACKIT Terraform Provider
        │
        │ 调用 STACKIT SKE API
        ▼
        SKE Cluster
		├── Managed Control Plane 
        │     ├── Kubernetes API Server 
        │     ├── Scheduler 
        │     ├── Controller Manager 
        │     └── etcd 
        │ 
        └── Node Pool: learning-pool-z1 
		    ├── Worker Node 1
		    │   ├── Operator Pod
		    │   ├── PostgreSQL Pod 1
		    │   │      └── 角色：Primary
		    │   └── 其他系统 Pod
		    │
		    ├── Worker Node 2
		    │   ├── PostgreSQL Pod 2
		    │   │      └── 角色：Replica
		    │   └── 其他系统 Pod
		    │
		    └── Worker Node 3
		        ├── PostgreSQL Pod 3
		        │      └── 角色：Replica
		        ├── Client Pod（可能在这里，也可能在其他 Node）
		        └── 其他系统 Pod
  ```

## SKE 组件

### 1. Kubernetes Cluster 
 是一个管理范围，包含：
- Control Plane
- 很多个Worker Node 
- 网络
- 存储接口
- Kubernetes API
- 系统组件
- 运行在其中的应用

### 2. Control plane：

#### API Server
是 Kubernetes 的统一管理入口，所有管理请求都要经过API server：
- `kubectl`
- Terraform Kubernetes Provider
- Argo CD
- Kubernetes Operator
- Kubernetes Dashboard
- 自定义程序


https请求进入API Server后：
- 验证用户身份。
- 检查用户权限。
- 验证资源格式和字段。
- 执行 Admission Controller 或 Webhook。
- 接受或拒绝请求。
- 将对象状态保存到 etcd。

例：
  ```
  kubectl apply -f cluster.yaml
  ```

#### Controller Manager

负责实现 API 对象所表达的行为，并持续让当前状态靠近目标状态。它持续执行：
- 用户想要什么？
- 现在实际是什么？
- 两者是否一致？
- 不一致的话怎样修正？

  

#### Scheduler

Control Plane 中的`kube-scheduler` 负责排除不满足资源和调度限制的节点，为尚未分配 Node 的 Pod 选择 Worker Node。

通常分为两个阶段：
1. 过滤：排除不满足条件的 Node，例如：
- CPU 不足
- 内存不足
- NodeSelector 不匹配
- Node Affinity 不匹配
- 存在无法容忍的 Taint
- PVC 或可用区条件不满足
- Pod Anti-Affinity 不允许

2. 评分：对剩余 Node 评分，并选择最适合的 Node。

#### etcd

Kubernetes backing store，保存 Kubernetes 的集群状态，如：
- Node
- Namespace
- Deployment
- Pod
- Service
- Secret
- ConfigMap
- CRD
- CR
- Pod 调度结果
- 用户声明的副本数量


#### 协作流程：

当用户提交一个 Deployment 时：

```
1. kubectl 把 YAML 发送给 API Server
2. API Server 验证请求
3. Deployment 对象被保存到 etcd
4. Deployment Controller 发现缺少 ReplicaSet
5. Controller 创建 ReplicaSet
6. ReplicaSet Controller 创建 Pod 对象
7. Scheduler 为 Pod 选择 Worker Node
8. 目标 Node 上的 kubelet发现该 Pod 被分配给自己
9. kubelet 调用 Container Runtime 创建容器
10. kubelet向 API Server 报告 Pod 状态
11. Controller 持续比较 Desired State 和 Current State
```


### 3. Worker Node
对应由SKE 创建和管理的VM，是实际运行应用工作负载的机器。
```
Worker Node / VM
│
├── kubelet
├── container runtime
├── Kubernetes 网络组件
├── CSI 存储组件
│
├── Pod A
│   └── my-app container
│
├── Pod B
│   └── PostgreSQL container
│         └── CloudNativePG controller-manager 程序
│
└── Pod C
	└── Operator container
```


#### kubelet

确保被分配到本 Node 的 Pod 和容器实际运行。负责：
- 查看哪些 Pod 被分配到本 Node
- 调用 Container Runtime 创建容器
- 确保容器保持运行
- 执行健康检查
- 挂载 Volume
- 报告 Node 状态
- 报告 Pod 状态

#### Container Runtime

负责真正创建和运行容器，包括：
- 拉取容器镜像
- 创建容器
- 启动容器进程
- 停止容器
- 删除容器
- 管理容器运行状态


### 4. Pod

Kubernetes 真正调度的最小单位。

Kubernetes不会直接调度 Docker Container，而是调度 Pod。Pod 是 Kubernetes 中最小的可部署计算单元。

一个 Pod 可以包含一个或多个紧密协作的容器，这些容器共享网络和可挂载的存储，并且总是一起被调度到同一台 Node：
```
Worker Node
└── Pod
	└── 一个主要应用Container
```


同一个pod中的containers共享：
- 同一个 Pod IP
- 同一个网络空间
- 可以通过 localhost 互相通信
- 可以挂载相同的 Volume
- 一起被放到同一台 Node
  

Pod 是临时的，可能因为以下原因消失，重新创建的 Pod，名字，IP 地址，所在 Node 可能改变：
- Node 故障
- 应用崩溃
- 升级
- 扩缩容
- 重新调度
- 人为删除


### 5. Service 和网络访问

为不断变化的 Pod 提供稳定地址。

Service 是 Kubernetes API 中的一个逻辑网络对象。它为一组 Pod 提供稳定的 IP 地址和 DNS 名称，即使后面的 Pod 被删除、重建或迁移，Service 仍然可以把流量转发给当前健康的 Pod。

Kubernetes 为 Service 和 Pod建立 DNS 记录，因此集群内的应用通常通过 Service 名称通信，而不是记忆 Pod IP。

职责：
- 稳定地址
- 服务发现
- 把流量转发到健康 Pod
- 在多个 Pod 之间分配连接

#### 常见service1- ClusterIP

只能从 Kubernetes 集群内部访问，数据库通常首先使用这种类型。
App Pod → PostgreSQL Service → PostgreSQL Pod

#### 常见service2 - LoadBalancer

通过云平台创建 Load Balancer，让集群外部可以访问
Internet → STACKIT Load Balancer → Service → App Pod

#### 常见service3 - NodePort

在每个 Worker Node 上开放一个端口


### 6. PVC、PV 和 StorageClass

PersistentVolumeClaim：用户或应用提交的存储申请。

PersistentVolume：Kubernetes 中代表实际存储的对象。

数据库需要把数据写入独立于 Pod 生命周期的 Persistent Volume。Kubernetes 的 PV 拥有独立于任何使用它的 Pod 的生命周期。
- PersistentVolume（PV）是集群中的一块存储；
- PVC 是用户对存储的申领；
- StorageClass 描述存储的类型和动态创建方式。

链路：
```
Database Pod
│
│ 挂载
▼
PVC：我要 10 GiB
│
│ 匹配或动态创建
▼
PV：Kubernetes 中代表这块存储的对象
│
│ 通过 CSI / 云平台存储接口
▼
STACKIT 底层磁盘或存储系统
```




1. ArgoCD
    - Application Healthy
    - Synced
2. kubectl

查看 Pod

```
kubectl get pods -n postgres-demo
```

查看 Service

```
kubectl get svc -n postgres-demo
```

取得密码

CloudNativePG 默认会建立 Secret：

```
kubectl get secrets -n postgres-demo
```







```
kubectl get cluster -n postgres-demo
```




显示：

```
Ready
Instances:4
```

3. **kubectl get pods**

```
managed-postgres-1 Running
managed-postgres-2 Running
managed-postgres-3 Running
managed-postgres-4 Running
```

进入 Pod

```
kubectl exec -it managed-postgres-1 \
    -n postgres-demo \
    -- bash
```

连接数据库：
```
psql \
    -h managed-postgres-rw \
    -U app \
    -d app
```

4. **psql**

```
SELECT version();
```

输出：

```
PostgreSQL 17.x
```

这样就完整证明了：

> Git → ArgoCD → Kubernetes → CloudNativePG → PostgreSQL → 我能够真正连接数据库。
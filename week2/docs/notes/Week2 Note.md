[[SKE Note]]
## 最终架构
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
		    │   │      └── Role：Primary
		    │   └── system Pod
		    │
		    ├── Worker Node 2
		    │   ├── PostgreSQL Pod 2
		    │   │      └── Role：Replica
		    │   └── system Pod
		    │
		    └── Worker Node 3
		        ├── PostgreSQL Pod 3
		        │      └── Role：Replica
		        ├── Client Pod（可能在这里，也可能在其他 Node）
		        └── system Pod
```

## SKE Understanding 

```
SKE
└── Kubernetes Cluster
    ├── Managed Control Plane
    │   ├── API Server
    │   ├── etcd
    │   ├── Scheduler
    │   └── Controller Manager
    │
    └── Node Pool
        ├── Worker Node 1
        │   ├── Pod
        │   └── Pod
        │
        └── Worker Node 2
            └── Pod
                 │
                 └── PVC → PV → Persistent Storage
```


Stackit Service account
julie-tf-qxennsi8@sa.stackit.cloud

service account key into server ENV:
STACKIT_SERVICE_ACCOUNT_KEY_PATH="$HOME/.config/stackit/week2-terraform-key.json"


## Terrafrom CMD
```
cd ~/terraform-lab/week2/infra
```

检查 State：

```
terraform state list
```


```
terraform plan -out=tfplan
terraform apply "tfplan"
```


## Export Cluster kubeconfig (replaced)
save into yaml file:
```
umask 077
terraform output -raw kubeconfig > kubeconfig.yaml
export KUBECONFIG="$PWD/kubeconfig.yaml"
```


## SKE Validation
### Control Plane
```
kubectl cluster-info
kubectl version
```

### Node Pool & Worker Node
```
kubectl get nodes -o wide
```

```
NODE_NAME="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"
echo "$NODE_NAME"
kubectl describe node "$NODE_NAME"
```

Check：
节点是否 Ready
Kubernetes 版本
节点内网 IP
容器运行时
操作系统

### System Pod
```
kubectl get pods -A -o wide
```

### StorageClass
```
kubectl get storageclass
```

Check：defualt 
### Namespace
```
kubectl get namespaces
```

### Deployment
```
kubectl get deployment -n cnpg-system
```

### CSI Driver
```
kubectl get csidriver
```

## Create Test Node
```
kubectl run smoke-test \
  --image=registry.k8s.io/pause:3.10 \
  --restart=Never
```
Check:
```
kubectl get pod smoke-test -o wide -w
```




## Operator 
### 1. Install
```
Kubectl apply --server-side -f

kubectl rollout status deployment \
  -n cnpg-system \
  cnpg-controller-manager
```

### 2. Check
```
echo "=== Namespace ==="
kubectl get namespace cnpg-system

echo
echo "=== Operator Deployment ==="
kubectl get deployment -n cnpg-system

echo
echo "=== ReplicaSet and Pod ==="
kubectl get replicaset,pod -n cnpg-system -o wide

echo
echo "=== Service and ServiceAccount ==="
kubectl get service,serviceaccount -n cnpg-system
```
#### ReplicaSet & Pod
```
kubectl get replicaset,pod -n cnpg-system -o wide
```

#### CloudNativePG CRD
```
kubectl get crd | grep postgresql.cnpg.io
```

Check: 是否成功扩展k8s API，如：
```
backups.**postgresql.cnpg.io**   
clusters.**postgresql.cnpg.io**   
databases.**postgresql.cnpg.io**
```

在确认kubernetes已经认识该CRD：
```
kubectl get crd clusters.postgresql.cnpg.io
```
#### Webhook
```
kubectl get validatingwebhookconfiguration,mutatingwebhookconfiguration |
grep cnpg
```

Check两类 webhook：

```
ValidatingWebhookConfiguration
MutatingWebhookConfiguration
```

确认名称包含 `cnpg` 的 webhook configuration。

#### Operator running properly
```
kubectl get deployment,pod -n cnpg-system
```

## PostgreSQL
### 1. Preparasion - create namespace

File path:
```
terraform-lab/week2/docs/postgres/namespace.yaml
```

Namespaces:
```
cnpg-system
└── Operator Pod

postgres-demo
├── PostgreSQL CR 
├── PostgreSQL Pod 
├── PVC 
├── Services 
└── Secrets
```

### 2. Create PostgreSQL CR
```
name: managed-postgres
→ 数据库集群名称

namespace: postgres-demo
→ 产品资源所在空间

instances: 1
→ 创建一个 PostgreSQL 实例

database: app
→ 初始化一个名为 app 的数据库

owner: app
→ 创建 app 用户作为数据库 owner

size: 5Gi
→ 申请 5 GiB 持久化磁盘

storageClass: premium-perf1-stackit
→ 使用 STACKIT Cinder Block Storage
```

### 3. Apply CR
```
kubectl apply -f docs/postgres/cluster.yaml
```

### 4. Observation
CR status:
```
kubectl get clusters.postgresql.cnpg.io \
  -n postgres-demo \
  -w
```

Resources creating:
```
kubectl get pod,service,pvc,secret \
  -n postgres-demo \
  -w
```

Service：
```
kubectl get pod,service,pvc,secret -n postgres-demo
```

### 5. Check Location
```
echo "=== Worker Nodes ==="
kubectl get nodes -o wide

echo
echo "=== Operator placement ==="
kubectl get pods -n cnpg-system -o wide

echo
echo "=== PostgreSQL placement ==="
kubectl get pods -n postgres-demo -o wide

echo
echo "=== Current Primary ==="
kubectl get cluster managed-postgres -n postgres-demo
```


## Client pod
### 1. Deployment
```
Deployment
└── ReplicaSet
    └── psql-client-xxxxxxxxxx-xxxxx Pod
```

查看 Deployment、ReplicaSet 和 Pod：
```
kubectl get deployment,replicaset,pod \
  -n postgres-demo \
  -l app=psql-client \
  -o wide
```

pod名称：
```
CLIENT_POD=$(kubectl get pod \
  -n postgres-demo \
  -l app=psql-client \
  -o jsonpath='{.items[0].metadata.name}')

echo "$CLIENT_POD"
```

## SQL Functions

### 1. 连接primary进入数据库：
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
### 2. SELECT：

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
### 3. TABLE：

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

### 4. INSERT
```
INSERT INTO platform_test (message)
VALUES ('Managed PostgreSQL on STACKIT SKE works')
RETURNING *; # return the line after inserting
```


### 5. UPDATE 修改数据
```
UPDATE platform_test
SET message = 'Client successfully connected through the rw Service'
WHERE id = 3
RETURNING *;
```

### 6.  DELETE
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

### 7. 退出SQL
```
\q
```

### 8. 连接Replica 测试INSERT
```
kubectl exec -it \
  -n postgres-demo \
  deployment/psql-client \
  -- env PGHOST=managed-postgres-ro psql
```

```
INSERT INTO platform_test (message)
VALUES ('This write should not succeed');
```

## Reconciliation Pattern

### 1.  Change Desired State

用户修改 Cluster CR
        ↓
API Server 保存 spec.instances=3
        ↓
Operator 收到资源变化
        ↓
发现实际只有一个实例
        ↓
创建两个 Replica Pod
        ↓
创建两个新 PVC
        ↓
复制 Primary 数据
        ↓
status.readyInstances 更新为 3

  
  

### 2.  Delete One DB Pod

手动删除一个 Replica pod
    ↓
实际状态偏离目标状态
    ↓
Operator 再次 Reconcile
    ↓
恢复数据库实例

  
选择Replica：
```
REPLICA=$(kubectl get pods \
  -n postgres-demo \
  -l 'cnpg.io/cluster=managed-postgres,cnpg.io/instanceRole=replica' \
  -o jsonpath='{.items[0].metadata.name}')

echo "Replica selected for deletion: $REPLICA"
```

记录 Pod 和 PVC 的 UID
```
OLD_POD_UID=$(kubectl get pod "$REPLICA" \
  -n postgres-demo \
  -o jsonpath='{.metadata.uid}')

OLD_PVC_UID=$(kubectl get pvc "$REPLICA" \
  -n postgres-demo \
  -o jsonpath='{.metadata.uid}')

echo "Old Pod UID: $OLD_POD_UID"
echo "PVC UID before failure: $OLD_PVC_UID"
```

打开观察窗口：
```
kubectl get cluster,pod,pvc \
  -n postgres-demo \
  -w
```

删除：
```
kubectl delete pod "$REPLICA" \
  -n postgres-demo
```
预期处理是：先将故障实例从 `-r` 和 `-ro` Service 中移除；如果 PVC 仍可用，就使用该 PVC 恢复 Pod；Ready 后重新加入 Service

验证：
```
NEW_POD_UID=$(kubectl get pod "$REPLICA" \
  -n postgres-demo \
  -o jsonpath='{.metadata.uid}')

NEW_PVC_UID=$(kubectl get pvc "$REPLICA" \
  -n postgres-demo \
  -o jsonpath='{.metadata.uid}')

echo "Old Pod UID: $OLD_POD_UID"
echo "New Pod UID: $NEW_POD_UID"

echo
echo "Old PVC UID: $OLD_PVC_UID"
echo "New PVC UID: $NEW_PVC_UID"
```


### 3. 观察产品内部关联

查看 CR、Pod、Service、Secret、PVC、PV 和 Client Deployment 如何彼此关联。

### 4. 删除 Primary，验证自动 failover  

 一个 Replica 被提升为新 Primary，rw Service 自动切换，Client 不需要修改地址。

## CI/CD Automation

CI：检查Merge Request 或 push 是否格式正确、语法正确、计划合理。

执行：
```
terraform fmt -check
terraform init
terraform validate
terraform plan

YAML syntax validation
Kubernetes manifest validation
```




Terraform CD：创建或修改云基础设施

Argo CD：持续部署集群内部产品，管理 Kubernetes manifests

CloudNativePG Operator：管理 PostgreSQL 产品生命周期

### 架构
```
你的 Mac 
├── VS Code 
├── Git 
├── Terraform 
└── kubectl
    │ git push
    ▼
STACKIT Git Repository
    │
    ├── infra/ Desired state changed
    │       ↓
    │   Terraform CI Pipeline
    │       ├── fmt
    │       ├── validate
    │       ├── plan
    │       └── manual apply
	│              ├── SKE Cluster
	│              ├── Node Pool
	│              └── 必要的云基础设施
    │
    └── platform/ Desired state changed
            ↓
        Argo CD 
	        ↓
        Kubernetes manifests 
            ├── compare Git and Cluster state
            ├── sync Kubernetes resources
            └── self-heal configuration drift
					├── CloudNativePG Operator
			        ├── reconcile PostgreSQL Cluster CR
			        └── Client Deployment
		                    ↓
		       Pod / Service / Secret / PVC
```


### Argo CD
控制链路：
```

client-application.yaml
        │
        │ 创建 Argo CD Application
        ▼
Argo CD 读取 Git Desired State
        │
        ▼
找到 platform/client/psql-client-deployment.yaml
        │
        ▼
将 Deployment 同步到 SKE
        │
        ▼
Deployment 创建 Client Pod


Argo CD
    │ read Git
    ▼
platform/ Kubernetes YAML
    │
    ├── Operator
    ├── PostgreSQL CR
    └── Client Deployment
            │
            ▼
        SKE Cluster
```


状态检查
```
kubectl get deployment,statefulset,pod,service \
  -n argocd
```

#### 1. 访问 UI

保持端口转发终端运行：

```
kubectl port-forward \
  service/argocd-server \
  -n argocd \
  8080:443
```

浏览器访问：

```
https://localhost:8080
```

获取初始管理员密码：

```
argocd admin initial-password -n argocd
```

#### 2. Create client pod application
```
kubectl apply \
  -f week2/platform/argocd/client-application.yaml
```

列出 Application：

```
kubectl get applications -n argocd
```

查看应用：

```
argocd app get psql-client
```

#### 3. Sync、Health 和 Diff
check annotation
```
kubectl get deployment psql-client \
  -n postgres-demo \
  -o jsonpath='{.metadata.annotations.argocd\.argoproj\.io/tracking-id}{"\n"}'
```

check diff （manual）
```
argocd app diff psql-client
```
confirm and apply (manual)
```
argocd app sync psql-client
```
手动同步
```
argocd app sync psql-client
```
强制刷新缓存：
```
argocd app get managed-postgres --refresh
```
#### 6. Test auto sync
1. Git-based auto-sync test:
修改 Git 中的 Desired State，改变Client Pod数量
replicas 改回1

2. Self-heal test: 制造手工 Drift，验证 Self Heal
```
kubectl scale deployment psql-client \
  -n postgres-demo \
  --replicas=0
```


这里发生的是两层 reconciliation， 第一层 ArgoCD
```
Argo CD Application Controller发现：
Git replicas=1
Live replicas=0

动作：
把 Deployment spec 恢复为 replicas=1
```

第二层Kubernetes ：

```
Deployment / ReplicaSet Controller发现：
期望 Pod=1
可用 Pod=0

动作：
ReplicaSet Controller 创建新 Pod
        ↓
Scheduler 选择 Node
        ↓
kubelet 创建容器
```

Argo CD 的 `selfHeal` 专门用于这种 Git 没有变化、但集群被手工修改的场景。
### GitOps
#### Create postgreSQL application

强制刷新：
```
argocd app get managed-postgres --refresh
```


### CI - Remote State

1. 防止 state 丢失
2. 避免多人同时 apply


```
week2/infra/*.tf
        │
        ├── 本地 terraform
        └── Forgejo Runner
                 │
                 ▼
      STACKIT Object Storage
      week2/infra/terraform.tfstate
                 │
                 ▼
          STACKIT SKE Cluster
```










Credential
Access key id
YUEBCF1V2XQ873Y9A063

key
ltYrm+621jlU6f4VVs0OYGJ3Swk+T8QVYdbtiI9B
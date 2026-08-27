# Cloud-Native Platform Engineering on STACKIT

A five-week hands-on cloud engineering project covering the path from infrastructure provisioning to a Kubernetes-based platform, REST API development, production security, and observability.

The project evolved through five layers:

```
OpenStack / Terraform
        ↓
STACKIT Kubernetes Engine
        ↓
PostgreSQL + Go REST API
        ↓
Authentication / RBAC / HTTPS
        ↓
Prometheus / Grafana / Operations
```

## Project Overview

The goal of this project was to understand and implement the major layers of a modern cloud-native platform.

It started with infrastructure provisioning on OpenStack and gradually moved toward a production-oriented Kubernetes environment with automated deployment, application APIs, security controls, monitoring, and troubleshooting.

### Core Technologies

- **Cloud Infrastructure:** OpenStack, Terraform
- **Container Platform:** STACKIT Kubernetes Engine, Kubernetes, Docker
- **Database:** PostgreSQL, CloudNativePG
- **Backend:** Go, REST APIs, OpenAPI
- **Delivery:** GitLab CI/CD, Argo CD, GitOps
- **Security:** JWT, RBAC, ownership-based authorization, HTTPS/TLS
- **Observability:** Structured Logging, Prometheus, Grafana
- **Operations:** HPA, load testing, Kubernetes troubleshooting

---

# Architecture Evolution

## Week 1 — Infrastructure as a Service

The first stage focused on the infrastructure layer using OpenStack.
```
OpenStack / IaaS
│
├── Network
├── Subnet
├── Router
├── Security Group
├── Virtual Machine
├── Floating IP
└── Terraform
```
Cloud resources were first explored as individual infrastructure components and then provisioned using Terraform as Infrastructure as Code.

This stage covered how networking and compute resources are connected:
```
External Network
      │
 Floating IP
      │
    Router
      │
    Subnet
      │
   Network
      │
      VM
```

### Key topics

- OpenStack networking
- Virtual machines
- Security groups
- Floating IPs
- Infrastructure as Code
- Terraform-based provisioning

---

## Week 2 — Kubernetes Platform and Managed PostgreSQL

The second stage moved from virtual infrastructure to a Kubernetes-based platform on STACKIT Kubernetes Engine.
```
STACKIT Kubernetes Engine
│
├── SKE Cluster
├── Worker Nodes
│
├── CloudNativePG Operator
│   ├── CRD
│   └── PostgreSQL Cluster CR
│
├── PostgreSQL
│   ├── Primary
│   └── Replicas
│
├── Persistent Volumes
└── Argo CD / GitOps
```

A PostgreSQL service was deployed using the CloudNativePG Operator.

Instead of manually managing database processes, the database lifecycle was described through Kubernetes custom resources and managed by the operator.

Persistent storage was attached through Kubernetes PVCs, while Argo CD was introduced for GitOps-based deployment.

### Key topics

- Kubernetes clusters and worker nodes
- Operators
- Custom Resource Definitions
- Custom Resources
- CloudNativePG
- PostgreSQL high availability
- Persistent storage
- GitOps
- Argo CD

---

## Week 3 — REST API and Platform Service Layer

The third stage introduced an application layer on top of Kubernetes.
```
Client
  │
  │ HTTP / REST
  ▼
Go REST API
  │
  ├── CRUD Endpoints
  ├── OpenAPI
  ├── Unit Tests
  │
  └── Kubernetes Client
          │
          ▼
      Kubernetes API
```

A REST API was developed in Go to expose platform operations through HTTP endpoints.

The API interacted with Kubernetes resources through the Kubernetes client rather than requiring users to manage resources directly with `kubectl`.

The application was containerized and deployed to SKE.
```
Source Code
    │
    ▼
GitLab CI/CD
    │
    ▼
Docker Image
    │
    ▼
Container Registry
    │
    ▼
Kubernetes Deployment
```

### Key topics

- Go REST API
- CRUD operations
- Kubernetes API integration
- OpenAPI specification
- Unit testing
- Docker
- Container registry
- Kubernetes ServiceAccount and RBAC
- GitLab CI/CD
- Horizontal Pod Autoscaling
- Load testing

---

## Week 4 — Production Access and Security

The fourth stage focused on exposing the application securely and introducing user-level access control.
```
User
 │
 │ HTTPS
 ▼
DNS
 │
 ▼
Load Balancer / Ingress
 │
 ▼
Kubernetes Service
 │
 ▼
Go API
 │
 ├── Authentication
 ├── JWT
 ├── Authorization
 ├── User / Admin Roles
 └── Resource Ownership
```
Authentication and authorization were added to the API using JWT-based access control.

Different permissions were introduced for users and administrators, and resources were associated with their owners to enforce resource-level isolation.

The application was exposed externally through Kubernetes networking and secured with HTTPS/TLS.

A web UI was also connected to the backend API.

### Key topics

- Authentication
- JWT
- Authorization
- User and admin roles
- Resource ownership
- Kubernetes RBAC
- DNS
- Load balancing
- Ingress
- Kubernetes Services
- HTTPS / TLS
- Web UI

---

## Week 5 — Observability and Operations

The final stage focused on understanding the behavior of the running platform.
```
Application
   │
   ├── Structured Logs
   │       └── Request IDs
   │
   └── /metrics
           │
           ▼
       Prometheus
           │
           ▼
        Grafana
```
Structured application logging was introduced together with request IDs to make individual requests easier to trace.

The Go API exposes Prometheus-compatible metrics through a `/metrics` endpoint.

Prometheus collects the metrics and Grafana is used to visualize application behavior.

Application metrics include areas such as:

- HTTP request traffic
- Request latency
- Instance operations
- Platform activity

Metrics and logs can then be combined with Kubernetes tooling during troubleshooting.

### Key topics

- Structured logging
- Request IDs
- Application metrics
- Prometheus `/metrics`
- Prometheus
- Grafana
- Monitoring
- Kubernetes troubleshooting
- HPA
- Load testing

---

# End-to-End Platform Flow

At the end of the project, the platform combined the different layers into one system

The infrastructure underneath the platform is managed through Terraform and Kubernetes manifests, while deployment is automated through CI/CD and GitOps.

---

# Engineering Topics Covered

This project provided hands-on experience across several areas of cloud platform engineering:

### Infrastructure

- OpenStack
- Cloud networking
- Terraform
- Infrastructure as Code

### Kubernetes

- Deployments
- Services
- Ingress
- RBAC
- ServiceAccounts
- Persistent storage
- Custom Resources
- Operators
- HPA

### Backend

- Go
- REST APIs
- OpenAPI
- Kubernetes client
- Unit testing

### Security

- Authentication
- JWT
- Role-based authorization
- Resource ownership
- HTTPS/TLS

### Delivery

- Docker
- Container registry
- GitLab CI/CD
- GitOps
- Argo CD

### Observability

- Structured logging
- Request IDs
- Prometheus metrics
- Grafana dashboards
- Monitoring and troubleshooting

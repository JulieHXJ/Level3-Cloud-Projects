# Week 1 Notes: OpenStack, Terraform and Kubernetes Fundamentals

## 1. Main Focus of the Week

The three main technologies introduced during the first week have different responsibilities:

* **OpenStack manages infrastructure**, including virtual machines, networks, storage and IP addresses.
* **Terraform manages infrastructure configuration** by describing the desired resources in code.
* **Kubernetes manages containerized applications** running inside virtual machines or physical servers.

The complete technology stack can be understood as follows:

```text
Physical servers
        ↓
OpenStack manages infrastructure
        ↓
Virtual machines, networks and storage
        ↓
Kubernetes manages workloads
        ↓
Containers and applications
```

Each layer solves a different problem. OpenStack provides computing infrastructure, Terraform automates its creation, and Kubernetes manages applications running on that infrastructure.

---

# 2. Cloud Computing and Service Models

## 2.1 What Is Cloud Computing?

Cloud computing means providing computing resources as services that users can request when needed.

Instead of manually preparing a physical server, installing an operating system, configuring a network and allocating storage, users can request resources through:

* A web portal
* A command-line interface
* An API
* Infrastructure-as-Code tools such as Terraform

For example, a user can request:

> Create an Ubuntu virtual machine with 4 virtual CPUs and 8 GB of memory.

The cloud platform then finds available resources, creates the virtual machine, connects it to a network and makes it available to the user.

---

## 2.2 IaaS: Infrastructure as a Service

**IaaS** stands for **Infrastructure as a Service**.

It turns infrastructure resources into services that can be requested through a portal or API.

Typical IaaS resources include:

* Virtual machines
* Virtual networks
* Virtual disks
* Public and private IP addresses
* Routers
* Firewalls and security rules

OpenStack mainly provides IaaS capabilities.

With IaaS, the cloud platform manages the physical infrastructure and virtualization layer. The user is usually responsible for:

* Choosing the operating system
* Installing software
* Configuring applications
* Maintaining the virtual machine

---

## 2.3 PaaS: Platform as a Service

**PaaS** stands for **Platform as a Service**.

PaaS provides a higher-level application platform. Users deploy applications without managing every virtual machine directly.

A PaaS platform may provide:

* Application runtimes
* Managed databases
* Deployment systems
* Automatic scaling
* Logging and monitoring
* Application networking

Kubernetes is not exactly a complete PaaS by itself, but it is commonly used as the foundation of a PaaS.

The relationship can be simplified as:

```text
IaaS → Provides virtual machines, networks and storage

PaaS → Provides an environment for deploying applications

SaaS → Provides a complete application to the end user
```

---

# 3. Physical Machines, Virtual Machines and Containers

## 3.1 Physical Machine

A physical machine is an actual computer with:

* Physical CPU cores
* Physical memory
* Physical disks
* Physical network interfaces

Without virtualization, one operating system normally controls the entire machine.

---

## 3.2 Virtual Machine

A virtual machine is a software-defined computer running on a physical machine.

Each virtual machine has virtual resources such as:

* Virtual CPUs
* Virtual memory
* Virtual disks
* Virtual network interfaces
* Its own operating system

Several virtual machines can run on the same physical server.

For example:

```text
Physical Server
├── VM 1: Ubuntu
├── VM 2: Debian
└── VM 3: Rocky Linux
```

Each virtual machine behaves like an independent computer.

---

## 3.3 Container

A container packages an application together with its dependencies.

Unlike a virtual machine, a container usually does not contain a complete operating system kernel. Containers share the host operating system kernel.

```text
Virtual Machine
├── Container 1: API application
├── Container 2: Database
└── Container 3: Frontend
```

Containers are generally:

* Smaller than virtual machines
* Faster to start
* Easier to reproduce
* Easier to move between environments

However, containers still need a machine on which to run. That machine may be a physical server or a virtual machine.

---

# 4. Hypervisors and Virtualization

A **hypervisor** is the software layer that creates and runs virtual machines.

It divides the resources of a physical machine into virtual resources.

For example, a physical server may have:

* 32 CPU cores
* 128 GB RAM
* 2 TB storage

The hypervisor can divide these resources among several virtual machines.

```text
Physical Server
        ↓
Hypervisor
        ↓
VM 1    VM 2    VM 3
```

The hypervisor provides each VM with virtual hardware, including:

* Virtual CPU
* Virtual memory
* Virtual disk
* Virtual network interface

In an OpenStack environment, Nova works together with a virtualization system such as **KVM/QEMU** through **libvirt**.

OpenStack does not normally emulate the CPU itself. Instead, it coordinates the underlying hypervisor and tells it when to create, start, stop or delete virtual machines.

---

# 5. OpenStack

## 5.1 What Is OpenStack?

OpenStack is an open-source software platform for building and managing cloud infrastructure.

It manages resources across one or many physical servers and allows users to create:

* Virtual machines
* Virtual networks
* Virtual disks
* IP addresses
* Routers
* Firewall rules

Users can access OpenStack through:

* A web dashboard
* The OpenStack CLI
* REST APIs
* Automation tools such as Terraform

OpenStack solves the problem of manually performing tasks such as:

* Creating a virtual machine
* Allocating CPU and memory
* Installing an operating system
* Assigning an IP address
* Connecting the machine to a network
* Configuring storage
* Applying firewall rules

OpenStack performs three major functions:

### Resource creation

OpenStack creates resources such as:

* Compute instances
* Networks
* Subnets
* Routers
* Ports
* Volumes
* Floating IP addresses

### Resource management

OpenStack manages:

* CPU capacity
* Memory
* Storage
* Network connectivity
* VM lifecycle

### Resource scheduling

OpenStack selects a suitable physical compute node on which a virtual machine can run.

---

# 6. Main OpenStack Components

OpenStack is not a single program. It consists of multiple services that communicate through APIs.

## 6.1 Keystone: Identity and Access Management

Keystone handles identity, authentication and authorization.

Its main responsibilities include:

* **Authentication:** Verifying who the user is
* **Authorization:** Determining what the user is allowed to do
* **Tokens:** Providing temporary credentials for API requests
* **Service Catalog:** Providing the API addresses of OpenStack services

A typical authentication process is:

```text
User credentials
        ↓
Keystone verifies the identity
        ↓
Keystone returns a token
        ↓
The client uses the token for later API requests
```

The token proves that the user has already authenticated.

---

## 6.2 Glance: Image Service

Glance stores and manages operating system images.

An image is a template that can be used to create a virtual machine.

Images may contain:

* Ubuntu
* Debian
* Rocky Linux
* CirrOS
* Other preconfigured operating systems

When a VM is created, Nova obtains the selected image from Glance.

---

## 6.3 Nova: Compute Service

Nova manages the lifecycle of compute instances.

It is responsible for operations such as:

* Creating a VM
* Starting a VM
* Stopping a VM
* Rebooting a VM
* Resizing a VM
* Deleting a VM

Nova answers questions such as:

* Where should the VM be created?
* How much CPU and memory does it need?
* Which compute node should run it?
* How should it be started or stopped?

Nova cooperates with:

* Keystone for authentication
* Glance for images
* Placement for available resources
* Neutron for networking
* Cinder for block storage

---

## 6.4 Placement: Resource Inventory

Placement records available infrastructure resources.

It helps Nova understand:

* Which compute nodes exist
* How many CPUs are available
* How much memory is available
* Which resources are already allocated
* Which node has sufficient capacity for a new VM

Nova Scheduler uses this information when selecting a compute host.

Placement does not normally start the VM itself. It provides resource inventory and allocation information.

---

## 6.5 Neutron: Networking Service

Neutron provides virtual networking.

It manages resources such as:

* Networks
* Subnets
* Ports
* Routers
* Floating IP addresses
* Security groups

Neutron answers questions such as:

* Which network should the VM join?
* Which private IP should it receive?
* What is the subnet?
* Which gateway should be used?
* How should the router be configured?
* Which floating IP should be associated?
* Which traffic should the security group allow?

---

## 6.6 Cinder: Block Storage

Cinder provides block storage volumes.

A Cinder volume is similar to an independent cloud disk. It can be:

* Created separately from a VM
* Attached to a VM
* Detached from a VM
* Reattached to another VM
* Managed independently

A Cinder volume should be distinguished from the VM's root disk or temporary disk.

A simplified comparison is:

```text
Root disk:
Usually created together with the VM and contains the operating system.

Cinder volume:
An independently managed virtual disk that can be attached to a VM.
```

---

## 6.7 Horizon: Web Dashboard

Horizon is the OpenStack web dashboard.

It allows users to manage OpenStack resources through a browser.

Through Horizon, users can:

* View and create virtual machines
* Create networks and subnets
* Manage volumes
* Allocate IP addresses
* View images
* Configure security groups

Horizon does not directly perform all infrastructure operations itself. It sends API requests to the relevant OpenStack services.

---

# 7. DevStack

## 7.1 What Is DevStack?

DevStack is a collection of scripts that installs an OpenStack development environment.

OpenStack is a large platform consisting of many services. A production OpenStack cloud may run across many physical servers.

For training and development, preparing a full production environment would be impractical. DevStack allows many OpenStack services to be installed on one Linux machine or a small number of machines.

A typical DevStack environment may include:

* Keystone
* Glance
* Nova
* Placement
* Neutron
* Cinder
* Horizon

After installation, the environment can be managed through:

* Horizon
* OpenStack CLI commands
* OpenStack APIs
* Terraform

---

## 7.2 Why DevStack Is Only Suitable for Development and Testing

DevStack is designed to create a working OpenStack environment quickly.

It usually places many services on a single machine and uses configurations intended for:

* Learning
* Development
* Testing
* Debugging
* Feature demonstrations

A production environment must also consider:

* High availability
* Multi-node redundancy
* Security hardening
* Secret and key management
* Backups
* Upgrade strategies
* Monitoring and alerting
* Capacity planning
* Fault domains
* Performance
* Continuous operations

DevStack focuses mainly on questions such as:

* Can OpenStack be installed quickly?
* Can developers test and debug OpenStack?
* Can OpenStack functionality be demonstrated?

DevStack is not a fake version of OpenStack. It runs real OpenStack components. However, its deployment architecture, configuration, reliability and operational model are not appropriate for production use.

---

# 8. API, CLI and Web Dashboard

OpenStack resources can be managed through different interfaces.

## 8.1 API

The API is the lowest-level programmatic interface.

Applications send HTTP requests to OpenStack service endpoints.

For example:

```text
POST request → Create a server
GET request  → Retrieve server information
DELETE request → Delete a server
```

---

## 8.2 CLI

CLI stands for **Command-Line Interface**.

The OpenStack CLI converts commands into API requests.

For example:

```bash
openstack server list
```

The CLI sends an API request to Nova and displays the response in the terminal.

---

## 8.3 Web Dashboard

A web dashboard such as Horizon or the STACKIT Portal provides a graphical interface.

When a user clicks a button such as **Create Server**, the dashboard sends API requests in the background.

The relationship is:

```text
User
├── Web Dashboard
├── CLI
├── Terraform
└── Custom Application
          ↓
       Cloud APIs
          ↓
OpenStack or another cloud platform
```

The interfaces are different, but they ultimately communicate with the same cloud APIs.

---

# 9. Network Fundamentals

Cloud infrastructure requires an understanding of basic networking concepts.

Important terms include:

* IP address
* Private IP
* Public IP
* Subnet
* Gateway
* Port
* Network interface
* DNS
* Firewall
* Security group
* NAT
* SSH

---

## 9.1 IP Address

An IP address identifies a network interface.

When a machine sends a packet, the packet normally contains:

* **Source IP:** Where the packet came from
* **Destination IP:** Where the packet should go

For example:

```text
Source IP:      192.168.1.10
Destination IP: 8.8.8.8
```

Routers use the destination IP to determine where the packet should be forwarded.

---

## 9.2 Private IP

A private IP is normally used inside an internal network.

Private IP ranges are not directly routed over the public internet.

Examples include:

```text
10.0.0.0/8
172.16.0.0/12
192.168.0.0/16
```

In the STACKIT environment, the server received the private IP:

```text
10.0.0.61
```

This address was allocated from the private network:

```text
10.0.0.0/25
```

---

## 9.3 Public IP

A public IP is reachable through the public internet, provided that routing and firewall rules allow the connection.

The server received the public IP:

```text
188.34.105.55
```

The cloud platform maps the public IP to the private network interface of the server:

```text
Public IP: 188.34.105.55
              ↓
Private IP: 10.0.0.61
```

The Mac is not part of the STACKIT private network, so it cannot directly reach `10.0.0.61`.

Instead, it connects to the server through the public IP:

```bash
ssh ubuntu@188.34.105.55
```

---

## 9.4 Subnet

A subnet defines a range of IP addresses that belong to the same logical network.

The subnet:

```text
10.0.0.0/25
```

contains addresses from approximately:

```text
10.0.0.0 to 10.0.0.127
```

Some addresses may be reserved for:

* The network address
* The gateway
* Broadcast or platform purposes
* DHCP or infrastructure services

The subnet tells the cloud platform which IP addresses it may assign to resources.

---

## 9.5 Gateway

A gateway connects a local network to other networks.

When a VM wants to communicate with a destination outside its own subnet, it normally sends the packet to its default gateway.

```text
VM
 ↓
Subnet gateway
 ↓
Router
 ↓
External network or internet
```

---

## 9.6 Port

The word **port** can have two different meanings.

### Network service port

A TCP or UDP port identifies a service running on a machine.

Examples:

* SSH: TCP port 22
* HTTP: TCP port 80
* HTTPS: TCP port 443
* Kubernetes API: TCP port 6443

An IP address identifies the machine or network interface. A port number identifies the service on that machine.

```text
188.34.105.55:22
```

This means TCP port 22 on the machine reachable through `188.34.105.55`.

### OpenStack or cloud network port

In cloud networking, a port can also mean a virtual connection point between a VM's virtual network interface and a virtual network.

Such a port may contain:

* A MAC address
* A private IP address
* A network ID
* Security group associations
* Device ownership information

These two meanings should not be confused.

---

## 9.7 NIC

NIC stands for **Network Interface Card**.

A physical computer requires a network interface to connect to a network. A virtual machine also requires a virtual network interface.

The NIC is the connection point for several cloud networking settings:

* Which network the server joins
* Which private IP the server receives
* Where the public IP is mapped
* Which security groups apply to the server

A simplified model is:

```text
Server
  ↓
Virtual NIC
  ↓
Private network
```

---

## 9.8 NAT

NAT stands for **Network Address Translation**.

NAT changes IP address information as packets pass through a router or gateway.

It is commonly used to allow a machine with a private IP to communicate with the public internet.

For outbound traffic:

```text
VM private IP
      ↓
NAT or SNAT
      ↓
Public or external address
      ↓
Internet
```

For inbound access through a public IP, the cloud platform can map the public address to the private IP of the VM.

---

## 9.9 DNS

DNS stands for **Domain Name System**.

DNS translates domain names into IP addresses.

For example:

```text
openstack.org
        ↓
DNS resolution
        ↓
An IP address
```

A machine may be able to reach `1.1.1.1` but still fail to reach `openstack.org` when DNS configuration is broken.

---

## 9.10 Firewall and Security Group

A firewall controls which network traffic is allowed or blocked.

A security group is a cloud-level virtual firewall.

Having a public IP does not mean that every service should be accessible from the internet.

For the DevStack server, the security group contained an ingress rule similar to:

```text
Direction:  Ingress
IP version: IPv4
Source:     188.244.102.157/32
Protocol:   TCP
Port:       22
```

This means:

> Allow TCP traffic from the public IP `188.244.102.157` to reach TCP port 22 on the server.

The `/32` suffix represents one specific IPv4 address.

Security group rules can control:

* Source IP ranges
* Destination ports
* Protocols
* Inbound traffic
* Outbound traffic

---

## 9.11 SSH

SSH stands for **Secure Shell**.

It provides encrypted remote access to a machine.

SSH normally listens on TCP port 22.

A connection command may look like:

```bash
ssh ubuntu@188.34.105.55
```

This command means:

* Connect using the SSH protocol
* Use the username `ubuntu`
* Connect to the host `188.34.105.55`
* Use TCP port 22 unless another port is specified

SSH requires all of the following:

1. The destination IP must be reachable.
2. Routing must work.
3. The security group must allow TCP port 22.
4. The operating system firewall must allow the connection.
5. The SSH service must be running.
6. Authentication must succeed.

---

## 9.12 Testing a TCP Port

The following command tests whether TCP port 22 is reachable:

```bash
nc -vz -w 10 188.34.105.55 22
```

Meaning:

* `nc`: Run Netcat
* `-v`: Display verbose output
* `-z`: Test the port without sending normal application data
* `-w 10`: Wait for a maximum of 10 seconds
* `188.34.105.55`: Destination IP
* `22`: Destination TCP port

A successful result confirms that a TCP connection to the port can be established. It does not by itself prove that SSH authentication will succeed.

---

# 10. Why a Service May Be Running but Still Be Unreachable

A service may be active inside a machine but inaccessible from another machine.

Possible reasons include:

* The service listens only on `127.0.0.1`
* The service listens on the wrong network interface
* The cloud security group blocks the port
* The operating system firewall blocks the port
* The service uses a different port
* The public IP is not correctly associated
* Routing is missing
* NAT is misconfigured
* DNS resolution is broken
* The application is unhealthy
* The service process has stopped

Many problems that appear to be Kubernetes or application problems are actually caused by:

* Ports
* Firewalls
* DNS
* Routing
* Network interfaces
* Incorrect IP addresses

Troubleshooting should therefore be performed layer by layer.

---

# 11. STACKIT Server Network Setup

The STACKIT server used for DevStack required several cloud resources.

## 11.1 Private Network

A private network named `julie_network` was created with the address range:

```text
10.0.0.0/25
```

The server received the private IP:

```text
10.0.0.61
```

This allowed the server to communicate within the STACKIT private network.

---

## 11.2 Public IP

Because the local Mac was not part of the STACKIT private network, it could not directly access `10.0.0.61`.

A public IP was therefore assigned:

```text
188.34.105.55
```

This public address was mapped to the server's private network interface.

---

## 11.3 Network Interface

The server's NIC connected the server to `julie_network`.

The NIC was associated with:

* The private network
* The private IP
* The public IP mapping
* One or more security groups

---

## 11.4 Security Group

A security group named `julie-devstack-ssh` allowed SSH traffic from the home public IP.

This restricted SSH access instead of exposing port 22 to the entire internet.

---

## 11.5 Connection Test

The TCP connection to SSH was tested with:

```bash
nc -vz -w 10 188.34.105.55 22
```

After confirming that port 22 was reachable, the server could be accessed with SSH.

---

# 12. How OpenStack Creates a Virtual Machine

A simplified VM creation process is:

## Step 1: Authentication

The user logs in through Horizon, the CLI or Terraform.

Keystone verifies the credentials and returns a token.

```text
User
  ↓
Keystone
  ↓
Authentication token
```

## Step 2: Server Request

The user requests a new VM and specifies information such as:

* Image
* Flavor
* Network
* Key pair
* Security group
* Storage

Nova API receives the request.

## Step 3: Resource Check

Nova asks Placement which compute nodes have enough available:

* CPU
* Memory
* Disk
* Other required resources

## Step 4: Scheduling

Nova Scheduler selects a suitable compute node.

## Step 5: Image Retrieval

Nova obtains the selected operating system image from Glance.

## Step 6: Network Preparation

Neutron creates or prepares a virtual network port.

The port may include:

* Private IP
* MAC address
* Network connection
* Security groups

## Step 7: Storage Preparation

The VM may use:

* A root disk created from the image
* An attached Cinder volume
* Both

## Step 8: VM Creation

Nova Compute instructs the hypervisor to create and start the VM.

## Step 9: Network Connection

The virtual NIC is connected to the Neutron network.

The VM can then obtain its configured private IP.

## Step 10: Public Access

A floating or public IP may be associated with the private IP.

Security group rules determine which inbound traffic is allowed.

The overall flow is:

```text
User
  ↓
Horizon, CLI or Terraform
  ↓
Keystone authentication
  ↓
Nova receives the VM request
  ↓
Placement reports available resources
  ↓
Nova Scheduler selects a compute node
  ↓
Glance provides the image
  ↓
Neutron prepares networking
  ↓
Cinder optionally provides storage
  ↓
Hypervisor starts the VM
```

---

# 13. Terraform

## 13.1 What Is Terraform?

Terraform is an Infrastructure-as-Code tool.

It allows infrastructure to be described in configuration files rather than created manually through a web interface.

Terraform can manage resources such as:

* Virtual machines
* Networks
* Subnets
* Routers
* Security groups
* Public IP addresses
* Kubernetes clusters

Advantages include:

* Repeatable infrastructure creation
* Reduced manual work
* Fewer configuration mistakes
* Version-controlled infrastructure
* Easier review of planned changes
* Easier deletion and recreation
* Consistent environments

Terraform does not replace OpenStack. It sends API requests to OpenStack on the user's behalf.

```text
Terraform configuration
          ↓
Terraform OpenStack provider
          ↓
OpenStack APIs
          ↓
OpenStack resources
```

---

## 13.2 Provider

A provider allows Terraform to communicate with an external platform.

For example, the OpenStack provider knows how to manage OpenStack resources.

```hcl
terraform {
  required_providers {
    openstack = {
      source = "terraform-provider-openstack/openstack"
    }
  }
}
```

The provider translates Terraform operations into OpenStack API requests.

---

## 13.3 Resource

A resource describes an infrastructure object that Terraform should create or manage.

For example:

```hcl
resource "openstack_compute_instance_v2" "vm" {
  name = "example-vm"
}
```

This resource represents an OpenStack compute instance.

---

## 13.4 Variable

A variable allows configuration values to be provided from outside the main Terraform code.

For example:

```hcl
variable "instance_name" {
  type = string
}
```

Variables make the configuration reusable.

Different values can be supplied without rewriting the resource definition.

---

## 13.5 Output

An output displays useful information after Terraform has created the infrastructure.

For example:

```hcl
output "server_ip" {
  value = openstack_compute_instance_v2.vm.access_ip_v4
}
```

Outputs may display:

* VM names
* Private IP addresses
* Public IP addresses
* Resource IDs
* URLs

---

## 13.6 State

Terraform state records the relationship between the Terraform configuration and the real infrastructure.

The state allows Terraform to understand:

* Which resources it created
* Which Terraform resource corresponds to which cloud resource
* The current known attributes of the resources
* Which changes are required

Terraform normally stores this information in:

```text
terraform.tfstate
```

The state file is important and may contain sensitive information. It should not be carelessly shared or edited manually.

---

# 14. Terraform Workflow

## 14.1 `terraform init`

```bash
terraform init
```

Initializes the Terraform working directory.

It downloads required providers and prepares the project.

---

## 14.2 `terraform fmt`

```bash
terraform fmt
```

Formats Terraform configuration files into a consistent style.

It does not create infrastructure.

---

## 14.3 `terraform validate`

```bash
terraform validate
```

Checks whether the Terraform configuration is structurally and syntactically valid.

It does not prove that the cloud platform will accept every requested resource.

---

## 14.4 `terraform plan`

```bash
terraform plan
```

Compares:

* The Terraform configuration
* The Terraform state
* The current infrastructure

It then displays the proposed changes.

Possible symbols include:

```text
+ create
~ update
- destroy
-/+ replace
```

The plan allows the user to review changes before applying them.

---

## 14.5 `terraform apply`

```bash
terraform apply
```

Executes the planned infrastructure changes.

Terraform sends API requests to the provider and creates, updates or deletes resources.

---

## 14.6 `terraform destroy`

```bash
terraform destroy
```

Deletes the infrastructure resources managed by the Terraform configuration.

This command should be reviewed carefully because it may remove virtual machines, networks, volumes and other resources.

---

# 15. Kubernetes

## 15.1 What Is Kubernetes?

Kubernetes is a platform for managing containerized applications.

OpenStack manages infrastructure resources such as:

* Virtual machines
* Virtual networks
* Virtual disks
* IP addresses

Kubernetes manages application resources such as:

* Pods
* Containers
* Deployments
* Services
* ConfigMaps
* Secrets
* Persistent volumes

Kubernetes normally runs on one or more machines called nodes. These nodes may be:

* Physical machines
* Virtual machines
* Cloud instances

---

## 15.2 Problems Solved by Kubernetes

Without Kubernetes, application operators must manually handle questions such as:

* Who restarts an application after it crashes?
* How can three copies of an API be run?
* How should a new application version be deployed?
* How can services discover each other?
* On which VM should a container run?
* How can VM resources be used efficiently?
* How should failed containers be replaced?
* How can traffic be distributed between replicas?

Kubernetes provides mechanisms for:

* Automatic restart
* Desired-state management
* Scheduling
* Scaling
* Service discovery
* Load balancing
* Rolling updates
* Self-healing
* Configuration management

---

## 15.3 Kubernetes Desired State

The user describes the desired application state.

For example:

```text
Run three replicas of the API application.
```

Kubernetes continuously compares:

```text
Desired state
      versus
Actual state
```

When the actual state differs, Kubernetes attempts to correct it.

For example, if one of three Pods crashes, Kubernetes creates a replacement Pod.

---

# 16. Relationship Between OpenStack, Terraform and Kubernetes

The three technologies should not be treated as competitors. They work at different layers.

```text
Physical infrastructure
        ↓
OpenStack
Creates and manages VMs, networks and storage
        ↓
Terraform
Automates the declaration and creation of infrastructure
        ↓
Virtual machines
        ↓
Kubernetes
Schedules and manages containerized applications
        ↓
Application containers
```

More precisely, Terraform can manage both OpenStack and Kubernetes resources. However, during this course, its primary role is to automate infrastructure creation.

A typical workflow is:

1. Terraform sends requests to the OpenStack APIs.
2. OpenStack creates virtual machines, networks and storage.
3. Kubernetes is installed on the virtual machines.
4. Kubernetes deploys and manages applications.

---

# 17. Key Week 1 Knowledge

By the end of the first week, the following concepts should be explainable.

## OpenStack

* OpenStack is an IaaS cloud platform.
* Keystone manages identity and tokens.
* Glance manages images.
* Nova manages VM lifecycle.
* Placement records resource inventory.
* Neutron manages virtual networking.
* Cinder manages block storage.
* Horizon provides a web dashboard.
* OpenStack services communicate through APIs.

## Networking

* A private IP is used inside a private network.
* A public IP provides external reachability.
* A subnet defines an IP address range.
* A gateway forwards traffic to other networks.
* A NIC connects a machine to a network.
* NAT translates between private and external addresses.
* DNS translates names into IP addresses.
* Security groups control allowed network traffic.
* SSH usually uses TCP port 22.

## Terraform

* A provider connects Terraform to a platform.
* A resource represents an infrastructure object.
* A variable makes the configuration reusable.
* An output displays useful values.
* State tracks Terraform-managed resources.
* `init`, `fmt`, `validate`, `plan`, `apply` and `destroy` represent the main Terraform workflow.

## Kubernetes

* Kubernetes manages containerized applications.
* A Kubernetes cluster runs on physical or virtual machines.
* Kubernetes maintains a desired application state.
* Kubernetes provides scheduling, scaling, service discovery, updates and self-healing.

---

# 18. Complete Example

A complete infrastructure and application deployment may follow this process:

```text
1. A user writes Terraform configuration.

2. Terraform authenticates with the cloud platform.

3. Terraform sends API requests to OpenStack.

4. OpenStack creates:
   - A network
   - A subnet
   - A router
   - A security group
   - A virtual machine
   - A public IP

5. Neutron connects the VM to the network.

6. Nova starts the VM using an image from Glance.

7. The user accesses the VM through SSH.

8. Kubernetes is installed on the VM.

9. Kubernetes starts containerized applications.

10. Kubernetes continuously maintains the desired application state.
```

This represents the main architecture studied during the first week:

```text
User or Terraform
        ↓
Cloud API
        ↓
OpenStack
        ↓
Virtual infrastructure
        ↓
Kubernetes
        ↓
Applications
```

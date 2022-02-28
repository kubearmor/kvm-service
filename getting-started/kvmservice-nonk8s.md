# Controlling Kubearmor on VM using non-k8s control plane

Kubearmor is a runtime security engine that protects the host/VM from unknown threats. 

With Kubearmor running on a VM, it is possible to enforce host based security policies and secure the VM at system level.

## Why control Kubearmor on VM ?
With Kubearmor running on multiple VMs within the same network, it is difficult and time consuming to enforce policies on all VMs one-by-one. Hence the solution is to manage all VMs in the network from one control plane.

### KVMService as a control plane
KVMService(Kubearmor VirtualMachine Service) allows Kubearmor in VMs within the same network to connect to it and manages the policy enforcement to all VMs.

Kvmservice acts as the control plane for VMs for non-k8s(stand-lone) environment.

## Design of kvmservice as non-k8s control plane
![Alt Text](./res/kvmservice-non-k8s-control-plane.png)

### Components Involved and it's use
* **Non-K8s Control Plane**
    * **Etcd** : Etcd is used as a key:value storage and is used to store the label information and unique identity of each configured VM.
    * **KvmService** : Manages connection with VM, handles VM onboarding/offboarding, label management and policy enforcement.
    * **Karmor** (Support utility) : A CLI utility which interacts with kvmservice for VM onbaoarding/offboarding, policy enforcement and label management.
* **VMs** (Actual VMs connected in the network)

## Installation Guide

Below covered are the E2E steps to onboard a VM and enforce policy in that VM using non-k8s control plane. 
1. Install and run kvmservice and dependencies on a non-k8s control plane(VM or standalone linux PC)
2. Onboard a new VM
3. Download installation script
4. Run kubearmor in VM
5. Enforce policies from control plane onto VM
6. Manage labels

*Note : All the steps are tested and carried out in debian based OS distribution.*

### Step 1: Install ETCD service in control plane
Install etcd using below command
```
$ sudo apt-get install etcd
Reading package lists... Done
Building dependency tree       
Reading state information... Done
The following NEW packages will be installed:
  etcd
0 upgraded, 1 newly installed, 0 to remove and 8 not upgraded.
Need to get 2,520 B of archives.
After this operation, 16.4 kB of additional disk space will be used.
Get:1 http://in.archive.ubuntu.com/ubuntu focal/universe amd64 etcd all 3.2.26+dfsg-6 [2,520 B]
Fetched 2,520 B in 0s (9,080 B/s)
Selecting previously unselected package etcd.
(Reading database ... 246471 files and directories currently installed.)
Preparing to unpack .../etcd_3.2.26+dfsg-6_all.deb ...
Unpacking etcd (3.2.26+dfsg-6) ...
Setting up etcd (3.2.26+dfsg-6) ...
$ 
```
Once etcd is installed you can check etcd installation status using ***sudo service etcd status*** command
```
$ sudo service etcd status
● etcd.service - etcd - highly-available key value store
     Loaded: loaded (/lib/systemd/system/etcd.service; enabled; vendor preset: enabled)
     Active: active (running) since Sun 2022-01-16 12:23:58 IST; 30min ago
       Docs: https://github.com/coreos/etcd
             man:etcd
   Main PID: 1087 (etcd)
      Tasks: 24 (limit: 18968)
     Memory: 84.2M
     CGroup: /system.slice/etcd.service
             └─1087 /usr/bin/etcd

Jan 16 12:23:57 LEGION etcd[1087]: 8e9e05c52164694d as single-node; fast-forwarding 9 ticks (election ticks 10)
Jan 16 12:23:58 LEGION etcd[1087]: 8e9e05c52164694d is starting a new election at term 88
Jan 16 12:23:58 LEGION etcd[1087]: 8e9e05c52164694d became candidate at term 89
Jan 16 12:23:58 LEGION etcd[1087]: 8e9e05c52164694d received MsgVoteResp from 8e9e05c52164694d at term 89
Jan 16 12:23:58 LEGION etcd[1087]: 8e9e05c52164694d became leader at term 89
Jan 16 12:23:58 LEGION etcd[1087]: raft.node: 8e9e05c52164694d elected leader 8e9e05c52164694d at term 89
Jan 16 12:23:58 LEGION etcd[1087]: published {Name:LEGION ClientURLs:[http://localhost:2379]} to cluster cdf818194e3a8c32
Jan 16 12:23:58 LEGION etcd[1087]: ready to serve client requests
Jan 16 12:23:58 LEGION systemd[1]: Started etcd - highly-available key value store.
Jan 16 12:23:58 LEGION etcd[1087]: serving insecure client requests on 127.0.0.1:2379, this is strongly discouraged!
$ 
```

### Step 2: Clone and run kvmservice in control plane
Clone kvmservice code and checkout to non-k8s branch.

```
$ git clone https://github.com/kubearmor/kvm-service.git
Cloning into 'kvm-service'...
remote: Enumerating objects: 1252, done.
remote: Counting objects: 100% (215/215), done.
remote: Compressing objects: 100% (111/111), done.
remote: Total 1252 (delta 122), reused 132 (delta 102), pack-reused 1037
Receiving objects: 100% (1252/1252), 139.62 MiB | 1.70 MiB/s, done.
Resolving deltas: 100% (702/702), done.
$ cd kvm-service/
$ git checkout non-k8s
Branch 'non-k8s' set up to track remote branch 'non-k8s' from 'origin'.
Switched to a new branch 'non-k8s'
$ 
```

Now navigate to kvm-service/src/service/ and do a make in that dir to compile kvm-service code. 
Once compilation is successful, run kvm-service using below command.
```
$ sudo ./kvmservice --non-k8s 2> /dev/null 
2022-01-16 13:06:16.304185      INFO    BUILD-INFO: commit:901ea26, branch: non-k8s, date: 2022-01-16T07:35:51Z, version: 
2022-01-16 13:06:16.304278      INFO    Initializing all the KVMS daemon attributes
2022-01-16 13:06:16.304325      INFO    Establishing connection with etcd service => http://localhost:2379
2022-01-16 13:06:16.333682      INFO    Initialized the ETCD client!
2022-01-16 13:06:16.333748      INFO    Initiliazing the KVMServer => podip:192.168.0.14 clusterIP:192.168.0.14 clusterPort:32770
2022-01-16 13:06:16.333771      INFO    KVMService attributes got initialized
2022-01-16 13:06:17.333915      INFO    Starting HTTP Server
2022-01-16 13:06:17.334005      INFO    Starting Cilium Node Registration Observer
2022-01-16 13:06:17.334040      INFO    Triggered the keepalive ETCD client
2022-01-16 13:06:17.334077      INFO    Starting gRPC server
2022-01-16 13:06:17.334149      INFO    ETCD: Getting raw values key:cilium/state/noderegister/v1

2022-01-16 13:06:17.335092      INFO    Successfully KVMServer Listening on port 32770
```
Now since kvmservice is up and running, next step is to run karmor(kubearmor-client) utility to support all other steps in kvmservice.

### Step 3: Install karmor
Run the below command to install karmor utility
```
curl -sfL https://raw.githubusercontent.com/kubearmor/kubearmor-client/main/install.sh | sudo sh -s -- -b /usr/local/bin
```

### Step 4: Onboard a VM using karmor command
Few example yamls are provided under kvm-service/example for VM onboarding. The same can be used for reference.
```
$ cat kvmpolicy1.yaml 
apiVersion: security.kubearmor.com/v1
kind: KubeArmorVirtualMachine
metadata:
  name: testvm1
  labels:
    name: vm1
    vm: true
$ 
$ ./karmor vm add kvmpolicy1.yaml 
Success
$ 
```
In the above yaml, the VM is given a name **testvm1** and is configured with 2 labels **name:vm1** and **vm:true**. 

When a new VM is onboarded, the kvmservice assigns a new identity to it.
To see the list of onboarded VMs, run the below command.
```
$ ./karmor vm list
List of configured vms are : 
[ VM : testvm1, Identity : 1090 ]
$
```

### Step 5: Download installation script for the configured VM
Download the VM installation script for the configured VM
```
$ ./karmor vm --non-k8s getscript -v testvm1
VM installation script copied to testvm1.sh
$ 
```

### Step 6: Run the installation script in a VM
Copy the downloaded installation script on to a VM and execute the installation script to run kubearmor.

The script runs/starts the kubearmor docker image and connects to the kvmservice in non-k8s control plane.

The Kubearmor in VM now waiting for policies from kvmservice.

### Step 7: Apply VM policies
Few example yamls are provided under kvm-service/example for kubearmor policy enforcement in VM. The same can be used for reference.

```
$ cat khp-example-vmname.yaml 
apiVersion: security.kubearmor.com/v1
kind: KubeArmorHostPolicy
metadata:
  name: khp-02
spec:
  nodeSelector:
    matchLabels:
      name: vm1
  severity: 5
  file:
    matchPaths:
    - path: /proc/cpuinfo
  action:
    Block
$ ./karmor vm --non-k8s policy add khp-example-vmname.yaml 
Success
$
``` 
### Step 8:  To verify enforced policy in VM
To verify the enforced policy in VM, run karmor in VM and watch on alerts.
```
$ karmor log
gRPC server: localhost:32767
Created a gRPC client (localhost:32767)
Checked the liveness of the gRPC server
Started to watch alerts
```

Now try accessing /proc/cpuinfo. The result shows ***permission denied*** and also see the block alerts in VM with karmor log.

```
$ cat /proc/cpuinfo 
cat: /proc/cpuinfo: Permission denied
$ 
$ karmor log
gRPC server: localhost:32767
Created a gRPC client (localhost:32767)
Checked the liveness of the gRPC server
Started to watch alerts

== Alert / 2022-01-16 08:24:33.153921 ==
Cluster Name: default
Host Name: 4511a8accc65
Policy Name: khp-02
Severity: 5
Type: MatchedHostPolicy
Source: cat
Operation: File
Resource: /proc/cpuinfo
Data: syscall=SYS_OPENAT fd=-100 flags=O_RDONLY
Action: Block
Result: Permission denied
```

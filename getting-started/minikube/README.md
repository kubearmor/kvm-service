# Setting up KVM Operator and KVM Service in minikube cluster
Follow the below steps to deploy kvmsoperator and kvmservice in minikube cluster

## Install minikube
To install minikube in the host machine run the installation script using the below command

```
$ cd KVMService/getting-started/minikube/
$ ./install_minikube.sh
```

## Start minikube
Once minikube installation is complete, start minikube using below command

```
$ minikube start
😄  minikube v1.23.2 on Ubuntu 20.04
✨  Automatically selected the docker driver. Other choices: virtualbox, none, ssh
👍  Starting control plane node minikube in cluster minikube
🚜  Pulling base image ...
🔥  Creating docker container (CPUs=2, Memory=3900MB) ...
🐳  Preparing Kubernetes v1.22.2 on Docker 20.10.8 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
$ 
```

The above confirms on the minikube up and running. 
Use `minikube status` to check the status of minikube cluster.

## Exposing ports to host machine
To expose minikube ports to host machine, open a new tab and run `minikube tunnel`.
Minikube will start exposing all cluster ports through minikube IP and continuous print will be displayed as shown below.

```
$ minikube tunnel
Status:
        machine: minikube
        pid: 206833
        route: 10.96.0.0/12 -> 192.168.49.2
        minikube: Running
        services: [kvmservice]
    errors: 
                minikube: no errors
                router: no errors
                loadbalancer emulator: no errors
```

## Apply VM and HostPolicy CRDs
After starting minikube, apply the VM/ExternalWorkload and Hostpolicy CRD using below commands.

```
$ minikube kubectl -- apply -f ../../deployments/CRD/KubeArmorExternalWorkloadPolicy.yaml 
customresourcedefinition.apiextensions.k8s.io/kubearmorexternalworkloads.security.kubearmor.com created
$ 
$ minikube kubectl -- apply -f ../../deployments/CRD/KubeArmorHostPolicy.yaml 
customresourcedefinition.apiextensions.k8s.io/kubearmorhostpolicies.security.kubearmor.com created
$ 
```

Once all CRDs are applied, the next step is to deploy kvmsoperator and kvmservice.

## Deploy etcd in minikube
Deploy etcd in minikube. ETCD is used for common data storage across pods.
```
$ minikube kubectl -- apply -f ../../deployments/etcd.yml
```

## Deploy kvmsoperator and kvmservice in minikube
Once all modifications are complete, deploy kvmsoperator and kvmservice using below commands.
```
$ minikube kubectl -- apply -f ../../src/operator/kvmsoperator.yaml
serviceaccount/kvmsoperator created
clusterrolebinding.rbac.authorization.k8s.io/kvmsoperator created
service/kvmsoperator created
deployment.apps/kvmsoperator created
$ 
$ minikube kubectl -- apply -f ../../src/service/kvmservice.yaml 
serviceaccount/kvmservice created
clusterrolebinding.rbac.authorization.k8s.io/kvmservice created
service/kvmservice created
deployment.apps/kvmservice created
$ 
```

To check the pods/services running, use the below command.
```
$ minikube kubectl -- get pods,svc -A
NAMESPACE     NAME                                   READY   STATUS    RESTARTS        AGE
kube-system   pod/coredns-78fcd69978-49sbp           1/1     Running   0               5m14s
kube-system   pod/etcd-minikube                      1/1     Running   0               5m26s
kube-system   pod/etcd0                              1/1     Running   0               4m38s
kube-system   pod/kube-apiserver-minikube            1/1     Running   0               5m26s
kube-system   pod/kube-controller-manager-minikube   1/1     Running   0               5m26s
kube-system   pod/kube-proxy-6cw8b                   1/1     Running   0               5m14s
kube-system   pod/kube-scheduler-minikube            1/1     Running   0               5m26s
kube-system   pod/kvmservice-78679f4d4b-qq8dq        1/1     Running   0               3m11s
kube-system   pod/kvmsoperator-d7696c8d7-s7mnf       1/1     Running   0               4m2s
kube-system   pod/storage-provisioner                1/1     Running   1 (4m43s ago)   5m25s

NAMESPACE     NAME                   TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                  AGE
default       service/kubernetes     ClusterIP      10.96.0.1       <none>         443/TCP                  5m28s
kube-system   service/etcd-client    ClusterIP      10.98.233.98    <none>         2379/TCP                 4m38s
kube-system   service/etcd0          ClusterIP      10.97.148.192   <none>         2379/TCP,2380/TCP        4m38s
kube-system   service/kube-dns       ClusterIP      10.96.0.10      <none>         53/UDP,53/TCP,9153/TCP   5m27s
kube-system   service/kvmservice     LoadBalancer   10.101.41.83    10.101.41.83   32770:30201/TCP          3m11s
kube-system   service/kvmsoperator   ClusterIP      10.104.57.14    <none>         40400/TCP                4m2s
$ 
```

As we could see in above output, the opertor and kvmpods are up and running.

## Configure new vm/workload
To configure a new VM/workload, apply a yaml with new vm CRD.
Some example yamls can be found under (https://github.com/kubearmor/KVMService/tree/main/examples)

Run below command to configure a new workload in kvmsoperator.
```
$ minikube kubectl -- apply -f ../../examples/kewpolicy.yaml 
kubearmorexternalworkloadpolicy.security.kubearmor.com/external-workload-01 created
$ 
```
To confirm on the configuration of new workload, refer kvmsoperator logs. 
```
$ minikube kubectl -- logs svc/kvmsoperator --namespace kube-system
2021-12-02 10:42:20.337325      INFO    Establishing connection with etcd service => http://10.97.148.192:2379
2021-12-02 10:42:20.338653      INFO    Initialized the ETCD client!
2021-12-02 10:42:20.338694      INFO    Successfully initialized KVMSOperator
2021-12-02 10:42:21.339010      INFO    Started the external workload CRD watcher
$ 
$ minikube kubectl -- logs svc/kvmservice --namespace kube-system | grep -v level
2021-12-02 10:43:03.866077      INFO    BUILD-INFO: commit:, branch: , date: , version: 
2021-12-02 10:43:03.876092      INFO    kvmservice external IP => 10.101.41.83
2021-12-02 10:43:03.876125      INFO    Initializing all the KVMS daemon attributes
2021-12-02 10:43:03.878345      INFO    Establishing connection with etcd service => http://10.97.148.192:2379
2021-12-02 10:43:03.880047      INFO    Initialized the ETCD client!
2021-12-02 10:43:03.880091      INFO    Initiliazing the KVMServer => podip:192.168.49.2 clusterIP:10.101.41.83 clusterPort:32770
2021-12-02 10:43:03.880101      INFO    KVMService attributes got initialized

2021-12-02 10:43:04.881037      INFO    K8S Client is successfully initialize
2021-12-02 10:43:04.881118      INFO    Watcher triggered for the host policies
2021-12-02 10:43:04.881179      INFO    Triggered the keepalive ETCD client
2021-12-02 10:43:04.881210      INFO    Starting gRPC server
2021-12-02 10:43:04.881770      INFO    Successfully KVMServer Listening on port 32770
$ 

```

## Download installation script to host machine using karmor
Install karmor with instruction from the link below.
https://github.com/kubearmor/kubearmor-client/blob/main/README.md

With the configured name, download the installation script to host machine using below karmor command.
```
$ ./karmor vm -v external-workload-01
VM installation script copied to external-workload-01.sh
$ 
```

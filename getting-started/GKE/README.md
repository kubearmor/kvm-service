# Steps for launching kvmsoperator in any GKE cluster
Follow the steps mentioned below to install/deploy kvmsoperator on any GKE cluster.
Connect to any GKE cluster of your choice. 

Once connected, follow the instructions below.

## Addng secrets to the cluster
Since both ETCD and kvmservice uses a server-side TLS for secure communication,
it is required to add the tls certificates as secrets to the cluster.
Follow the below steps to add the secrets to the cluster. 

```
$ cd certs/
$ ./gen-script.sh 
Certificates generated successfully 
Generating RSA private key, 4096 bit long modulus (2 primes)
..............................................................................................................................................................................................................................................................................................................................++++
........++++
e is 65537 (0x010001)
Generating RSA private key, 4096 bit long modulus (2 primes)
...........................................................................................................++++
............++++
e is 65537 (0x010001)
Signature ok
subject=C = IN, ST = TN, O = kubearmor, CN = kubearmor.com
Getting CA Private Key
Attaching certificates to k8s secrets 
secret/server-certs created
secret/ca-cert created
Removing local copy of certs 
$ 
```
Once all certificates are added to the cluster, continue to the next step.

## Deploying our own etcd. 
GKE has it own etcd and also manages the same in its control pane. 
The pods in control pane are not visible outside. 

For kvmsoperator to be deployed, we require etcd to be up and running  and also the etcd service's IP visible. 
Since we do not have access to that information, we will be deploying our own etcd. 

Before following the below instructions, navigate to getting-started/GKE folder.

### Deploy ETCD
Deploy etcd using below commands.

```
$ kubectl apply -f ../../deployments/etcd.yml 
service/etcd-client created
pod/etcd0 created
service/etcd0 created
$ 
```

## Apply VM and HostPolicy CRDs
Apply the VirtualMachine and Hostpolicy CRD using below commands.

```
$ kubectl apply -f ../../deployments/CRD/KubeVirtualMachine.yaml 
customresourcedefinition.apiextensions.k8s.io/kubearmorvirtualmachines.security.kubearmor.com created
$ 
$ kubectl apply -f ../../deployments/CRD/KubeArmorHostPolicy.yaml 
customresourcedefinition.apiextensions.k8s.io/kubearmorhostpolicies.security.kubearmor.com created
$ 
```
Once all CRDs are applied, the next step is to deploy kvmsoperator and kvmservice.

## Deploy kvmsoperator and kvmservice in minikube
Once all modifications are complete, deploy kvmsoperator and kvmservice using below commands.
```
$ kubectl apply -f ../../src/operator/kvmsoperator.yaml
serviceaccount/kvmsoperator created
clusterrolebinding.rbac.authorization.k8s.io/kvmsoperator created
service/kvmsoperator created
deployment.apps/kvmsoperator created
$ 
$ kubectl apply -f ../../src/service/kvmservice.yaml 
serviceaccount/kvmservice created
clusterrolebinding.rbac.authorization.k8s.io/kvmservice created
service/kvmservice created
deployment.apps/kvmservice created
$ 
```

To check the pods/services running, use the below command.
```
$ kubectl get pods,svc -A
NAMESPACE     NAME                                                                 READY   STATUS              RESTARTS   AGE
kube-system   pod/etcd0                                                            1/1     Running             0          28h
kube-system   pod/konnectivity-agent-5c57995cbd-8sfpw                              1/1     Running             0          2d1h
kube-system   pod/konnectivity-agent-5c57995cbd-9l2wb                              1/1     Running             0          2d
kube-system   pod/konnectivity-agent-autoscaler-5c49cb58bb-87rjh                   1/1     Running             0          2d1h
kube-system   pod/kube-dns-599484b884-t9dwr                                        0/3     ContainerCreating   0          2d
kube-system   pod/kube-dns-599484b884-vhwfc                                        3/3     Running             0          2d
kube-system   pod/kube-dns-autoscaler-844c9d9448-629b2                             1/1     Running             0          2d1h
kube-system   pod/kube-flannel-ds-amd64-9qfzv                                      1/1     Running             0          2d
kube-system   pod/kube-flannel-ds-amd64-hclds                                      1/1     Running             0          2d
kube-system   pod/kube-proxy-gke-backend-core-tri-backend-core-tri-e0d16b4d-c3ao   1/1     Running             0          2d
kube-system   pod/kube-proxy-gke-backend-core-tri-backend-core-tri-e0d16b4d-jxka   1/1     Running             0          15d
kube-system   pod/kubearmor-f9jsw                                                  1/1     Running             0          15d
kube-system   pod/kubearmor-host-policy-manager-69cfc96948-57khf                   2/2     Running             0          2d1h
kube-system   pod/kubearmor-n6g7c                                                  1/1     Running             0          15d
kube-system   pod/kubearmor-policy-manager-986bd8dbc-jqxnx                         2/2     Running             0          2d1h
kube-system   pod/kubearmor-relay-645667c695-xsg6l                                 1/1     Running             0          2d1h
kube-system   pod/kvmservice-5cfd97d69f-4wsjt                                      1/1     Running             0          28h
kube-system   pod/kvmsoperator-d7696c8d7-4vk8q                                     1/1     Running             0          28h
kube-system   pod/l7-default-backend-865b4c8f8b-pw4x8                              1/1     Running             0          2d1h
kube-system   pod/metrics-server-v0.4.4-857776bc9c-rvqk7                           2/2     Running             0          2d1h
kube-system   pod/pdcsi-node-bzjf7                                                 2/2     Running             0          15d
kube-system   pod/pdcsi-node-hpnnn                                                 2/2     Running             0          15d

NAMESPACE     NAME                                                    TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)             AGE
default       service/kubernetes                                      ClusterIP      10.92.0.1      <none>           443/TCP             37d
kube-system   service/default-http-backend                            NodePort       10.92.0.215    <none>           80:31820/TCP        37d
kube-system   service/etcd-client                                     ClusterIP      10.92.15.197   <none>           2379/TCP            3d7h
kube-system   service/etcd0                                           ClusterIP      10.92.11.30    <none>           2379/TCP,2380/TCP   3d7h
kube-system   service/kube-dns                                        ClusterIP      10.92.0.10     <none>           53/UDP,53/TCP       37d
kube-system   service/kubearmor                                       ClusterIP      10.92.7.176    <none>           32767/TCP           31d
kube-system   service/kubearmor-host-policy-manager-metrics-service   ClusterIP      10.92.0.251    <none>           8443/TCP            31d
kube-system   service/kubearmor-policy-manager-metrics-service        ClusterIP      10.92.1.107    <none>           8443/TCP            31d
kube-system   service/kvmservice                                      LoadBalancer   10.92.5.137    34.133.108.207   32770:31666/TCP     28h
kube-system   service/kvmsoperator                                    ClusterIP      10.92.12.133   <none>           40400/TCP           28h
kube-system   service/metrics-server                                  ClusterIP      10.92.6.70     <none>           443/TCP             37d
$ 
```
As we could see in above output, the opertor and kvmpods are up and running.

## Configure new vm/workload
To configure a new VM, apply a yaml with new vm CRD.
Some example yamls can be found under (https://github.com/kubearmor/KVMService/tree/main/examples)

Run below command to configure a new workload in kvmsoperator.
```

$ kubectl apply -f ../../examples/kvmpolicy.yaml 
kubearmorvirtualmachine.security.kubearmor.com/kvm1 created
$ 
```
To confirm on the configuration of new workload, refer kvmsoperator logs. 
```
$ kubectl logs svc/kvmsoperator -n kube-system
2021-12-02 10:42:20.337325      INFO    Establishing connection with etcd service => https://10.92.11.30:2379
2021-12-02 10:42:20.338653      INFO    Initialized the ETCD client!
2021-12-02 10:42:20.338694      INFO    Successfully initialized KVMSOperator
2021-12-02 10:42:21.339010      INFO    Started the external workload CRD watcher
$ 
$ kubectl logs svc/kvmservice -n kube-system | grep -v level
2021-12-02 10:43:03.866077      INFO    BUILD-INFO: commit:, branch: , date: , version: 
2021-12-02 10:43:03.876092      INFO    kvmservice external IP => 10.101.41.83
2021-12-02 10:43:03.876125      INFO    Initializing all the KVMS daemon attributes
2021-12-02 10:43:03.878345      INFO    Establishing connection with etcd service => https://10.92.11.30:2379
2021-12-02 10:43:03.880047      INFO    Initialized the ETCD client!
2021-12-02 10:43:03.880091      INFO    Initiliazing the KVMServer => podip:192.168.49.2 clusterIP:34.133.108.207 clusterPort:32770
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
$ ./karmor vm -v kvm1
VM installation script copied to kvm1.sh
$ 
```
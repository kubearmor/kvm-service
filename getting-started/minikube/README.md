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

## Apply VM and HostPolicy CRDs
After starting minikube, apply the Virtual Machine and Hostpolicy CRD using below commands.

```
$ minikube kubectl -- apply -f ../../deployments/CRD/KubeArmorVirtualMachinePolicy.yaml 
customresourcedefinition.apiextensions.k8s.io/kubearmorvirtualmachines.security.kubearmor.com created
$ 
$ minikube kubectl -- apply -f ../../deployments/CRD/KubeArmorHostPolicy.yaml 
customresourcedefinition.apiextensions.k8s.io/kubearmorhostpolicies.security.kubearmor.com created
$ 
```

Once all CRDs are applied, the next step is to deploy kvmsoperator and kvmservice.

## Modify yamls
Before deploying the kvmsoperator and kvmservice, the minikube IP needs to be updated on the yaml.
Run below command to get minikube IP.
```
$ minikube ip
192.168.49.2
$ 
```
Update this ip in extrnalIPs and ipAddress fields of the yaml.

## Deploy kvmsoperator and kvmservice in minikube
Once all modifications are complete, deploy kvmsoperator and kvmservice using below commands.
```
$ minikube kubectl -- apply -f kvmsoperator.yaml 
serviceaccount/kvmsoperator created
clusterrolebinding.rbac.authorization.k8s.io/kvmsoperator created
service/kvmsoperator created
deployment.apps/kvmsoperator created
$ 
$ minikube kubectl -- apply -f kvmservice.yaml 
serviceaccount/kvmservice created
clusterrolebinding.rbac.authorization.k8s.io/kvmservice created
service/kvmservice created
deployment.apps/kvmservice created
$ 
```

To check the pods/services running, use the below command.
```
$ minikube kubectl -- get pods,svc -A 
NAMESPACE     NAME                                   READY   STATUS    RESTARTS      AGE
kube-system   pod/coredns-78fcd69978-4mtbv           1/1     Running   0             29m
kube-system   pod/etcd-minikube                      1/1     Running   0             30m
kube-system   pod/kube-apiserver-minikube            1/1     Running   0             30m
kube-system   pod/kube-controller-manager-minikube   1/1     Running   0             30m
kube-system   pod/kube-proxy-cttzc                   1/1     Running   0             29m
kube-system   pod/kube-scheduler-minikube            1/1     Running   0             30m
kube-system   pod/kvmservice-7887c65686-l9wwk        1/1     Running   0             80s
kube-system   pod/kvmsoperator-7cf87cc795-jkfm2      1/1     Running   0             85s
kube-system   pod/storage-provisioner                1/1     Running   1 (29m ago)   30m

NAMESPACE     NAME                   TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                  AGE
default       service/kubernetes     ClusterIP      10.96.0.1       <none>         443/TCP                  30m
kube-system   service/kube-dns       ClusterIP      10.96.0.10      <none>         53/UDP,53/TCP,9153/TCP   30m
kube-system   service/kvmservice     LoadBalancer   10.108.40.148   192.168.49.2   4040:31916/TCP           80s
kube-system   service/kvmsoperator   ClusterIP      10.96.125.248   <none>         32770/TCP                85s
$ 
```

As we could see in above output, the opertor and kvmpods are up and running.

## Configure new vm
To configure a new VM, apply a yaml with new vm CRD.
Some example yamls can be found under (https://github.com/kubearmor/KVMService/tree/main/examples)

Run below command to configure a new vm in kvmsoperator.
```
$ minikube kubectl -- apply -f ../../examples/kvmpolicy.yaml 
kubearmorvirtualmachinepolicy.security.kubearmor.com/kvm1 created
$ 
```
To confirm on the configuration of new vm, refer kvmsoperator logs. 
```
$ minikube kubectl -- logs kvmsoperator-7cf87cc795-jkfm2 --namespace kube-system
2021-10-26 11:25:56.019147      INFO    Initialized the ETCD client!
2021-10-26 11:25:56.019210      INFO    Initiliazing the CLIHandler => Port:32770
2021-10-26 11:25:56.019222      INFO    Successfully initialized the KVMSOperator with args => (clusterIp:'192.168.49.2' clusterPort:40400
2021-10-26 11:25:57.020349      INFO    Started the Virtual Machine CRD watcher
2021-10-26 11:25:57.020439      INFO    Started the CLI Handler
2021-10-26 11:25:57.020862      INFO    Successfully CLIHandler Listening on port 32770
2021-10-26 11:32:36.041556      INFO    Received Virtual Machine policy request!!!
2021-10-26 11:32:36.041615      INFO    New Virtual Machine CRD is configured! => kvm1
2021-10-26 11:32:36.041642      INFO    Mappings identity to ewName=> map[65168:kvm1]
2021-10-26 11:32:36.041654      INFO    ETCD: putting key:/kvm-opr-map-identity-to-ewname/65168 value:kvm1
2021-10-26 11:32:36.042456      INFO    Mappings ewName to identity => map[kvm1:65168]
2021-10-26 11:32:36.042493      INFO    ETCD: putting key:/kvm-opr-map-ewname-to-identity/kvm1 value:65168
2021-10-26 11:32:36.042938      INFO    Generated the identity(kvm1) for this CRD:65168
2021-10-26 11:32:36.042958      INFO    Updating identity to label map identity:65168 label:abc=xyz
2021-10-26 11:32:36.042966      INFO    ETCD: putting key:/kvm-opr-identity-to-label-maps/65168 value:abc=xyz
2021-10-26 11:32:36.043496      INFO    ETCD: putting key:/kvm-opr-label-to-identities-map/abc=xyz value:[65168]
$ 
```

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

## Download installation script to host machine using karmor
With the configured name, download the installation script to host machine using below karmor command.
```
$ ./karmor vm -n kvm1

VM installation script copied to kvm1.sh
$ 
```

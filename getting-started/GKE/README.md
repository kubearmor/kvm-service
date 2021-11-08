# Steps for launching kvmsoperator in any GKE cluster
Follow the steps mentioned below to install/deploy kvmsoperator on any GKE cluster.
Connect to any GKE cluster of your choice. 

Once connected, follow the instructions below. 

## Deploying our own etcd. 
GKE has it own etcd and also manages the same in its control pane. 
The pods in control pane are not visible outside. 

For kvmsoperator to be deployed, we require etcd to be up and running  and also the etcd service's IP visible. 
Since we do not have access to that information, we will be deploying our own etcd. 

### Deploy ETCD
Deploy etcd using below commands.

```
$ kubectl apply -f etcd.yml 
service/etcd-client created
pod/etcd0 created
service/etcd0 created
$ 
```
Once etcd is deployed, get the clusterIP of the deployed etcd service using below command.

```
$ kubectl get services -A
NAMESPACE     NAME                                            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
default       kubernetes                                      ClusterIP   10.92.0.1      <none>        443/TCP             5d22h
kube-system   default-http-backend                            NodePort    10.92.0.215    <none>        80:31820/TCP        5d22h
kube-system   etcd-client                                     ClusterIP   10.92.11.42    <none>        2379/TCP            71s
kube-system   etcd0                                           ClusterIP   10.92.14.153   <none>        2379/TCP,2380/TCP   69s
kube-system   kube-dns                                        ClusterIP   10.92.0.10     <none>        53/UDP,53/TCP       5d22h
kube-system   kubearmor                                       ClusterIP   10.92.7.176    <none>        32767/TCP           50m
kube-system   kubearmor-host-policy-manager-metrics-service   ClusterIP   10.92.0.251    <none>        8443/TCP            50m
kube-system   kubearmor-policy-manager-metrics-service        ClusterIP   10.92.1.107    <none>        8443/TCP            50m
kube-system   kvmsoperator-59df4c8897-m7sx2-jqrsm             ClusterIP   10.92.5.66     <none>        32770/TCP           2d23h
kube-system   metrics-server                                  ClusterIP   10.92.6.70     <none>        443/TCP             5d22h
$ 

```
Make a note of the etcd0 cluster IP. Here the etcd cluster IP is 10.92.14.153
This IP is required to be configured in kvmsoperator.

## Deploy kvmsoperator service
Before deploying kvmsoperator, 2 things to be noted down.
1. The IP of the connected GKE cluster.
2. The cluster IP of the running etcd0 service.

### kvmsoperator operator yaml modifications
1. In line 66, replace the IP with the GKE cluster IP.
2. In line 69, replace the IP with the etcd0 service's cluster IP.

Once the above modifications are complete, deploy kvmsoperator with below commands.

```
$ kubectl apply -f kvmsoperator.yaml 
serviceaccount/kvmsoperator created
clusterrolebinding.rbac.authorization.k8s.io/kvmsoperator created
service/kvmsoperator created
deployment.apps/kvmsoperator created
$ 
```

kvmsoperator is deployed. To check the logs of kvmsoperator, run below command

```
$ kubectl logs -n kube-system kvmsoperator-68c4bccdd4-6mzlw
2021-11-01 06:17:47.509570      INFO    ENTRY!
2021-11-01 06:17:47.510076      INFO    Getting lease
2021-11-01 06:17:47.518585      INFO    Initialized the ETCD client!
2021-11-01 06:17:47.518659      INFO    Initiliazing the CLIHandler => Port:32770
2021-11-01 06:17:47.518738      INFO    Successfully initialized the KVMSOperator with args => (clusterIp:'35.238.22.7' clusterPort:40400
2021-11-01 06:17:48.521887      INFO    Started the external workload CRD watcher
2021-11-01 06:17:48.521964      INFO    Started the CLI Handler
2021-11-01 06:17:48.523054      INFO    Successfully CLIHandler Listening on port 32770
$ 
```

kvmsoperator is successfuly deployed and running.
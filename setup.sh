#!/bin/bash


cluster_clean() {
    kubectl delete -f /home/vagrant/KVMService/KVMService/service/kvmservice.yaml

    kubectl delete -f /home/vagrant/KVMService/KVMService/operator/kvmsoperator.yaml

    CRDS=`kubectl get kvm | awk '(NR>1)' |awk '{print $1}'`
    echo "MDEBUG: $CRDS"

    kubectl delete kvm $CRDS

    KHPS=`kubectl get khp | awk '(NR>1)' |awk '{print $1}'`
    echo "MDEBUG:$KHPS"

    kubectl delete khp $KHPS

    sudo rm /mnt/gen-script/*

    kubectl get kvm
    kubectl get khp
}

cluster_up() {
    kubectl apply -f /home/vagrant/KVMService/KVMService/service/kvmservice.yaml

    kubectl apply -f /home/vagrant/KVMService/KVMService/operator/kvmsoperator.yaml
}

if [ $# -eq 1 ]; then
    if [ $1 == "clean" ]; then
        echo "Cleaning the cluster"
        cluster_clean
    else
        echo "err: invalid commad"
    fi
else
    echo "Bringing up the cluster"
    cluster_up
fi

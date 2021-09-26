#!/bin/bash

DOCKER_OPTS=" -d -p 32769:32769 --log-driver syslog --restart always"
DOCKER_OPTS+=" --privileged --add-host kvms.kvmoperator.io:1.1.1.1"
DOCKER_OPTS+=" --volume /var/run/docker.sock:/var/run/docker.sock"
DOCKER_OPTS+=" --volume /usr/src:/usr/src"
DOCKER_OPTS+=" --volume /lib/modules:/lib/modules"
DOCKER_OPTS+=" --volume /sys/kernel/debug:/sys/kernel/debug"
DOCKER_OPTS+=" --volume /etc/kubernetes/pki/etcd/:/etc/kubernetes/pki/etcd/"
DOCKER_OPTS+=" --volume /etc/os-release:/media/root/etc/os-release"

kvmoperator_OPTS="-port=32769 -ipAddress=1.1.1.1"

if [ -n "$(sudo docker ps -a -q -f name=kvmoperator)" ]; then
    echo "Shutting down running kvmoperator agent"
    sudo docker rm -f kvmoperator || true
fi

kvmoperator_IMAGE="quay.io/maddy007_maha/kvmoperator:v3"

echo "Launching kvmoperator agent..."
sudo docker run --name kvmoperator $DOCKER_OPTS $kvmoperator_IMAGE kvmoperator $kvmoperator_OPTS
#sudo docker run --name kvmoperator -d -p 32769:32769 --log-driver syslog --restart always  --volume /var/run/docker.sock:/var/run/docker.sock --volume /etc/kubernetes/pki/etcd/:/etc/kubernetes/pki/etcd/ --privileged $kvmoperator_IMAGE kvmoperator $kvmoperator_OPTS


